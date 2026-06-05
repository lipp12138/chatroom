// Package chatroom provides word-pair generation for "Who Is Undercover".
package chatroom

import (
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Pair is one civilian/undercover word pair.
type Pair struct {
	Civilian   string `json:"civilian"`
	Undercover string `json:"undercover"`
}

var (
	errNoPairs = errors.New("chatroom: no word pairs")

	defaultPicker = mustNewPicker(defaultPairs)
	activePicker  = defaultPicker
	activeMu      sync.RWMutex

	roomTTL       = 2 * time.Hour
	recentLimit   = 5
	defaultRoom   = "_default"
	roomHistories = make(map[string]*roomHistory)
	roomMu        sync.Mutex
)

type picker struct {
	pairs []wordPair
	rng   *rand.Rand
	mu    sync.Mutex
}

type wordPair struct {
	pair Pair
	key  string
}

type roomHistory struct {
	recent    []string
	recentSet map[string]struct{}
	lastUsed  time.Time
}

// Update loads a local word library into memory.
//
// Supported formats:
//   - .txt: one pair per line, for example: 苹果,梨
//   - .json: [{"civilian":"苹果","undercover":"梨"}]
//
// If loading fails, Update switches back to the built-in default library and
// returns the error. Pick never reads files; it only uses the in-memory library.
func Update(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		setActive(defaultPicker)
		return nil
	}

	p, err := loadFile(path)
	if err != nil {
		setActive(defaultPicker)
		return err
	}

	setActive(p)
	return nil
}

// Pick returns a random word pair from the in-memory word library.
//
// Pass a room id to keep recent words independent between rooms. The same room
// will not repeat a word pair within its latest 5 successful picks when the
// active library has enough pairs.
func Pick(roomID ...string) (Pair, error) {
	activeMu.RLock()
	p := activePicker
	activeMu.RUnlock()

	now := time.Now()
	room := normalizeRoom(roomID...)

	roomMu.Lock()
	defer roomMu.Unlock()

	cleanupRoomsLocked(now)
	history := roomHistories[room]
	if history == nil {
		history = newRoomHistory()
		roomHistories[room] = history
	}
	history.lastUsed = now

	pair, key, err := p.pickAvoid(history.recentSet)
	if err != nil {
		return Pair{}, err
	}

	history.remember(key)
	return pair, nil
}

func newRoomHistory() *roomHistory {
	return &roomHistory{
		recentSet: make(map[string]struct{}, recentLimit),
	}
}

func setActive(p *picker) {
	activeMu.Lock()
	activePicker = p
	activeMu.Unlock()

	roomMu.Lock()
	roomHistories = make(map[string]*roomHistory)
	roomMu.Unlock()
}

func loadFile(path string) (*picker, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(filepath.Ext(path), ".json") {
		return loadJSON(data)
	}
	return loadText(data)
}

func loadJSON(data []byte) (*picker, error) {
	var pairs []Pair
	if err := json.Unmarshal(data, &pairs); err != nil {
		return nil, err
	}
	return newPicker(pairs)
}

func loadText(data []byte) (*picker, error) {
	lines := strings.Split(string(data), "\n")
	pairs := make([]Pair, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		pair, ok := parseTextLine(line)
		if ok {
			pairs = append(pairs, pair)
		}
	}

	return newPicker(pairs)
}

func parseTextLine(line string) (Pair, bool) {
	for _, sep := range []string{",", "，", "|", "\t"} {
		left, right, ok := strings.Cut(line, sep)
		if !ok {
			continue
		}

		pair := Pair{
			Civilian:   strings.TrimSpace(left),
			Undercover: strings.TrimSpace(right),
		}
		return pair, pair.Civilian != "" && pair.Undercover != ""
	}

	return Pair{}, false
}

func newPicker(pairs []Pair) (*picker, error) {
	cleaned := make([]Pair, 0, len(pairs))
	seen := make(map[string]struct{}, len(pairs))

	for _, pair := range pairs {
		pair.Civilian = strings.TrimSpace(pair.Civilian)
		pair.Undercover = strings.TrimSpace(pair.Undercover)
		if pair.Civilian == "" || pair.Undercover == "" || pair.Civilian == pair.Undercover {
			continue
		}
		key := pair.Civilian + "\x00" + pair.Undercover
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		cleaned = append(cleaned, pair)
	}

	if len(cleaned) == 0 {
		return nil, errNoPairs
	}

	words := make([]wordPair, 0, len(cleaned))
	for _, pair := range cleaned {
		words = append(words, wordPair{
			pair: pair,
			key:  pairKey(pair),
		})
	}

	return &picker{
		pairs: words,
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func (p *picker) pick() (Pair, error) {
	pair, _, err := p.pickAvoid(nil)
	return pair, err
}

func (p *picker) pickAvoid(avoid map[string]struct{}) (Pair, string, error) {
	if p == nil || len(p.pairs) == 0 {
		return Pair{}, "", errNoPairs
	}

	p.mu.Lock()
	usableAvoid := avoid
	if len(usableAvoid) >= len(p.pairs) {
		usableAvoid = nil
	}

	var picked wordPair
	for {
		picked = p.pairs[p.rng.Intn(len(p.pairs))]
		if _, found := usableAvoid[picked.key]; !found {
			break
		}
	}
	p.mu.Unlock()

	return picked.pair, picked.key, nil
}

func normalizeRoom(roomID ...string) string {
	if len(roomID) == 0 {
		return defaultRoom
	}

	room := strings.TrimSpace(roomID[0])
	if room == "" {
		return defaultRoom
	}
	return room
}

func cleanupRoomsLocked(now time.Time) {
	for room, history := range roomHistories {
		if now.Sub(history.lastUsed) > roomTTL {
			delete(roomHistories, room)
		}
	}
}

func (h *roomHistory) remember(key string) {
	h.recent = append(h.recent, key)
	h.recentSet[key] = struct{}{}
	if len(h.recent) > recentLimit {
		old := h.recent[0]
		copy(h.recent, h.recent[1:])
		h.recent = h.recent[:recentLimit]
		if !containsString(h.recent, old) {
			delete(h.recentSet, old)
		}
	}
}

func pairKey(pair Pair) string {
	return pair.Civilian + "\x00" + pair.Undercover
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func mustNewPicker(pairs []Pair) *picker {
	picker, err := newPicker(pairs)
	if err != nil {
		panic(err)
	}
	return picker
}

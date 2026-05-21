// Package chatroom provides word-pair generation for "Who Is Undercover".
package chatroom

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Pair is one civilian/undercover word pair.
type Pair struct {
	Civilian   string `json:"civilian"`
	Undercover string `json:"undercover"`
}

// Role is a player's assignment in one round.
type Role struct {
	Player       int    `json:"player"`
	Word         string `json:"word"`
	IsUndercover bool   `json:"is_undercover"`
}

// RoundResult is a complete generated game round.
type RoundResult struct {
	Pair  Pair   `json:"pair"`
	Roles []Role `json:"roles"`
}

// Config controls which word library is loaded.
type Config struct {
	// File is a local JSON word library path.
	//
	// If File is empty or does not exist, Load falls back to the built-in
	// library. If File exists but cannot be read or parsed, Load returns the
	// error so bad custom libraries are not silently ignored.
	File string `json:"file"`
}

var (
	ErrNoPairs       = errors.New("chatroom: no word pairs")
	ErrInvalidPlayer = errors.New("chatroom: invalid player count")

	defaultPicker = mustNew(defaultPairs)
)

// Picker stores a loaded word library.
//
// Create it once with New, LoadFile, or LoadURL, then reuse it.
type Picker struct {
	pairs []Pair
	rng   *rand.Rand
	mu    sync.Mutex
}

// Pick returns a random word pair.
func Pick() Pair {
	pair, _ := defaultPicker.Pick()
	return pair
}

// Round creates shuffled player assignments for one game.
func Round(playerCount, undercoverCount int) (RoundResult, error) {
	return defaultPicker.Round(playerCount, undercoverCount)
}

// Default returns a picker backed by the built-in word library.
func Default() *Picker {
	return mustNew(defaultPairs)
}

// All returns a copy of the built-in word library.
func All() []Pair {
	out := make([]Pair, len(defaultPairs))
	copy(out, defaultPairs)
	return out
}

// Load creates a picker from config.
//
// If config.File is empty or the file does not exist, Load uses the built-in
// word library. Use this for app startup configuration.
func Load(config Config) (*Picker, error) {
	path := strings.TrimSpace(config.File)
	if path == "" {
		return Default(), nil
	}

	picker, err := LoadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return nil, err
	}
	return picker, nil
}

// New creates a picker from custom pairs.
func New(pairs []Pair) (*Picker, error) {
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
		return nil, ErrNoPairs
	}

	return &Picker{
		pairs: cleaned,
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// LoadJSON loads pairs from JSON.
//
// Expected format:
//
//	[
//	  {"civilian":"苹果","undercover":"梨"},
//	  {"civilian":"可乐","undercover":"雪碧"}
//	]
func LoadJSON(r io.Reader) (*Picker, error) {
	var pairs []Pair
	if err := json.NewDecoder(r).Decode(&pairs); err != nil {
		return nil, err
	}
	return New(pairs)
}

// LoadFile loads pairs from a local JSON file.
func LoadFile(path string) (*Picker, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return LoadJSON(file)
}

// LoadURL loads pairs from a remote JSON API.
func LoadURL(ctx context.Context, url string) (*Picker, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("chatroom: load url failed: %s", resp.Status)
	}

	return LoadJSON(resp.Body)
}

// Pick returns a random word pair from this picker.
func (p *Picker) Pick() (Pair, error) {
	if p == nil || len(p.pairs) == 0 {
		return Pair{}, ErrNoPairs
	}

	p.mu.Lock()
	pair := p.pairs[p.rng.Intn(len(p.pairs))]
	p.mu.Unlock()

	return pair, nil
}

// Round creates shuffled player assignments from this picker's word library.
func (p *Picker) Round(playerCount, undercoverCount int) (RoundResult, error) {
	if playerCount < 3 || undercoverCount < 1 || undercoverCount >= playerCount {
		return RoundResult{}, ErrInvalidPlayer
	}

	pair, err := p.Pick()
	if err != nil {
		return RoundResult{}, err
	}

	roles := make([]Role, playerCount)
	for i := range roles {
		roles[i] = Role{
			Player: i + 1,
			Word:   pair.Civilian,
		}
	}
	for i := 0; i < undercoverCount; i++ {
		roles[i].Word = pair.Undercover
		roles[i].IsUndercover = true
	}

	p.mu.Lock()
	p.rng.Shuffle(len(roles), func(i, j int) {
		roles[i], roles[j] = roles[j], roles[i]
	})
	p.mu.Unlock()

	for i := range roles {
		roles[i].Player = i + 1
	}

	return RoundResult{Pair: pair, Roles: roles}, nil
}

func mustNew(pairs []Pair) *Picker {
	picker, err := New(pairs)
	if err != nil {
		panic(err)
	}
	return picker
}

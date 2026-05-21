// Package chatroom provides word-pair generation for "Who Is Undercover".
package chatroom

import (
	"encoding/json"
	"errors"
	"math/rand"
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

var (
	errNoPairs = errors.New("chatroom: no word pairs")

	defaultPicker = mustNewPicker(defaultPairs)
	fileCache     = make(map[string]cachedPicker)
	fileCacheMu   sync.Mutex
)

type picker struct {
	pairs []Pair
	rng   *rand.Rand
	mu    sync.Mutex
}

type cachedPicker struct {
	modTime time.Time
	size    int64
	picker  *picker
}

// Pick returns a random word pair.
//
// If file is empty, Pick uses the built-in word library. If file is provided
// and exists, Pick uses that local JSON word library. If file is provided but
// does not exist, Pick falls back to the built-in word library.
func Pick(file ...string) (Pair, error) {
	p, err := pickerFor(file...)
	if err != nil {
		return Pair{}, err
	}
	return p.pick()
}

func pickerFor(file ...string) (*picker, error) {
	if len(file) == 0 || strings.TrimSpace(file[0]) == "" {
		return defaultPicker, nil
	}

	path := strings.TrimSpace(file[0])
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaultPicker, nil
		}
		return nil, err
	}
	if info.IsDir() {
		return defaultPicker, nil
	}

	fileCacheMu.Lock()
	cached, ok := fileCache[path]
	if ok && cached.modTime.Equal(info.ModTime()) && cached.size == info.Size() {
		fileCacheMu.Unlock()
		return cached.picker, nil
	}
	fileCacheMu.Unlock()

	p, err := loadFile(path)
	if err != nil {
		return nil, err
	}

	fileCacheMu.Lock()
	fileCache[path] = cachedPicker{
		modTime: info.ModTime(),
		size:    info.Size(),
		picker:  p,
	}
	fileCacheMu.Unlock()

	return p, nil
}

func loadFile(path string) (*picker, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pairs []Pair
	if err := json.Unmarshal(data, &pairs); err != nil {
		return nil, err
	}
	return newPicker(pairs)
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

	return &picker{
		pairs: cleaned,
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func (p *picker) pick() (Pair, error) {
	if p == nil || len(p.pairs) == 0 {
		return Pair{}, errNoPairs
	}

	p.mu.Lock()
	pair := p.pairs[p.rng.Intn(len(p.pairs))]
	p.mu.Unlock()

	return pair, nil
}

func mustNewPicker(pairs []Pair) *picker {
	picker, err := newPicker(pairs)
	if err != nil {
		panic(err)
	}
	return picker
}

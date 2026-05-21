// Package keywords provides fast word-pair generation for "Who Is Undercover".
package keywords

import (
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"
)

// Difficulty describes how easy it is to distinguish the two words.
type Difficulty string

const (
	Easy   Difficulty = "easy"
	Normal Difficulty = "normal"
	Hard   Difficulty = "hard"
)

// Pair is one valid civilian/undercover word pair.
type Pair struct {
	Civilian    string     `json:"civilian"`
	Undercover  string     `json:"undercover"`
	Category    string     `json:"category"`
	Difficulty  Difficulty `json:"difficulty"`
	Description string     `json:"description,omitempty"`
}

// Role is a player's assignment in one round.
type Role struct {
	Player       int    `json:"player"`
	Word         string `json:"word"`
	IsUndercover bool   `json:"is_undercover"`
}

// Round is a complete generated game round.
type Round struct {
	Pair  Pair   `json:"pair"`
	Roles []Role `json:"roles"`
}

// Option filters Pick and Round generation.
type Option func(*pickOptions)

type pickOptions struct {
	category   string
	difficulty Difficulty
}

// WithCategory filters by category, for example "food" or "daily".
func WithCategory(category string) Option {
	return func(o *pickOptions) {
		o.category = strings.TrimSpace(category)
	}
}

// WithDifficulty filters by difficulty.
func WithDifficulty(difficulty Difficulty) Option {
	return func(o *pickOptions) {
		o.difficulty = difficulty
	}
}

// Bank stores a validated word bank and lookup indexes.
//
// Create it once and reuse it. Pick and Round are safe for concurrent use.
type Bank struct {
	pairs        []Pair
	all          []int
	byCategory   map[string][]int
	byDifficulty map[Difficulty][]int
	byFilter     map[string][]int
	rng          *rand.Rand
	mu           sync.Mutex
}

var (
	ErrNoPairs       = errors.New("keywords: no word pairs matched")
	ErrInvalidPair   = errors.New("keywords: invalid word pair")
	ErrInvalidPlayer = errors.New("keywords: invalid player count")
)

// Default returns a bank backed by the built-in Chinese word library.
func Default() *Bank {
	b, err := New(DefaultPairs())
	if err != nil {
		panic(err)
	}
	return b
}

// New builds an indexed bank from pairs.
func New(pairs []Pair) (*Bank, error) {
	if len(pairs) == 0 {
		return nil, ErrNoPairs
	}

	b := &Bank{
		pairs:        make([]Pair, 0, len(pairs)),
		all:          make([]int, 0, len(pairs)),
		byCategory:   make(map[string][]int),
		byDifficulty: make(map[Difficulty][]int),
		byFilter:     make(map[string][]int),
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	seen := make(map[string]struct{}, len(pairs))
	for _, pair := range pairs {
		pair = normalize(pair)
		if err := validate(pair); err != nil {
			return nil, err
		}

		key := pair.Civilian + "\x00" + pair.Undercover
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		idx := len(b.pairs)
		b.pairs = append(b.pairs, pair)
		b.all = append(b.all, idx)
		b.byCategory[pair.Category] = append(b.byCategory[pair.Category], idx)
		b.byDifficulty[pair.Difficulty] = append(b.byDifficulty[pair.Difficulty], idx)
		b.byFilter[filterKey(pair.Category, pair.Difficulty)] = append(b.byFilter[filterKey(pair.Category, pair.Difficulty)], idx)
	}

	if len(b.pairs) == 0 {
		return nil, ErrNoPairs
	}
	return b, nil
}

// LoadJSON loads pairs from a JSON array and returns an indexed bank.
func LoadJSON(r io.Reader) (*Bank, error) {
	var pairs []Pair
	if err := json.NewDecoder(r).Decode(&pairs); err != nil {
		return nil, err
	}
	return New(pairs)
}

// Len returns the pair count after validation and de-duplication.
func (b *Bank) Len() int {
	if b == nil {
		return 0
	}
	return len(b.pairs)
}

// Categories returns all available categories.
func (b *Bank) Categories() []string {
	if b == nil {
		return nil
	}
	out := make([]string, 0, len(b.byCategory))
	for category := range b.byCategory {
		out = append(out, category)
	}
	return out
}

// Pick returns a random pair matching all options.
func (b *Bank) Pick(opts ...Option) (Pair, error) {
	if b == nil || len(b.pairs) == 0 {
		return Pair{}, ErrNoPairs
	}

	options := pickOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	candidates := b.candidates(options)
	if len(candidates) == 0 {
		return Pair{}, ErrNoPairs
	}

	b.mu.Lock()
	idx := candidates[b.rng.Intn(len(candidates))]
	b.mu.Unlock()

	return b.pairs[idx], nil
}

// Round creates shuffled player assignments for one game.
func (b *Bank) Round(playerCount, undercoverCount int, opts ...Option) (Round, error) {
	if playerCount < 3 || undercoverCount < 1 || undercoverCount >= playerCount {
		return Round{}, ErrInvalidPlayer
	}

	pair, err := b.Pick(opts...)
	if err != nil {
		return Round{}, err
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

	b.mu.Lock()
	b.rng.Shuffle(len(roles), func(i, j int) {
		roles[i], roles[j] = roles[j], roles[i]
	})
	b.mu.Unlock()

	for i := range roles {
		roles[i].Player = i + 1
	}

	return Round{Pair: pair, Roles: roles}, nil
}

func (b *Bank) candidates(options pickOptions) []int {
	if options.category == "" && options.difficulty == "" {
		return b.all
	}

	if options.category != "" && options.difficulty == "" {
		return b.byCategory[options.category]
	}

	if options.category == "" && options.difficulty != "" {
		return b.byDifficulty[options.difficulty]
	}

	return b.byFilter[filterKey(options.category, options.difficulty)]
}

func normalize(pair Pair) Pair {
	pair.Civilian = strings.TrimSpace(pair.Civilian)
	pair.Undercover = strings.TrimSpace(pair.Undercover)
	pair.Category = strings.TrimSpace(pair.Category)
	if pair.Category == "" {
		pair.Category = "misc"
	}
	if pair.Difficulty == "" {
		pair.Difficulty = Normal
	}
	return pair
}

func filterKey(category string, difficulty Difficulty) string {
	return category + "\x00" + string(difficulty)
}

func validate(pair Pair) error {
	if pair.Civilian == "" || pair.Undercover == "" || pair.Civilian == pair.Undercover {
		return ErrInvalidPair
	}
	switch pair.Difficulty {
	case Easy, Normal, Hard:
		return nil
	default:
		return ErrInvalidPair
	}
}

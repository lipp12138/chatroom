// Package chatroom provides word-pair generation for "Who Is Undercover".
package chatroom

import (
	"errors"
	"math/rand"
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

var (
	ErrNoPairs       = errors.New("chatroom: no word pairs")
	ErrInvalidPlayer = errors.New("chatroom: invalid player count")

	defaultBank = newBank(defaultPairs)
)

type bank struct {
	pairs []Pair
	rng   *rand.Rand
	mu    sync.Mutex
}

// Pick returns a random word pair.
func Pick() Pair {
	pair, _ := defaultBank.pick()
	return pair
}

// Round creates shuffled player assignments for one game.
func Round(playerCount, undercoverCount int) (RoundResult, error) {
	return defaultBank.round(playerCount, undercoverCount)
}

// All returns a copy of the built-in word library.
func All() []Pair {
	out := make([]Pair, len(defaultPairs))
	copy(out, defaultPairs)
	return out
}

func newBank(pairs []Pair) *bank {
	cleaned := make([]Pair, 0, len(pairs))
	seen := make(map[string]struct{}, len(pairs))

	for _, pair := range pairs {
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

	return &bank{
		pairs: cleaned,
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (b *bank) pick() (Pair, error) {
	if b == nil || len(b.pairs) == 0 {
		return Pair{}, ErrNoPairs
	}

	b.mu.Lock()
	pair := b.pairs[b.rng.Intn(len(b.pairs))]
	b.mu.Unlock()

	return pair, nil
}

func (b *bank) round(playerCount, undercoverCount int) (RoundResult, error) {
	if playerCount < 3 || undercoverCount < 1 || undercoverCount >= playerCount {
		return RoundResult{}, ErrInvalidPlayer
	}

	pair, err := b.pick()
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

	b.mu.Lock()
	b.rng.Shuffle(len(roles), func(i, j int) {
		roles[i], roles[j] = roles[j], roles[i]
	})
	b.mu.Unlock()

	for i := range roles {
		roles[i].Player = i + 1
	}

	return RoundResult{Pair: pair, Roles: roles}, nil
}

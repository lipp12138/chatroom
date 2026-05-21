package chatroom

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestPick(t *testing.T) {
	pair := Pick()
	if pair.Civilian == "" || pair.Undercover == "" {
		t.Fatalf("empty pair: %+v", pair)
	}
	if pair.Civilian == pair.Undercover {
		t.Fatalf("same words: %+v", pair)
	}
}

func TestAll(t *testing.T) {
	pairs := All()
	if len(pairs) < 200 {
		t.Fatalf("default library too small: %d", len(pairs))
	}

	pairs[0].Civilian = ""
	if All()[0].Civilian == "" {
		t.Fatal("All must return a copy")
	}
}

func TestRound(t *testing.T) {
	round, err := Round(8, 2)
	if err != nil {
		t.Fatalf("round failed: %v", err)
	}
	if len(round.Roles) != 8 {
		t.Fatalf("unexpected role count: %d", len(round.Roles))
	}

	undercover := 0
	for i, role := range round.Roles {
		if role.Player != i+1 {
			t.Fatalf("unexpected player index: %+v", role)
		}
		if role.IsUndercover {
			undercover++
			if role.Word != round.Pair.Undercover {
				t.Fatalf("wrong undercover word: %+v", role)
			}
			continue
		}
		if role.Word != round.Pair.Civilian {
			t.Fatalf("wrong civilian word: %+v", role)
		}
	}
	if undercover != 2 {
		t.Fatalf("unexpected undercover count: %d", undercover)
	}
}

func TestRoundRejectsInvalidPlayerCount(t *testing.T) {
	_, err := Round(2, 1)
	if err != ErrInvalidPlayer {
		t.Fatalf("expected ErrInvalidPlayer, got %v", err)
	}
}

func TestLoadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "words.json")
	err := os.WriteFile(path, []byte(`[
		{"civilian":"苹果","undercover":"梨"},
		{"civilian":"可乐","undercover":"雪碧"}
	]`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	picker, err := LoadFile(path)
	if err != nil {
		t.Fatalf("load file failed: %v", err)
	}

	pair, err := picker.Pick()
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if pair.Civilian == "" || pair.Undercover == "" {
		t.Fatalf("empty pair: %+v", pair)
	}
}

func TestLoadURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"civilian":"火锅","undercover":"麻辣烫"}]`))
	}))
	defer server.Close()

	picker, err := LoadURL(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("load url failed: %v", err)
	}

	pair, err := picker.Pick()
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if pair.Civilian != "火锅" || pair.Undercover != "麻辣烫" {
		t.Fatalf("unexpected pair: %+v", pair)
	}
}

func TestNewRejectsEmptyLibrary(t *testing.T) {
	_, err := New([]Pair{{Civilian: "苹果", Undercover: "苹果"}})
	if err != ErrNoPairs {
		t.Fatalf("expected ErrNoPairs, got %v", err)
	}
}

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
		{"civilian":"apple","undercover":"pear"},
		{"civilian":"cola","undercover":"soda"}
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
		_, _ = w.Write([]byte(`[{"civilian":"hotpot","undercover":"malatang"}]`))
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
	if pair.Civilian != "hotpot" || pair.Undercover != "malatang" {
		t.Fatalf("unexpected pair: %+v", pair)
	}
}

func TestLoadConfigUsesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "words.json")
	err := os.WriteFile(path, []byte(`[{"civilian":"custom","undercover":"word"}]`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	picker, err := Load(Config{File: path})
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	pair, err := picker.Pick()
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if pair.Civilian != "custom" || pair.Undercover != "word" {
		t.Fatalf("unexpected pair: %+v", pair)
	}
}

func TestLoadConfigFallsBackToDefault(t *testing.T) {
	picker, err := Load(Config{File: filepath.Join(t.TempDir(), "missing.json")})
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	pair, err := picker.Pick()
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if pair.Civilian == "" || pair.Undercover == "" {
		t.Fatalf("empty pair: %+v", pair)
	}
}

func TestLoadConfigReturnsInvalidFileError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	err := os.WriteFile(path, []byte(`not json`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Load(Config{File: path})
	if err == nil {
		t.Fatal("expected invalid json error")
	}
}

func TestNewRejectsEmptyLibrary(t *testing.T) {
	_, err := New([]Pair{{Civilian: "same", Undercover: "same"}})
	if err != ErrNoPairs {
		t.Fatalf("expected ErrNoPairs, got %v", err)
	}
}

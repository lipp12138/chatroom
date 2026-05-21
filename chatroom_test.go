package chatroom

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPickUsesDefaultLibrary(t *testing.T) {
	resetDefault(t)

	pair, err := Pick()
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	assertValidPair(t, pair)
}

func TestUpdateUsesTextLibrary(t *testing.T) {
	resetDefault(t)

	path := filepath.Join(t.TempDir(), "words.txt")
	err := os.WriteFile(path, []byte("custom,word\napple|pear\n# comment\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if err := Update(path); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	pair, err := Pick()
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if pair.Civilian != "custom" && pair.Civilian != "apple" {
		t.Fatalf("unexpected text pair: %+v", pair)
	}
}

func TestUpdateUsesJSONLibrary(t *testing.T) {
	resetDefault(t)

	path := filepath.Join(t.TempDir(), "words.json")
	err := os.WriteFile(path, []byte(`[{"civilian":"custom","undercover":"word"}]`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if err := Update(path); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	pair, err := Pick()
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if pair.Civilian != "custom" || pair.Undercover != "word" {
		t.Fatalf("unexpected json pair: %+v", pair)
	}
}

func TestUpdateReloadsOnlyWhenCalled(t *testing.T) {
	resetDefault(t)

	path := filepath.Join(t.TempDir(), "words.txt")
	err := os.WriteFile(path, []byte("first,one\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	if err := Update(path); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	err = os.WriteFile(path, []byte("second,two\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	pair, err := Pick()
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if pair.Civilian != "first" || pair.Undercover != "one" {
		t.Fatalf("pick should keep in-memory library until Update is called: %+v", pair)
	}

	if err := Update(path); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	pair, err = Pick()
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if pair.Civilian != "second" || pair.Undercover != "two" {
		t.Fatalf("unexpected reloaded pair: %+v", pair)
	}
}

func TestUpdateFailureFallsBackToDefault(t *testing.T) {
	resetDefault(t)

	path := filepath.Join(t.TempDir(), "words.txt")
	err := os.WriteFile(path, []byte("custom,word\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	if err := Update(path); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	badPath := filepath.Join(t.TempDir(), "bad.txt")
	err = os.WriteFile(badPath, []byte("same,same\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if err := Update(badPath); err != errNoPairs {
		t.Fatalf("expected errNoPairs, got %v", err)
	}

	pair, err := Pick()
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	assertValidPair(t, pair)
	if pair.Civilian == "custom" || pair.Undercover == "word" {
		t.Fatalf("expected default library after update failure, got %+v", pair)
	}
}

func TestUpdateEmptyPathResetsDefault(t *testing.T) {
	resetDefault(t)

	path := filepath.Join(t.TempDir(), "words.txt")
	err := os.WriteFile(path, []byte("custom,word\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	if err := Update(path); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if err := Update(""); err != nil {
		t.Fatalf("reset failed: %v", err)
	}

	pair, err := Pick()
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	assertValidPair(t, pair)
}

func resetDefault(t *testing.T) {
	t.Helper()
	if err := Update(""); err != nil {
		t.Fatalf("reset default failed: %v", err)
	}
}

func assertValidPair(t *testing.T, pair Pair) {
	t.Helper()
	if pair.Civilian == "" || pair.Undercover == "" {
		t.Fatalf("empty pair: %+v", pair)
	}
	if pair.Civilian == pair.Undercover {
		t.Fatalf("same words: %+v", pair)
	}
}

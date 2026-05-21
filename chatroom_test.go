package chatroom

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPickUsesDefaultLibrary(t *testing.T) {
	pair, err := Pick()
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	assertValidPair(t, pair)
}

func TestPickUsesFileLibrary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "words.json")
	err := os.WriteFile(path, []byte(`[{"civilian":"custom","undercover":"word"}]`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	pair, err := Pick(path)
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if pair.Civilian != "custom" || pair.Undercover != "word" {
		t.Fatalf("unexpected pair: %+v", pair)
	}
}

func TestPickFallsBackWhenFileDoesNotExist(t *testing.T) {
	pair, err := Pick(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	assertValidPair(t, pair)
}

func TestPickReloadsChangedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "words.json")
	err := os.WriteFile(path, []byte(`[{"civilian":"first","undercover":"one"}]`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	pair, err := Pick(path)
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if pair.Civilian != "first" || pair.Undercover != "one" {
		t.Fatalf("unexpected first pair: %+v", pair)
	}

	err = os.WriteFile(path, []byte(`[{"civilian":"second","undercover":"two"}]`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	pair, err = Pick(path)
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if pair.Civilian != "second" || pair.Undercover != "two" {
		t.Fatalf("unexpected second pair: %+v", pair)
	}
}

func TestPickFallsBackAndReturnsInvalidFileError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	err := os.WriteFile(path, []byte(`not json`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	pair, err := Pick(path)
	if err == nil {
		t.Fatal("expected invalid json error")
	}
	assertValidPair(t, pair)
}

func TestPickFallsBackAndReturnsEmptyLibraryError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.json")
	err := os.WriteFile(path, []byte(`[{"civilian":"same","undercover":"same"}]`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	pair, err := Pick(path)
	if err != errNoPairs {
		t.Fatalf("expected errNoPairs, got %v", err)
	}
	assertValidPair(t, pair)
}

func TestPickFallsBackWhenChangedFileBecomesInvalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "words.json")
	err := os.WriteFile(path, []byte(`[{"civilian":"custom","undercover":"word"}]`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	pair, err := Pick(path)
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if pair.Civilian != "custom" || pair.Undercover != "word" {
		t.Fatalf("unexpected custom pair: %+v", pair)
	}

	err = os.WriteFile(path, []byte(`not json`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	pair, err = Pick(path)
	if err == nil {
		t.Fatal("expected invalid json error")
	}
	assertValidPair(t, pair)
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

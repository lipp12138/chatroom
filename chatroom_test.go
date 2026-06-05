package chatroom

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestBundledTextLibrary(t *testing.T) {
	resetDefault(t)

	if err := Update(filepath.Join("data", "ciku", "a.txt")); err != nil {
		t.Fatalf("update bundled text library failed: %v", err)
	}

	pair, err := Pick()
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	assertValidPair(t, pair)
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

func TestPickAvoidsRecentFivePairsInSameRoom(t *testing.T) {
	resetDefault(t)

	path := filepath.Join(t.TempDir(), "words.txt")
	err := os.WriteFile(path, []byte("a,1\nb,2\nc,3\nd,4\ne,5\nf,6\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	if err := Update(path); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	recent := make([]string, 0, recentLimit)
	for i := 0; i < 30; i++ {
		pair, err := Pick("room-a")
		if err != nil {
			t.Fatalf("pick failed: %v", err)
		}

		key := pairKey(pair)
		for _, old := range recent {
			if key == old {
				t.Fatalf("pair repeated within recent %d picks: %+v", recentLimit, pair)
			}
		}

		recent = append(recent, key)
		if len(recent) > recentLimit {
			recent = recent[1:]
		}
	}
}

func TestPickRoomHistoriesAreIndependent(t *testing.T) {
	resetDefault(t)

	path := filepath.Join(t.TempDir(), "words.txt")
	err := os.WriteFile(path, []byte("a,1\nb,2\nc,3\nd,4\ne,5\nf,6\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	if err := Update(path); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if _, err := Pick("room-a"); err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if _, err := Pick("room-b"); err != nil {
		t.Fatalf("pick failed: %v", err)
	}

	roomMu.Lock()
	defer roomMu.Unlock()

	if roomHistories["room-a"] == nil || roomHistories["room-b"] == nil {
		t.Fatalf("expected independent room histories, got %+v", roomHistories)
	}
}

func TestRoomHistoryExpires(t *testing.T) {
	resetDefault(t)

	if _, err := Pick("old-room"); err != nil {
		t.Fatalf("pick failed: %v", err)
	}

	roomMu.Lock()
	roomHistories["old-room"].lastUsed = time.Now().Add(-roomTTL - time.Minute)
	roomMu.Unlock()

	if _, err := Pick("new-room"); err != nil {
		t.Fatalf("pick failed: %v", err)
	}

	roomMu.Lock()
	defer roomMu.Unlock()

	if roomHistories["old-room"] != nil {
		t.Fatal("expected old room history to be cleaned up")
	}
	if roomHistories["new-room"] == nil {
		t.Fatal("expected new room history to exist")
	}
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

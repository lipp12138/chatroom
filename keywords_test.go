package keywords

import "testing"

func TestDefaultBank(t *testing.T) {
	b := Default()
	if b.Len() < 200 {
		t.Fatalf("default bank too small: %d", b.Len())
	}

	pair, err := b.Pick(WithCategory("food"), WithDifficulty(Easy))
	if err != nil {
		t.Fatalf("pick failed: %v", err)
	}
	if pair.Category != "food" || pair.Difficulty != Easy {
		t.Fatalf("unexpected pair: %+v", pair)
	}
}

func TestRound(t *testing.T) {
	b := Default()
	round, err := b.Round(8, 2)
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

func TestNewRejectsInvalidPair(t *testing.T) {
	_, err := New([]Pair{{Civilian: "苹果", Undercover: "苹果"}})
	if err != ErrInvalidPair {
		t.Fatalf("expected ErrInvalidPair, got %v", err)
	}
}

package chatroom

import "testing"

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

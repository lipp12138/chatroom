package chatroom

import "testing"

func BenchmarkPick(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = Pick()
	}
}

func BenchmarkRound(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := Round(8, 2); err != nil {
			b.Fatal(err)
		}
	}
}

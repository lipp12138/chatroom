package keywords

import "testing"

func BenchmarkPick(b *testing.B) {
	bank := Default()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := bank.Pick(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPickWithFilter(b *testing.B) {
	bank := Default()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := bank.Pick(WithCategory("food"), WithDifficulty(Easy)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRound(b *testing.B) {
	bank := Default()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := bank.Round(8, 2); err != nil {
			b.Fatal(err)
		}
	}
}

package chatroom

import "testing"

func BenchmarkPickDefault(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := Pick(); err != nil {
			b.Fatal(err)
		}
	}
}

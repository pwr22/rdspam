package main

import "testing"

func BenchmarkGenChunk(b *testing.B) {
	buf := make([]byte, bufLen)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		genDataChunk(buf)
	}
}

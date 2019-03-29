package main

import (
	"testing"

	"github.com/kbjorklu/xoshiro"
)

func BenchmarkGenChunk(b *testing.B) {
	buf := make([]byte, bufLen)
	rndSrc := xoshiro.NewXoshiro256StarStar(11)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		genDataChunk(buf, rndSrc)
	}
}

package main

import (
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"github.com/c2h5oh/datasize"
	"github.com/kbjorklu/xoshiro"
)

func BenchmarkXoshiro64Bits(b *testing.B) {
	rndSrc := xoshiro.NewXoshiro256StarStar(11)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rndSrc.Uint64()
	}
}
func BenchmarkWriteDataDiscard(b *testing.B) {
	streams := runtime.GOMAXPROCS(0)
	for i := 0; i < b.N; i++ {
		writeData(int(4*datasize.MB), 11, streams, ioutil.Discard)
	}
}

func BenchmarkWriteDataDevNull(b *testing.B) {
	streams := runtime.GOMAXPROCS(0)

	out, err := os.Create(os.DevNull)
	if err != nil {
		b.Fatal(err)
	}
	defer out.Close()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		writeData(int(4*datasize.MB), 11, streams, out)
	}
}

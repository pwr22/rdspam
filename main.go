package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/kbjorklu/xoshiro"
	flag "github.com/spf13/pflag"
)

const version = "v0.0.1"

var size = flag.IntP("size", "s", 0, "number of bytes to write or 0 (default) to keep going forever")
var printVersion = flag.BoolP("version", "V", false, "print version information")
var seed = flag.Int64P("seed", "S", 0, "seed to use for the data source (defaults to the current time)")

// TODO optimise for different output types?
const bufLen = 64 * datasize.KB // optimised for piping

// main will parse flags, do anything needed there then start writing out data as requested.
// If *size == 0 then it will go on forever.
// TODO Sequential for now, use goroutines for concurrency.
func main() {
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		return
	}

	if *size < 0 {
		fmt.Fprintf(os.Stderr, "size must be greater than 0 but you gave %v\n", *size)
		os.Exit(2)
	}

	if *seed == 0 { // if not set or user gave 0 - xoroshift doesn't like 0
		rand.Seed(time.Now().UnixNano())
		*seed = int64(rand.Uint64())
	}

	writeData()
}

// writeData writes the requested amount of random data to stdout then returns.
// If *size == 0 then it will keep generating forever and never return.
func writeData() {
	rndSrc := xoshiro.NewXoshiro256StarStar(*seed)
	buf := make([]byte, bufLen)
	bytesWritten := 0

	for {
		genDataChunk(buf, rndSrc)

		// handle the last write potentially being smaller and exit
		if *size > 0 && *size-bytesWritten <= len(buf) {
			os.Stdout.Write(buf[:*size-bytesWritten])
			break
		} else { // or do a full write an count the bytes
			os.Stdout.Write(buf)
			bytesWritten += len(buf)
		}
	}
}

// genDataChunk generates the next chunk of random data in c.
func genDataChunk(c []byte, r *xoshiro.Xoshiro256StarStar) {
	for i := 0; i < len(c); i += 8 {
		binary.LittleEndian.PutUint64(c[i:], r.Uint64())
	}
}

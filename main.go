package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/c2h5oh/datasize"
	"github.com/jes/lfsr64"
	flag "github.com/spf13/pflag"
)

const version = "v0.0.1"

var size = flag.IntP("size", "s", 0, "number of bytes to write or 0 to keep going forever")
var printVersion = flag.BoolP("version", "V", false, "print version information")

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

	writeData()
}

// writeData writes the requested amount of random data to stdout then returns.
// If *size == 0 then it will keep generating forever and never return.
func writeData() {
	buf := make([]byte, bufLen)
	bytesWritten := 0

	for {
		genDataChunk(buf)

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

// r is the source of randomness used for data generation.
// TODO Allow this to be seeded.
var r = lfsr64.NewLfsr64(99)

// genDataChunk generates the next chunk of random data in c.
func genDataChunk(c []byte) {
	for i := 0; i < len(c); i += 8 {
		binary.LittleEndian.PutUint64(c[i:], r.Uint64())
	}
}

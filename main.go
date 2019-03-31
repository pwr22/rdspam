package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/kbjorklu/xoshiro"
	flag "github.com/spf13/pflag"
)

const version = "v0.0.1"

var size = flag.StringP("size", "s", "0", "number of bytes to write or 0 (default) to keep going forever")
var printVersion = flag.BoolP("version", "V", false, "print version information")
var seed = flag.Int64P("seed", "S", 0, "seed to use for the data source (defaults to the current time)")

var bytesToWrite datasize.ByteSize

// TODO optimise for different output types?
const genBufLen = 4 * datasize.MB  // optimised minimising channel overheads
const writeSize = 64 * datasize.KB // optimised for piping
const buffers = 2                  // there doesn't seem to be any benefit of raising this as we're bottle necked on data generation anyway

// main will parse flags, do anything needed there then start the generator and writer
// If *size == 0 then it will go on forever.
func main() {
	// uncomment for cpu profiling
	// defer profile.Start().Stop()

	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		return
	}

	if err := bytesToWrite.UnmarshalText([]byte(*size)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if *seed == 0 { // if not set or user gave 0 - xoroshift doesn't like 0
		rand.Seed(time.Now().UnixNano())
		*seed = int64(rand.Uint64())
	}

	writeData(int(bytesToWrite), *seed, os.Stdout)
}

// writeData writes the requested amount of random data to out then returns.
// If *size == 0 then it will keep generating forever and never return.
func writeData(size int, seed int64, out io.Writer) {
	dataIn, dataOut := startGenerating(size, seed)

	writesDone := make(chan bool, 1)
	go writeBuffers(out, dataOut, dataIn, writesDone)

	<-writesDone
}

// startGenerating starts up a goroutine to generate data in buffers and returns channels to receive and return these
func startGenerating(size int, seed int64) (chan []byte, chan []byte) {
	bufIn, bufOut := make(chan []byte, buffers), make(chan []byte, buffers)

	for i := 0; i < buffers; i++ {
		bufIn <- make([]byte, genBufLen)
	}

	go genData(size, seed, bufIn, bufOut)

	return bufIn, bufOut
}

// genData reads in buffers, fills them with random data and sends them back out.
// It closes the bufOut channel to signal when it's done generating data
func genData(size int, seed int64, bufIn, bufOut chan []byte) {
	rndSrc := xoshiro.NewXoshiro256StarStar(seed)
	bytes := 0

	for buf := range bufIn {
		// fill the buffer
		for i := 0; i < len(buf); i += 8 {
			b := buf[i:] // if not doing this in a separate statement things are slower!
			binary.LittleEndian.PutUint64(b, rndSrc.Uint64())
		}
		bytes += len(buf)

		// handle the last buffer potentially needing to be smaller and finishing up
		if size > 0 && bytes >= size {
			s := size + int(genBufLen) - bytes // work out how much of this block we want
			buf = buf[:s]                      // just shrink the buffer / harmless if its good already
			bufOut <- buf
			break
		}

		bufOut <- buf
	}

	close(bufOut) // signal the writer that we're done
	// we can't close bufIn as it may still be putting back
}

// writeBuffers reads buffers, writes them to out and then returns the buffer.
// It signals on done when finished
func writeBuffers(out io.Writer, bufIn, bufOut chan []byte, done chan bool) {
	total := 0
	startTime := time.Now()

	for b := range bufIn {
		offset, writeSize := 0, int(writeSize)

		// write out all but the last blocks - these will all be full writeLen.
		// len(b)-writeSize is the last point where we could do a full writeLen (<0 for sizes < writeLen)
		for ; offset < len(b)-writeSize; offset += writeSize {
			n, err := out.Write(b[offset : offset+writeSize])
			total += n

			if err != nil {
				panic(err)
			}
		}

		// write the last chunk whatever size it is
		n, err := out.Write(b[offset:])
		total += n

		if err != nil {
			panic(err)
		}

		bufOut <- b
	}

	// emit statistics
	duration := time.Now().Sub(startTime)
	bytesPerS := datasize.ByteSize(float64(total) / duration.Seconds())
	fmt.Fprintf(os.Stderr, "wrote %s in %v at an average of %s/s\n", datasize.ByteSize(total).HR(), duration, bytesPerS.HR())

	done <- true
}

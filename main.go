package main

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/kbjorklu/xoshiro"
	flag "github.com/spf13/pflag"
)

const version = "v0.0.2"

var size = flag.StringP("size", "s", "0", "number of bytes to write or 0 (default) to keep going forever")
var printVersion = flag.BoolP("version", "V", false, "print version information")
var seed = flag.Int64P("seed", "S", 0, "seed to use for the data source except for go-crypto (defaults to the current time)")
var randSrc = flag.StringP("rand-source", "r", "xoshiro256**", fmt.Sprintf("source to use for random data, one of: %v", getAllowedRandSrcs()))

var bytesToWrite datasize.ByteSize

// TODO optimise for different output types?
const genBufLen = 4 * datasize.MB // optimised minimising channel overheads
const pipeWriteSize = 64 * datasize.KB
const fileWriteSize = genBufLen
const buffers = 2 // there doesn't seem to be any benefit of raising this as we're bottle necked on data generation anyway

var randSrcs = map[string]bool{"xoshiro256**": true, "go-math": true, "go-crypto": true}

func getAllowedRandSrcs() string {
	keys := make([]string, 0, len(randSrcs))
	for k := range randSrcs {
		keys = append(keys, k)
	}

	return strings.Join(keys, " ")
}

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

	if *seed == 0 { // if not set or user gave 0 - xoshiro256** doesn't like 0
		rand.Seed(time.Now().UnixNano())
		*seed = int64(rand.Uint64())
	} else if *randSrc == "go-crypto" {
		fmt.Fprintln(os.Stderr, "go-crypto cannot be seeded")
		os.Exit(2)
	}

	if _, ok := randSrcs[*randSrc]; !ok {
		fmt.Fprintf(os.Stderr, "%v is not a supported random data source - must be one of: %v\n", *randSrc, getAllowedRandSrcs())
		os.Exit(2)
	}

	writeData(int(bytesToWrite), *seed, os.Stdout)
}

// writeData writes the requested amount of random data to out then returns.
// If *size == 0 then it will keep generating forever and never return.
func writeData(size int, seed int64, out io.Writer) {
	dataIn, dataOut := startGenerating(size, seed)
	writesDone := startWriting(out, dataOut, dataIn)

	<-writesDone
}

// startGenerating starts up a goroutine to generate data in buffers and returns channels to receive and return these
func startGenerating(size int, seed int64) (chan []byte, chan []byte) {
	bufIn, bufOut := make(chan []byte, buffers), make(chan []byte, buffers)

	for i := 0; i < buffers; i++ {
		bufIn <- make([]byte, genBufLen)
	}

	// its much slower to call interface methods than concrete type ones and our main bottle neck is in the calls to Source64.Uint64.
	// so we make this concrete here to work around that - it means we'll have to have a generator per type probably.
	// an upside is we'll be able to support different generator interfaces without wrappers!
	if *randSrc == "xoshiro256**" {
		go genXoshiro256StarStar(size, seed, bufIn, bufOut)
	} else if *randSrc == "go-math" {
		go genMath(size, seed, bufIn, bufOut)
	} else if *randSrc == "go-crypto" {
		go genCrypto(size, seed, bufIn, bufOut)
	} else {
		panic(fmt.Errorf("Unsupported random data source %v", *randSrc))
	}

	return bufIn, bufOut
}

// genXoshiro256StarStar reads in buffers, fills them with random data and sends them back out.
// It closes the bufOut channel to signal when it's done generating data
func genXoshiro256StarStar(size int, seed int64, bufIn, bufOut chan []byte) {
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

// genMath reads in buffers, fills them with random data from math/rand and sends them back out.
// It closes the bufOut channel to signal when it's done generating data
func genMath(size int, seed int64, bufIn, bufOut chan []byte) {
	rand := rand.New(rand.NewSource(seed))
	bytes := 0

	lastBufferAfter := size - int(genBufLen)
	for buf := range bufIn {
		// handle the last buffer potentially needing to be smaller and finishing up
		if size > 0 && bytes >= lastBufferAfter {
			buf = buf[:size-bytes] // just shrink the buffer / harmless if its good already
			rand.Read(buf)

			bufOut <- buf
			break
		} else {
			rand.Read(buf)
			bytes += len(buf)

			bufOut <- buf
		}
	}

	close(bufOut) // signal the writer that we're done
	// we can't close bufIn as it may still be putting back
}

// genCrypto reads in buffers, fills them with random data from crypto/rand and sends them back out.
// It closes the bufOut channel to signal when it's done generating data
func genCrypto(size int, seed int64, bufIn, bufOut chan []byte) {
	bytes := 0

	lastBufferAfter := size - int(genBufLen)
	for buf := range bufIn {
		// handle the last buffer potentially needing to be smaller and finishing up
		if size > 0 && bytes >= lastBufferAfter {
			buf = buf[:size-bytes] // just shrink the buffer / harmless if its good already
			crand.Read(buf)

			bufOut <- buf
			break
		} else {
			crand.Read(buf)
			bytes += len(buf)

			bufOut <- buf
		}
	}

	close(bufOut) // signal the writer that we're done
	// we can't close bufIn as it may still be putting back
}

func startWriting(out io.Writer, dataIn, bufReturn chan []byte) chan bool {
	done := make(chan bool)
	go writeBuffers(out, dataIn, bufReturn, done)
	return done
}

// writeBuffers reads buffers, writes them to out and then returns the buffer.
// It signals on done when finished
func writeBuffers(out io.Writer, bufIn, bufOut chan []byte, done chan bool) {
	total := 0
	startTime := time.Now()
	writeSize := fileWriteSize

	// work out if we're writing to a pipe
	if f, ok := out.(*os.File); ok {
		fi, err := f.Stat()
		if err != nil {
			panic(err)
		}

		// going beyond the size of the pipe is slower
		if fi.Mode()&os.ModeNamedPipe != 0 {
			writeSize = pipeWriteSize
		}
	}

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

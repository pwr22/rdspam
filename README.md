# rdspam

[![Build Status](https://travis-ci.com/pwr22/rdspam.svg?branch=master)](https://travis-ci.com/pwr22/rdspam)
[![Build status](https://ci.appveyor.com/api/projects/status/cuptxx2040f6f9sa/branch/master?svg=true)](https://ci.appveyor.com/project/pwr22/rdspam/branch/master)
[![codecov](https://codecov.io/gh/pwr22/rdspam/branch/master/graph/badge.svg)](https://codecov.io/gh/pwr22/rdspam)
[![Go Report Card](https://goreportcard.com/badge/github.com/pwr22/rdspam)](https://goreportcard.com/report/github.com/pwr22/rdspam)
[![Downloads](https://img.shields.io/github/downloads/pwr22/rdspam/total.svg)](https://github.com/pwr22/rdspam/releases)
[![](https://tokei.rs/b1/github/pwr22/rdspam)](https://github.com/pwr22/rdspam)

Simple command line tool to spam random data on stdout

## Usage

    rdspam -s 1024 > /dev/null # write 1KiB random data to /dev/null

    -s, --size int   number of bytes to write or 0 to keep going forever (default 0)   
    -V, --version    print version information

## Installation

Head over to the [releases](https://github.com/pwr22/rdspam/releases) page, download the binary for your operating system and put it somewhere in your `$PATH`

## Why

I work in the storage business and as part of our testing we end up creating in house tools to write a lot of data to our solutions

rdspam is an open source solution to this problem, and probably a few others! lot

## Performance

rdspam aims to be as fast as possible so it uses the [xoshiro256**](http://xoshiro.di.unimi.it/) pseudorandom number generator (PRNG) as its source of data. It's [not cryptographically secure](http://www.pcg-random.org/posts/a-quick-look-at-xoshiro256.html) but its good enough for storage testing and doesn't deduplicate or compress easily
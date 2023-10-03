// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runes

import (
	"bufio"
	"errors"
	"fmt"
	"unicode/utf8"
)

// defaultBufSize is the default buffer size.
const defaultBufSize = 1024

var (
	errRuneDecode    = errors.New("UTF-8 rune decoding error")
	errBufferFull    = fmt.Errorf("buffer full")
	errNegativeCount = fmt.Errorf("negative count")
)

// RuneReader reads a byte stream of UTF-8 text as runes.
type RuneReader struct {
	// rd is the underlying reader that reads bytes..
	rd *bufio.Reader

	// buf is the rune lookahead buffer.
	buf []rune

	// r is the read index into the buffer.
	r int

	// e is the end index of the buffer.
	e int

	// err is the last read error that occurred.
	err error
}

// NewReader returns a new RuneReader with whose buffer has the
// default size.
func NewReader(r *bufio.Reader) *RuneReader {
	return NewReaderSize(r, defaultBufSize)
}

// NewReaderSize returns a new RuneReader with whose buffer has the
// specified size.
func NewReaderSize(r *bufio.Reader, size int) *RuneReader {
	return &RuneReader{
		rd:  r,
		buf: make([]rune, defaultBufSize),
	}
}

// Buffered returns the number of runes that can be read from the current
// buffer.
func (r *RuneReader) Buffered() int {
	return r.e - r.r
}

// Read reads runes into p and returns the number of runes read. If the number
// of runes read is different than the size of p, then an error is returned
// explaining the reason.
func (r *RuneReader) Read(p []rune) (int, error) {
	i := 0
	var err error
	for err == nil && i < len(p) {
		var rn rune
		rn, _, err = r.ReadRune()
		if err != nil {
			return i, err
		}

		p[i] = rn
		i++
	}
	return i, err
}

// ReadRune reads a single UTF-8 encoded Unicode character and returns the
// rune and its size in bytes. If the encoded rune is invalid, it consumes one
// byte and returns unicode.ReplacementChar (U+FFFD) with a size of 1. If an
// error occurs then it returns (utf8.RuneError, err).
func (r *RuneReader) ReadRune() (rune, int, error) {
	if r.Buffered() == 0 {
		r.fill()
	}

	if r.Buffered() == 0 {
		return 0, 0, r.readErr()
	}

	rn := r.buf[r.r]
	r.r++
	return rn, utf8.RuneLen(rn), nil
}

// Discard discards n runes and return the number discarded. If the number
// of runes discarded is different than n, then an error is returned explaining
// the reason. If 0 <= n <= r.Buffered(), Discard is guaranteed to succeed without
// reading from the underlying reader.
func (r *RuneReader) Discard(n int) (int, error) {
	if n < 0 {
		return 0, errNegativeCount
	}

	for i := 0; i < n; i++ {
		_, _, err := r.ReadRune()
		if err != nil {
			return i, err
		}
	}

	return n, nil
}

// Peek returns the next n bytes without advancing the reader. The bytes stop
// being valid at the next read call. If Peek returns fewer than n bytes, it also
// returns an error explaining why the read is short. An error is returned if
// n is larger than the reader's buffer size.
func (r *RuneReader) Peek(n int) ([]rune, error) {
	if n < 0 {
		return nil, errNegativeCount
	}

	if n > len(r.buf) {
		return nil, errBufferFull
	}

	if n > r.Buffered() {
		r.fill()
	}

	var err error
	if n > r.Buffered() {
		n = r.Buffered()
		err = r.readErr()
		if err == nil {
			err = errBufferFull
		}
	}

	return r.buf[r.r : r.r+n], err
}

// fill fills the RuneReader's buffer.
func (r *RuneReader) fill() {
	// Slide existing data to beginning.
	if r.r > 0 && r.Buffered() < len(r.buf) {
		copy(r.buf, r.buf[r.r:r.e])
		r.e -= r.r
		r.r = 0
	}

	// Fill the rest of the buffer.
	for ; r.e < len(r.buf); r.e++ {
		rn, _, err := r.rd.ReadRune()
		if err != nil {
			r.err = err
			break
		}
		r.buf[r.e] = rn
	}
}

// readErr returns the last readErr and clears it.
func (r *RuneReader) readErr() error {
	err := r.err
	r.err = nil
	return err
}

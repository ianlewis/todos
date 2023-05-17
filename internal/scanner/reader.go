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

package scanner

import (
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

// utfByteSize is the maximum number of bytes in a UTF-8 character.
const utfByteSize = 4

var (
	errRuneDecode = errors.New("UTF-8 rune decoding error")
	errPeekBuffer = fmt.Errorf("peek index greater than buffer size")
)

// RuneReader reads a byte stream of UTF-8 text as runes.
type RuneReader struct {
	r io.Reader

	// atEOF indicates that we have seen EOF on the reader.
	atEOF bool

	// buf is local buffer for decoding to runes.
	buf []byte

	// lookahead is the rune lookahead buffer.
	lookahead []rune

	// minLookaheadSize is the minimum desired lookahead buffer size.
	minLookaheadSize int

	// lookaheadSize is the lookahead buffer size.
	lookaheadSize int

	// bufSize is the byte reader buffer size.
	bufSize int
}

// NewRuneReader returns a new rune reader for an io.Reader.
func NewRuneReader(r io.Reader) *RuneReader {
	rdr := &RuneReader{
		r: r,
	}
	rdr.minLookaheadSize = 16
	rdr.lookaheadSize = rdr.minLookaheadSize * 16
	rdr.bufSize = rdr.lookaheadSize * utfByteSize
	return rdr
}

// Read reads runes into p and returns the number of runes read. If the number
// of runes read is different than the size of p, then an error is returned
// explaining the reason.
func (r *RuneReader) Read(p []rune) (int, error) {
	i := 0
	var err error
	for err == nil && i < len(p) {
		var rn rune
		rn, err = r.Next()
		if err != nil {
			return i, err
		}

		p[i] = rn
		i++
	}
	return i, err
}

// Next returns the next rune from in the buffer. If the next rune cannot be
// read then an error is returned.
func (r *RuneReader) Next() (rune, error) {
	err := r.readLookahead()
	if err != nil {
		return 0, err
	}
	if len(r.lookahead) == 0 {
		return 0, io.EOF
	}
	rn := r.lookahead[0]
	r.lookahead = r.lookahead[1:]
	return rn, nil
}

// Discard discards n runes and return the number discarded. If the number
// of runes discarded is different than the size of p, then an error is returned
// explaining the reason.
func (r *RuneReader) Discard(n int) (int, error) {
	for i := 0; i < n; i++ {
		_, err := r.Next()
		if err != nil {
			return i, err
		}
	}
	return n, nil
}

// Peek returns the next n bytes without advancing the reader. The bytes stop
// being valid at the next read call. If Peek returns fewer than n bytes, it also
// returns an error explaining why the read is short. An error is returned if
// n is larger than r's buffer size.
func (r *RuneReader) Peek(n int) ([]rune, error) {
	if n > r.lookaheadSize {
		return nil, errPeekBuffer
	}

	if err := r.readLookahead(); err != nil {
		return nil, err
	}

	var err error
	if n > len(r.lookahead) {
		n = len(r.lookahead)
		err = io.EOF
	}
	return r.lookahead[:n], err
}

func (r *RuneReader) readLookahead() error {
	lenBuf := len(r.lookahead)
	if lenBuf >= r.minLookaheadSize || r.atEOF {
		return nil
	}

	buf := make([]rune, lenBuf)
	copy(buf, r.lookahead)
	for i := lenBuf; i < r.lookaheadSize; i++ {
		rn, size, err := r.readRune()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}
		if size > 0 {
			//nolint:makezero // buf is prealocated with length for copy function.
			buf = append(buf, rn)
		}
	}
	r.lookahead = buf
	return nil
}

// readRune decodes a rune off off the bytes buffer and returns it.
func (r *RuneReader) readRune() (rune, int, error) {
	err := r.readToBuf()
	if err != nil {
		return 0, 0, err
	}

	if len(r.buf) == 0 {
		return 0, 0, io.EOF
	}

	rn, size := utf8.DecodeRune(r.buf)
	if rn == utf8.RuneError && size < 2 {
		return rn, size, errRuneDecode
	}
	r.buf = r.buf[size:]

	return rn, size, err
}

func (r *RuneReader) readToBuf() error {
	lenBuf := len(r.buf)
	if lenBuf >= utfByteSize || r.atEOF {
		return nil
	}
	buf := make([]byte, r.bufSize)
	copy(buf, r.buf)
	n, err := r.r.Read(buf[lenBuf:])
	if err != nil && err != io.EOF {
		return fmt.Errorf("reading to byte buffer: %w", err)
	}
	r.buf = buf[:n+lenBuf]
	r.atEOF = err == io.EOF
	return nil
}

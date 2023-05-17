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
	"bytes"
	"errors"
	"io"
	"testing"
)

type expectation interface {
	expect(*testing.T, *RuneReader)
}

type expectedPeek struct {
	size          int
	expectedRunes []rune
	expectedErr   error
}

func (e *expectedPeek) expect(t *testing.T, r *RuneReader) {
	t.Helper()
	p, err := r.Peek(e.size)
	if got, want := err, e.expectedErr; !errors.Is(got, want) {
		t.Errorf("expected error: got: %v, want: %v", got, want)
	}

	if got, want := p, e.expectedRunes; !compareRunes(got, want) {
		t.Errorf("Read: got: %q, want: %q", string(got), string(want))
	}
}

type expectedRead struct {
	size          int
	expectedNum   int
	expectedErr   error
	expectedRunes []rune
}

func (e *expectedRead) expect(t *testing.T, r *RuneReader) {
	t.Helper()
	p := make([]rune, e.size)
	n, err := r.Read(p)
	if got, want := err, e.expectedErr; !errors.Is(got, want) {
		t.Errorf("expected error: got: %v, want: %v", got, want)
	}

	if got, want := n, e.expectedNum; got != want {
		t.Errorf("Read: got: %v, want: %v", got, want)
	}
	if got, want := p[:n], e.expectedRunes; !compareRunes(got, want) {
		t.Errorf("Read: got: %q, want: %q", string(got), string(want))
	}
}

func TestRuneReader_ReadPeek(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		bytes    []byte
		expected []expectation
	}{
		{
			name:  "single read",
			bytes: []byte("Hello, 世界/World/Universe!"),
			expected: []expectation{
				&expectedRead{
					size:          9,
					expectedNum:   9,
					expectedRunes: []rune{'H', 'e', 'l', 'l', 'o', ',', ' ', '世', '界'},
				},
			},
		},
		{
			name:  "single read not exact",
			bytes: []byte("Hello, 世界"),
			expected: []expectation{
				&expectedRead{
					size:          12,
					expectedNum:   9,
					expectedRunes: []rune{'H', 'e', 'l', 'l', 'o', ',', ' ', '世', '界'},
					expectedErr:   io.EOF,
				},
			},
		},
		{
			name:  "multiple reads exact",
			bytes: []byte("Hello, 世界/World/Universe!"),
			expected: []expectation{
				&expectedRead{
					size:          3,
					expectedNum:   3,
					expectedRunes: []rune{'H', 'e', 'l'},
				},
				&expectedRead{
					size:          5,
					expectedNum:   5,
					expectedRunes: []rune{'l', 'o', ',', ' ', '世'},
				},
				&expectedRead{
					size:          1,
					expectedNum:   1,
					expectedRunes: []rune{'界'},
				},
			},
		},
		{
			name:  "multiple reads not exact",
			bytes: []byte("Hello, 世界"),
			expected: []expectation{
				&expectedRead{
					size:          3,
					expectedNum:   3,
					expectedRunes: []rune{'H', 'e', 'l'},
				},
				&expectedRead{
					size:          5,
					expectedNum:   5,
					expectedRunes: []rune{'l', 'o', ',', ' ', '世'},
				},
				&expectedRead{
					size:          8,
					expectedNum:   1,
					expectedRunes: []rune{'界'},
					expectedErr:   io.EOF,
				},
			},
		},
		{
			name:  "single peek",
			bytes: []byte("Hello, 世界/World/Universe!"),
			expected: []expectation{
				&expectedPeek{
					size:          3,
					expectedRunes: []rune{'H', 'e', 'l'},
				},
			},
		},
		{
			name:  "multiple peek",
			bytes: []byte("Hello, 世界/World/Universe!"),
			expected: []expectation{
				&expectedPeek{
					size:          3,
					expectedRunes: []rune{'H', 'e', 'l'},
				},
				&expectedPeek{
					size:          5,
					expectedRunes: []rune{'H', 'e', 'l', 'l', 'o'},
				},
			},
		},
		{
			name:  "read and peek",
			bytes: []byte("Hello, 世界/World/Universe!"),
			expected: []expectation{
				&expectedRead{
					size:          3,
					expectedNum:   3,
					expectedRunes: []rune{'H', 'e', 'l'},
				},
				&expectedPeek{
					size:          5,
					expectedRunes: []rune{'l', 'o', ',', ' ', '世'},
				},
			},
		},
		{
			name:  "read and peek multi",
			bytes: []byte("Hello, 世界/World/Universe!"),
			expected: []expectation{
				&expectedRead{
					size:          3,
					expectedNum:   3,
					expectedRunes: []rune{'H', 'e', 'l'},
				},
				&expectedPeek{
					size:          5,
					expectedRunes: []rune{'l', 'o', ',', ' ', '世'},
				},
				&expectedRead{
					size:          4,
					expectedNum:   4,
					expectedRunes: []rune{'l', 'o', ',', ' '},
				},
				&expectedPeek{
					size:          1,
					expectedRunes: []rune{'世'},
				},
			},
		},
		{
			name:  "peek exact",
			bytes: []byte("Hello, 世界/World/Universe!"),
			expected: []expectation{
				&expectedPeek{
					size:          9,
					expectedRunes: []rune{'H', 'e', 'l', 'l', 'o', ',', ' ', '世', '界'},
				},
			},
		},
		{
			name:  "peek not exact",
			bytes: []byte("Hello, 世界"),
			expected: []expectation{
				&expectedPeek{
					size:          11,
					expectedRunes: []rune{'H', 'e', 'l', 'l', 'o', ',', ' ', '世', '界'},
					expectedErr:   io.EOF,
				},
			},
		},
	}

	for i := range testCases {
		c := testCases[i]

		r := NewRuneReader(bytes.NewReader(c.bytes))
		r.lookaheadSize = 11
		r.bufSize = 24
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			for _, e := range c.expected {
				e.expect(t, r)
			}
		})
	}
}

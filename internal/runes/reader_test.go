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
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/ianlewis/todos/internal/utils"
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

	if got, want := p, e.expectedRunes; !utils.SliceEqual(got, want) {
		t.Errorf("Peek: got: %q, want: %q", string(got), string(want))
	}
}

type expectedDiscard struct {
	size        int
	expected    int
	expectedErr error
}

func (e *expectedDiscard) expect(t *testing.T, r *RuneReader) {
	t.Helper()
	n, err := r.Discard(e.size)
	if got, want := err, e.expectedErr; !errors.Is(got, want) {
		t.Errorf("expected error: got: %v, want: %v", got, want)
	}

	if got, want := n, e.expected; got != want {
		t.Errorf("Discard: got: %d, want: %d", got, want)
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
	if got, want := p[:n], e.expectedRunes; !utils.SliceEqual(got, want) {
		t.Errorf("Read: got: %q, want: %q", string(got), string(want))
	}
}

func TestRuneReader(t *testing.T) {
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
					expectedRunes: []rune("Hello, 世界"),
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
					expectedRunes: []rune("Hello, 世界"),
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
					expectedRunes: []rune("Hel"),
				},
				&expectedRead{
					size:          5,
					expectedNum:   5,
					expectedRunes: []rune("lo, 世"),
				},
				&expectedRead{
					size:          1,
					expectedNum:   1,
					expectedRunes: []rune("界"),
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
					expectedRunes: []rune("Hel"),
				},
				&expectedRead{
					size:          5,
					expectedNum:   5,
					expectedRunes: []rune("lo, 世"),
				},
				&expectedRead{
					size:          8,
					expectedNum:   1,
					expectedRunes: []rune("界"),
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
					expectedRunes: []rune("Hel"),
				},
			},
		},
		{
			name:  "multiple peek",
			bytes: []byte("Hello, 世界/World/Universe!"),
			expected: []expectation{
				&expectedPeek{
					size:          3,
					expectedRunes: []rune("Hel"),
				},
				&expectedPeek{
					size:          5,
					expectedRunes: []rune("Hello"),
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
					expectedRunes: []rune("Hel"),
				},
				&expectedPeek{
					size:          5,
					expectedRunes: []rune("lo, 世"),
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
					expectedRunes: []rune("Hel"),
				},
				&expectedPeek{
					size:          5,
					expectedRunes: []rune("lo, 世"),
				},
				&expectedRead{
					size:          4,
					expectedNum:   4,
					expectedRunes: []rune("lo, "),
				},
				&expectedPeek{
					size:          1,
					expectedRunes: []rune("世"),
				},
			},
		},
		{
			name:  "peek exact",
			bytes: []byte("Hello, 世界/World/Universe!"),
			expected: []expectation{
				&expectedPeek{
					size:          9,
					expectedRunes: []rune("Hello, 世界"),
				},
			},
		},
		{
			name:  "peek not exact",
			bytes: []byte("Hello, 世界"),
			expected: []expectation{
				&expectedPeek{
					size:          11,
					expectedRunes: []rune("Hello, 世界"),
					expectedErr:   io.EOF,
				},
			},
		},
		{
			name:  "peek neg count",
			bytes: []byte("Hello, 世界"),
			expected: []expectation{
				&expectedPeek{
					size:        -1,
					expectedErr: errNegativeCount,
				},
			},
		},
		{
			name:  "discard exact",
			bytes: []byte("Hello, 世界/World/Universe!"),
			expected: []expectation{
				&expectedDiscard{
					size:     9,
					expected: 9,
				},
				&expectedPeek{
					size:          6,
					expectedRunes: []rune("/World"),
				},
			},
		},
		{
			name:  "discard not exact",
			bytes: []byte("Hello, 世界"),
			expected: []expectation{
				&expectedDiscard{
					size:        11,
					expected:    9,
					expectedErr: io.EOF,
				},
				&expectedPeek{
					size:        5,
					expectedErr: io.EOF,
				},
			},
		},
		{
			name:  "discard neg count",
			bytes: []byte("Hello, 世界"),
			expected: []expectation{
				&expectedDiscard{
					size:        -1,
					expected:    0,
					expectedErr: errNegativeCount,
				},
			},
		},
	}

	for i := range testCases {
		c := testCases[i]

		r := NewReaderSize(bufio.NewReader(bytes.NewReader(c.bytes)), 11)
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			for _, e := range c.expected {
				e.expect(t, r)
			}
		})
	}
}

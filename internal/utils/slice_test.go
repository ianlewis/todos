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

package utils

import "testing"

func TestSliceEqual_int(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		l     []int
		r     []int
		equal bool
	}{
		"equal": {
			l:     []int{1, 2, 3, 4},
			r:     []int{1, 2, 3, 4},
			equal: true,
		},
		"equal empty": {
			l:     []int{},
			r:     []int{},
			equal: true,
		},
		"equal nil": {
			l:     nil,
			r:     nil,
			equal: true,
		},
		"not equal": {
			l:     []int{1, 2, 3, 4},
			r:     []int{1, 2, 3, 5},
			equal: false,
		},
		"not equal right shorter": {
			l:     []int{1, 2, 3, 4},
			r:     []int{1, 2, 3},
			equal: false,
		},
		"not equal left shorter": {
			l:     []int{1, 2, 3},
			r:     []int{1, 2, 3, 4},
			equal: false,
		},
		"not equal left empty": {
			l:     []int{},
			r:     []int{1, 2, 3, 4},
			equal: false,
		},
		"not equal right empty": {
			l:     []int{1, 2, 3, 4},
			r:     []int{},
			equal: false,
		},
		"not equal left nil": {
			l:     nil,
			r:     []int{1, 2, 3, 4},
			equal: false,
		},
		"not equal right nil": {
			l:     []int{1, 2, 3, 4},
			r:     nil,
			equal: false,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got, want := SliceEqual(tc.l, tc.r), tc.equal; got != want {
				t.Errorf("unexpected return, got: %v, want: %v", got, want)
			}
		})
	}
}

func TestSliceEqual_rune(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		l     []rune
		r     []rune
		equal bool
	}{
		"equal": {
			l:     []rune{1, 2, 3, 4},
			r:     []rune{1, 2, 3, 4},
			equal: true,
		},
		"equal empty": {
			l:     []rune{},
			r:     []rune{},
			equal: true,
		},
		"equal nil": {
			l:     nil,
			r:     nil,
			equal: true,
		},
		"not equal": {
			l:     []rune{1, 2, 3, 4},
			r:     []rune{1, 2, 3, 5},
			equal: false,
		},
		"not equal right shorter": {
			l:     []rune{1, 2, 3, 4},
			r:     []rune{1, 2, 3},
			equal: false,
		},
		"not equal left shorter": {
			l:     []rune{1, 2, 3},
			r:     []rune{1, 2, 3, 4},
			equal: false,
		},
		"not equal left empty": {
			l:     []rune{},
			r:     []rune{1, 2, 3, 4},
			equal: false,
		},
		"not equal right empty": {
			l:     []rune{1, 2, 3, 4},
			r:     []rune{},
			equal: false,
		},
		"not equal left nil": {
			l:     nil,
			r:     []rune{1, 2, 3, 4},
			equal: false,
		},
		"not equal right nil": {
			l:     []rune{1, 2, 3, 4},
			r:     nil,
			equal: false,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got, want := SliceEqual(tc.l, tc.r), tc.equal; got != want {
				t.Errorf("unexpected return, got: %v, want: %v", got, want)
			}
		})
	}
}

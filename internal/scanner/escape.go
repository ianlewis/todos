// Copyright 2024 Google LLC
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
	"slices"
)

// EscapeFunc is function that checks for escaped string characters.
type EscapeFunc func(s *CommentScanner, stringEnd []rune) ([]rune, error)

// NoEscape is a no-op.
func NoEscape(_ *CommentScanner, _ []rune) ([]rune, error) {
	return nil, nil
}

// CharEscape checks for the given escape character in the scanner.
func CharEscape(c rune) EscapeFunc {
	return func(s *CommentScanner, stringEnd []rune) ([]rune, error) {
		b := append([]rune{c}, stringEnd...)
		eq, err := s.peekEqual(b)
		if err != nil {
			return nil, err
		}
		if eq {
			return b, nil
		}
		return nil, nil
	}
}

// DoubleEscape checks for a repeated string ending character in the string.
func DoubleEscape(s *CommentScanner, stringEnd []rune) ([]rune, error) {
	b := slices.Clone(stringEnd)
	b = append(b, stringEnd...)
	eq, err := s.peekEqual(b)
	if err != nil {
		return nil, err
	}
	if eq {
		return b, nil
	}
	return nil, nil
}

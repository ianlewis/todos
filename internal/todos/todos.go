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

package todos

import (
	"regexp"

	"github.com/ianlewis/todos/internal/scanner"
)

var todoMatch = []*regexp.Regexp{
	// Empty comment
	regexp.MustCompile(`(TODO|FIXME|BUG|HACK|XXX)\s*$`),

	// Comment
	regexp.MustCompile(`(TODO|FIXME|BUG|HACK|XXX): .*`),

	// With Link
	regexp.MustCompile(`(TODO|FIXME|BUG|HACK|XXX)\(.*\): .*`),
}

// TODO is a todo comment.
type TODO struct {
	// Line is the line number where todo was found..
	Line int

	// Type is the todo type, such as "FIXME", "BUG", etc.
	Type string

	// Text is the full comment text.
	Text string
}

// TODOScanner scans for TODO comments.
type TODOScanner struct {
	next TODO
	s    *scanner.CommentScanner
}

// NewTODOScanner returns a new TODOScanner.
func NewTODOScanner(s *scanner.CommentScanner) *TODOScanner {
	return &TODOScanner{
		s: s,
	}
}

// Scan scans for the next TODO.
func (t *TODOScanner) Scan() bool {
	for t.s.Scan() {
		next := t.s.Next()
		text := next.String()

		match := findMatch(text)
		if match != "" {
			t.next = TODO{
				Line: next.Line(),
				Type: match,
				Text: next.String(),
			}
			return true
		}
	}
	return false
}

func findMatch(text string) string {
	for _, m := range todoMatch {
		match := m.FindAllStringSubmatch(text, 1)
		if len(match) != 0 {
			return match[0][1]
		}
	}
	return ""
}

// Next returns the next TODO.
func (t *TODOScanner) Next() TODO {
	return t.next
}

// Err returns the first error encountered.
func (t *TODOScanner) Err() error {
	return t.s.Err()
}

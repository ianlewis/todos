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
	"strings"

	"github.com/ianlewis/todos/internal/scanner"
)

// TODO is a todo comment.
type TODO struct {
	// Type is the todo type, such as "FIXME", "BUG", etc.
	Type string

	// Text is the full comment text.
	Text string

	// Line is the line number where todo was found..
	Line int

	// CommentLine is the line where the comment starts.
	CommentLine int
}

// Config is configuration for the TODOScanner.
type Config struct {
	Types []string
}

// CommentScanner is a type that scans code text for comments.
type CommentScanner interface {
	// Scan scans for the next comment. It returns true if there is more data
	// to scan.
	Scan() bool

	// Next returns the next Comment.
	Next() *scanner.Comment

	// Err returns an error if one occurred.
	Err() error
}

// TODOScanner scans for TODO comments.
type TODOScanner struct {
	next      *TODO
	s         CommentScanner
	todoMatch []*regexp.Regexp
}

// NewTODOScanner returns a new TODOScanner.
func NewTODOScanner(s CommentScanner, config *Config) *TODOScanner {
	snr := &TODOScanner{
		s: s,
	}

	var quotedTypes []string
	for _, tp := range config.Types {
		quotedTypes = append(quotedTypes, regexp.QuoteMeta(tp))
	}

	typesMatch := strings.Join(quotedTypes, "|")
	snr.todoMatch = []*regexp.Regexp{
		// Empty comment
		regexp.MustCompile(`(` + typesMatch + `)\s*$`),

		// Comment
		regexp.MustCompile(`(` + typesMatch + `): .*`),

		// With Link
		regexp.MustCompile(`(` + typesMatch + `)\(.*\): .*`),
	}

	return snr
}

// Scan scans for the next TODO.
func (t *TODOScanner) Scan() bool {
	for t.s.Scan() {
		next := t.s.Next()
		text := next.String()

		match, text, lineNo := t.findMatch(text)
		if match != "" {
			t.next = &TODO{
				Type: match,
				Text: text,
				// Add the line relative to the file.
				Line:        next.Line + lineNo - 1,
				CommentLine: next.Line,
			}
			return true
		}
	}
	return false
}

// findMatch returns the TODO type, the full TODO line, and the line number it
// was found on or zero if it was not found.
func (t *TODOScanner) findMatch(text string) (string, string, int) {
	for i, line := range strings.Split(text, "\n") {
		for _, m := range t.todoMatch {
			match := m.FindAllStringSubmatch(line, 1)
			if len(match) != 0 {
				return match[0][1], line, i + 1
			}
		}
	}

	return "", "", 0
}

// Next returns the next TODO.
func (t *TODOScanner) Next() *TODO {
	return t.next
}

// Err returns the first error encountered.
func (t *TODOScanner) Err() error {
	//nolint:wrapcheck
	return t.s.Err()
}

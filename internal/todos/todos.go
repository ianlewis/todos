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

// DefaultTypes is the default set of TODO types.
//
//nolint:gochecknoglobals // DefaultTypes is an overridable default.
var DefaultTypes = []string{
	"TODO",
	"Todo",
	"todo",
	"FIXME",
	"Fixme",
	"fixme",
	"BUG",
	"Bug",
	"bug",
	"HACK",
	"Hack",
	"hack",
	"XXX",
	"COMBAK",
}

// TODO is a todo comment.
type TODO struct {
	// Type is the todo type, such as "FIXME", "BUG", etc.
	Type string

	// Text is the full text of the TODO. For single line comments this is the
	// whole comment. For multi-line comments this is the line where the TODO
	// appears.
	Text string

	// Label is the label part (the part in parenthesis)
	Label string

	// Message is the comment message (the part after the parenthesis).
	Message string

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
	// Config return the configuration.
	Config() *scanner.Config

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
	next           []*TODO
	s              CommentScanner
	lineMatch      []*regexp.Regexp
	multilineMatch *regexp.Regexp
}

// NewTODOScanner returns a new TODOScanner.
func NewTODOScanner(s CommentScanner, config *Config) *TODOScanner {
	snr := &TODOScanner{
		s: s,
	}

	if config == nil {
		config = &Config{
			Types: DefaultTypes,
		}
	}

	quotedTypes := make([]string, 0, len(config.Types))
	for _, tp := range config.Types {
		quotedTypes = append(quotedTypes, regexp.QuoteMeta(tp))
	}
	// match[0][1]
	typesMatch := strings.Join(quotedTypes, "|")

	msgMatch := strings.Join([]string{
		`\s*`,                       // Naked
		`\s*[:\-/]+\s*(.*)`,         // With message (match[0][3])
		`\((.*)\)\s*`,               // Naked w/ label (match[0][4])
		`\((.*)\)\s*[:\-/]*\s*(.*)`, // With label (match[0][5]) and message (match[0][6])
	}, "|")

	msgMatch2 := strings.Join([]string{
		`\s*`,                       // Naked
		`\s*[:\-/]*\s*(.*)`,         // With message (match[0][3])
		`\((.*)\)\s*`,               // Naked w/ label (match[0][4])
		`\((.*)\)\s*[:\-/]*\s*(.*)`, // With label (match[0][5]) and message (match[0][6])
	}, "|")

	snr.lineMatch = []*regexp.Regexp{
		regexp.MustCompile(`^\s*@?(` + typesMatch + `)(` + msgMatch + `)$`),
		regexp.MustCompile(`^@?(` + typesMatch + `)(` + msgMatch2 + `)$`),
	}
	snr.multilineMatch = regexp.MustCompile(
		`^\s*\*?\s*@?(` + typesMatch + `)(` + msgMatch + `)$`)

	return snr
}

// Scan scans for the next TODO.
func (t *TODOScanner) Scan() bool {
	if len(t.next) > 0 {
		t.next = t.next[1:]
		if len(t.next) > 0 {
			return true
		}
	}

	for t.s.Scan() {
		next := t.s.Next()

		if next.Multiline {
			matches := t.findMultilineMatches(next)
			t.next = append(t.next, matches...)
			if len(t.next) > 0 {
				return true
			}
		} else {
			match := t.findLineMatch(next)
			if match != nil {
				t.next = append(t.next, match)
				return true
			}
		}
	}
	return false
}

// findMultilineMatch returns the TODO for the comment if it was found.
func (t *TODOScanner) findMultilineMatches(c *scanner.Comment) []*TODO {
	var matches []*TODO
	var i int
	for line := range strings.SplitSeq(strings.TrimSpace(c.Text), "\n") {
		lineText := strings.TrimLeft(line, string(c.MultilineConfig.Start))
		lineText = strings.TrimRight(lineText, string(c.MultilineConfig.End))

		match := t.multilineMatch.FindAllStringSubmatch(lineText, 1)
		if len(match) != 0 && len(match[0]) > 2 && match[0][1] != "" {
			label := match[0][4]
			if label == "" {
				label = match[0][5]
			}

			message := match[0][3]
			if message == "" {
				message = match[0][6]
			}

			matches = append(matches, &TODO{
				Type:    strings.TrimSpace(match[0][1]),
				Text:    strings.TrimSpace(line),
				Label:   strings.TrimSpace(label),
				Message: strings.TrimSpace(message),
				// Add the line relative to the file.
				Line:        c.Line + i,
				CommentLine: c.Line,
			})
		}
		i++
	}
	return matches
}

// findLineMatch returns the TODO for the comment if it was found.
func (t *TODOScanner) findLineMatch(c *scanner.Comment) *TODO {
	trimmedText := strings.TrimSpace(c.Text)
	commentText := strings.TrimLeft(trimmedText, string(c.LineConfig.Start))
	for _, lnMatch := range t.lineMatch {
		match := lnMatch.FindAllStringSubmatch(commentText, 1)
		if len(match) != 0 && len(match[0]) > 2 && match[0][1] != "" {
			label := match[0][4]
			if label == "" {
				label = match[0][5]
			}

			message := match[0][3]
			if message == "" {
				message = match[0][6]
			}

			return &TODO{
				Type:    match[0][1],
				Text:    trimmedText,
				Label:   strings.TrimSpace(label),
				Message: strings.TrimSpace(message),
				// Add the line relative to the file.
				Line:        c.Line,
				CommentLine: c.Line,
			}
		}
	}

	return nil
}

// Next returns the next TODO.
func (t *TODOScanner) Next() *TODO {
	if len(t.next) > 0 {
		return t.next[0]
	}
	return nil
}

// Err returns the first error encountered.
func (t *TODOScanner) Err() error {
	//nolint:wrapcheck
	return t.s.Err()
}

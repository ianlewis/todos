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
	next           *TODO
	s              CommentScanner
	lineMatch      []*regexp.Regexp
	multilineMatch *regexp.Regexp
}

// NewTODOScanner returns a new TODOScanner.
func NewTODOScanner(s CommentScanner, config *Config) *TODOScanner {
	snr := &TODOScanner{
		s: s,
	}

	sConfig := s.Config()
	var commentStarts []string
	for _, lineCommentStart := range sConfig.LineCommentStart {
		commentStarts = append(commentStarts, regexp.QuoteMeta(lineCommentStart))
	}
	commentStartMatch := strings.Join(commentStarts, "|")

	multiStartMatch := regexp.QuoteMeta(sConfig.MultilineComment.Start)

	if config == nil {
		config = &Config{
			Types: DefaultTypes,
		}
	}

	var quotedTypes []string
	for _, tp := range config.Types {
		quotedTypes = append(quotedTypes, regexp.QuoteMeta(tp))
	}
	// match[0][2]
	typesMatch := strings.Join(quotedTypes, "|")

	msgMatch := strings.Join([]string{
		`\s*`,                       // Naked
		`\s*[:\-/]+\s*(.*)`,         // With message (match[0][4])
		`\((.*)\)\s*`,               // Naked w/ label (match[0][5])
		`\((.*)\)\s*[:\-/]*\s*(.*)`, // With label (match[0][6]) and message (match[0][7])
	}, "|")

	msgMatch2 := strings.Join([]string{
		`\s*`,                       // Naked
		`\s*[:\-/]*\s*(.*)`,         // With message (match[0][4])
		`\((.*)\)\s*`,               // Naked w/ label (match[0][5])
		`\((.*)\)\s*[:\-/]*\s*(.*)`, // With label (match[0][6]) and message (match[0][7])
	}, "|")

	snr.lineMatch = []*regexp.Regexp{
		regexp.MustCompile(`^\s*(` + commentStartMatch + `)\s*(` + typesMatch + `)(` + msgMatch + `)$`),
		regexp.MustCompile(`^\s*(` + commentStartMatch + `)(` + typesMatch + `)(` + msgMatch2 + `)$`),
	}
	snr.multilineMatch = regexp.MustCompile(
		`^(` + multiStartMatch + `\s*|\s*\*?\s*)?(` + typesMatch + `)(` + msgMatch + `)$`)

	return snr
}

// Scan scans for the next TODO.
func (t *TODOScanner) Scan() bool {
	for t.s.Scan() {
		next := t.s.Next()

		match := t.findMatch(next)
		if match != nil {
			t.next = match
			return true
		}
	}
	return false
}

// findMatch returns the TODO type, the full TODO line, the label, message, and
// the line number it was found on or zero if it was not found.
func (t *TODOScanner) findMatch(c *scanner.Comment) *TODO {
	if c.Multiline {
		for i, line := range strings.Split(c.Text, "\n") {
			match := t.multilineMatch.FindAllStringSubmatch(line, 1)
			if len(match) != 0 && len(match[0]) > 2 && match[0][2] != "" {
				label := match[0][5]
				if label == "" {
					label = match[0][6]
				}

				message := match[0][4]
				if message == "" {
					message = match[0][7]
				}

				return &TODO{
					Type:    match[0][2],
					Text:    strings.TrimSpace(line),
					Label:   strings.TrimSpace(label),
					Message: strings.TrimSpace(message),
					// Add the line relative to the file.
					Line:        c.Line + i,
					CommentLine: c.Line,
				}
			}
		}
	}

	for _, lnMatch := range t.lineMatch {
		match := lnMatch.FindAllStringSubmatch(c.Text, 1)
		if len(match) != 0 && len(match[0]) > 2 && match[0][2] != "" {
			label := match[0][5]
			if label == "" {
				label = match[0][6]
			}

			message := match[0][4]
			if message == "" {
				message = match[0][7]
			}

			return &TODO{
				Type:    match[0][2],
				Text:    strings.TrimSpace(c.Text),
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
	return t.next
}

// Err returns the first error encountered.
func (t *TODOScanner) Err() error {
	//nolint:wrapcheck
	return t.s.Err()
}

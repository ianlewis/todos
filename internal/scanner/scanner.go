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
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/ianlewis/linguist"
	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/ianaindex"
)

var (
	// errDetectCharset is an error detecting a charset.
	errDetectCharset = errors.New("detect charset")

	// errDecodeCharset is an error when decoding a charset.
	errDecodeCharset = errors.New("decoding charset")

	// linguistLock manages a global lock on lingist language detection which
	// is not thread-safe.
	linguistLock sync.Mutex
)

// config is configuration for a generic comment scanner.
type config struct {
	LineCommentStart      [][]rune
	MultilineCommentStart []rune
	MultilineCommentEnd   []rune
	Strings               [][2][]rune
}

func convertConfig(c *Config) *config {
	var c2 config
	c2.LineCommentStart = stringsToRunes(c.LineCommentStart)
	c2.MultilineCommentStart = []rune(c.MultilineCommentStart)
	c2.MultilineCommentEnd = []rune(c.MultilineCommentEnd)
	for i := range c.Strings {
		c2.Strings = append(c2.Strings, [2][]rune{
			[]rune(c.Strings[i][0]),
			[]rune(c.Strings[i][1]),
		})
	}
	return &c2
}

func stringsToRunes(s []string) [][]rune {
	var r [][]rune
	for i := range s {
		r = append(r, []rune(s[i]))
	}
	return r
}

// FromFile returns an appropriate CommentScanner for the given file. The
// language is auto-detected and a relevant configuration is used to initialize the scanner.
func FromFile(f *os.File) (*CommentScanner, error) {
	rawContents, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", f.Name(), err)
	}

	return FromBytes(f.Name(), rawContents)
}

// FromBytes returns an appropriate CommentScanner for the given contents. The
// language is auto-detected and a relevant configuration is used to initialize the scanner.
func FromBytes(fileName string, rawContents []byte) (*CommentScanner, error) {
	// Detect the character set.
	det := chardet.NewTextDetector()
	result, err := det.DetectBest(rawContents)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errDetectCharset, err)
	}

	// If we just detect it as ascii (latin1) then treat it as UTF-8 since they
	// are compatible.
	charset := result.Charset
	if charset == "ISO-8859-1" {
		charset = "UTF-8"
	}
	// See: https://github.com/saintfish/chardet/issues/2
	if charset == "GB-18030" {
		charset = "GB18030"
	}

	e, err := ianaindex.IANA.Encoding(charset)
	if err != nil {
		// The detected charset should not be unregistered.
		panic(fmt.Errorf("%w: %s: %w", errDecodeCharset, charset, err))
	}
	if e == nil {
		return nil, fmt.Errorf("%w: %s: unsupported character set", errDecodeCharset, charset)
	}

	decodedContents, err := e.NewDecoder().Bytes(rawContents)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", errDecodeCharset, charset, err)
	}

	if linguist.ShouldIgnoreContents(decodedContents) {
		return nil, nil
	}

	// Detect the programming language.
	linguistLock.Lock()
	defer linguistLock.Unlock()
	lang := linguist.LanguageByContents(decodedContents, linguist.LanguageHints(fileName))

	// Detect the language encoding.
	if config, ok := languageMap[lang]; ok {
		return New(bytes.NewReader(decodedContents), config), nil
	}

	return nil, nil
}

// New returns a new CommentScanner that scans code returned by r with the given Config.
func New(r io.Reader, c *Config) *CommentScanner {
	return &CommentScanner{
		originalConfig: c,
		config:         convertConfig(c),
		reader:         NewRuneReader(r),

		// Starting state
		state: &stateCode{},
		line:  1, // NOTE: lines are 1 indexed
	}
}

// CommentScanner is a generic code comment scanner.
type CommentScanner struct {
	reader         *RuneReader
	originalConfig *Config
	config         *config

	// state is the current state-machine state.
	state state

	// line is the current line in the input.
	line int

	// next is the next comment to be returned by Next.
	next *Comment

	// err is the error returned by Err.
	err error
}

// Config returns the scanners configuration.
func (s *CommentScanner) Config() *Config {
	return s.originalConfig
}

// Next returns the next Comment.
func (s *CommentScanner) Next() *Comment {
	return s.next
}

// Err returns an error if one occurred.
func (s *CommentScanner) Err() error {
	if errors.Is(s.err, io.EOF) {
		return nil
	}
	return s.err
}

// Scan implements a simple state machine to parse comments out of generic
// code.
func (s *CommentScanner) Scan() bool {
	for {
		if s.err != nil {
			return false
		}

		switch st := s.state.(type) {
		case *stateCode:
			s.state, s.err = s.processCode(st)
		case *stateString:
			s.state, s.err = s.processString(st)
		case *stateLineComment:
			s.state, s.err = s.processLineComment(st)
			if _, ok := s.state.(*stateLineComment); !ok {
				return true
			}
		case *stateMultilineComment:
			s.state, s.err = s.processMultilineComment(st)
			if _, ok := s.state.(*stateMultilineComment); !ok {
				return true
			}
		}
	}
}

// processCode processes source code and returns the next state.
func (s *CommentScanner) processCode(st *stateCode) (state, error) {
	for {
		// Check for line comment
		for _, start := range s.config.LineCommentStart {
			eq, err := s.peekEqual(start)
			if err != nil {
				return st, err
			}
			if eq {
				return &stateLineComment{}, nil
			}
		}

		// Check for line comment
		if len(s.config.MultilineCommentStart) != 0 {
			if eq, err := s.peekEqual(s.config.MultilineCommentStart); err == nil && eq {
				// Discard the opening. It will be added by processMultilineComment.
				if _, errDiscard := s.reader.Discard(len(s.config.MultilineCommentStart)); errDiscard != nil {
					return st, errDiscard
				}
				return &stateMultilineComment{
					line: s.line,
				}, nil
			} else if err != nil {
				return st, err
			}
		}

		// Check for string
		for i, strs := range s.config.Strings {
			eq, err := s.peekEqual(strs[0])
			if err != nil {
				return st, err
			}
			if eq {
				// Discard the string opening.
				if _, err := s.reader.Discard(len(strs[0])); err != nil {
					return st, err
				}
				return &stateString{
					index: i,
				}, nil
			}
		}

		// Process the next rune.
		if _, err := s.nextRune(); err != nil {
			return st, err
		}
	}
}

// processString processes strings and returns the next state.
func (s *CommentScanner) processString(st *stateString) (state, error) {
	for {
		// Handle escaped characters.
		escaped, err := s.peekEqual([]rune{'\\'})
		if err != nil {
			return st, err
		}
		if escaped {
			// Skip the backslash character.
			if _, err := s.reader.Discard(1); err != nil {
				return st, err
			}
		} else {
			// Look for the end of the string.
			stringEnd, err := s.peekEqual(s.config.Strings[st.index][1])
			if err != nil {
				return st, err
			}
			if stringEnd {
				if _, err := s.reader.Discard(len(s.config.Strings[st.index][1])); err != nil {
					return st, err
				}
				return &stateCode{}, nil
			}
		}

		if _, err := s.nextRune(); err != nil {
			return st, err
		}
	}
}

// processLineComment processes line comments and returns the next state.
func (s *CommentScanner) processLineComment(st *stateLineComment) (state, error) {
	var b strings.Builder
	for {
		lineEnd, err := s.isLineEnd()
		if err != nil {
			return st, err
		}
		if lineEnd {
			s.next = &Comment{
				Text:      b.String(),
				Line:      s.line,
				Multiline: false,
			}
			return &stateCode{}, nil
		}

		rn, err := s.nextRune()
		if err != nil {
			return st, err
		}

		_, err = b.WriteRune(rn)
		if err != nil {
			return st, fmt.Errorf("writing rune %q: %w", rn, err)
		}
	}
}

// processMultilineComment processes multi-line comments and returns the next state.
func (s *CommentScanner) processMultilineComment(st *stateMultilineComment) (state, error) {
	var b strings.Builder
	b.WriteString(string(s.config.MultilineCommentStart))
	for {
		// Look for the end of the comment.
		mlEnd, err := s.peekEqual(s.config.MultilineCommentEnd)
		if err != nil {
			return st, err
		}
		if mlEnd {
			if _, errDiscard := s.reader.Discard(len(s.config.MultilineCommentEnd)); errDiscard != nil {
				return st, errDiscard
			}
			// Add the ending to the builder.
			b.WriteString(string(s.config.MultilineCommentEnd))
			s.next = &Comment{
				Text:      b.String(),
				Line:      st.line,
				Multiline: true,
			}
			return &stateCode{}, nil
		}

		rn, err := s.nextRune()
		if err != nil {
			return st, err
		}

		_, err = b.WriteRune(rn)
		if err != nil {
			return st, fmt.Errorf("writing rune %q: %w", rn, err)
		}
	}
}

func (s *CommentScanner) nextRune() (rune, error) {
	rn, err := s.reader.Next()
	if err != nil {
		return rn, err
	}
	if rn == '\n' {
		s.line++
	}
	return rn, nil
}

func (s *CommentScanner) isLineEnd() (bool, error) {
	nixNL, err := s.peekEqual([]rune{'\n'})
	if errors.Is(err, io.EOF) {
		// NOTE: if we get EOF here we have no more runes to read. Handle EOF
		// as the end of a line.
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if nixNL {
		return true, nil
	}

	winNL, err := s.peekEqual([]rune{'\r', '\n'})
	if errors.Is(err, io.EOF) {
		// NOTE: We may only have one character left so return false so it can
		// be processed.
		return false, nil
	}
	return winNL, err
}

func (s *CommentScanner) peekEqual(val []rune) (bool, error) {
	runes, err := s.reader.Peek(len(val))
	if err != nil {
		return false, err
	}
	return compareRunes(runes, val), nil
}

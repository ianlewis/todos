// Copyright 2023 Google LLC
// Copyright 2025 Ian Lewis
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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/ianlewis/runeio"
	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/ianaindex"
)

var (
	// errDetectCharset is an error detecting a charset.
	errDetectCharset = errors.New("detect charset")

	// errDecodeCharset is an error when decoding a charset.
	errDecodeCharset = errors.New("decoding charset")
)

var (
	// ErrUnsupportedCharset indicates that the detected language is not
	// supported.
	ErrUnsupportedLanguage = errors.New("unsupported language")

	// ErrBinaryFile indicates that the file is a binary file.
	ErrBinaryFile = errors.New("binary file")
)

type StringConfig struct {
	Start      []rune
	End        []rune
	EscapeFunc EscapeFunc
}

type LineCommentConfig struct {
	// Start is the starting sequence for the line comment.
	Start []rune
}

type MultilineCommentConfig struct {
	// Start is the starting sequence for the multiline comment.
	Start []rune

	// End is the ending sequence for the multiline comment.
	End []rune

	// AtLineStart indicates that the multiline comment must start at the
	// beginning of a line.
	AtLineStart bool

	// Nested indicates that the multi-line comment can be nested.
	Nested bool
}

// Config is configuration for a generic comment scanner.
type Config struct {
	LineComments      []LineCommentConfig
	MultilineComments []MultilineCommentConfig
	Strings           []StringConfig
}

// FromFile returns an appropriate CommentScanner for the given file. The
// language is auto-detected and a relevant configuration is used to initialize the scanner.
func FromFile(f *os.File, charset string) (*CommentScanner, error) {
	rawContents, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", f.Name(), err)
	}

	return FromBytes(f.Name(), rawContents, charset)
}

// FromBytes returns an appropriate CommentScanner for the given contents. The
// language is auto-detected and a relevant configuration is used to initialize the scanner.
func FromBytes(fileName string, rawContents []byte, charset string) (*CommentScanner, error) {
	// Ignore binary files.
	if enry.IsBinary(rawContents) {
		return nil, ErrBinaryFile
	}

	if charset == "detect" {
		// Detect the character set.
		det := chardet.NewTextDetector()
		result, err := det.DetectBest(rawContents)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errDetectCharset, err)
		}

		charset = result.Charset
	}

	// If given ascii (latin1) then treat it as UTF-8 since they
	// are compatible.
	if charset == "ISO-8859-1" {
		charset = "UTF-8"
	}
	// See: https://github.com/saintfish/chardet/issues/2
	if charset == "GB-18030" {
		charset = "GB18030"
	}

	e, err := ianaindex.IANA.Encoding(charset)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", errDecodeCharset, charset, err)
	}
	if e == nil {
		return nil, fmt.Errorf("%w: %s: unsupported character set", errDecodeCharset, charset)
	}

	decodedContents, err := e.NewDecoder().Bytes(rawContents)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", errDecodeCharset, charset, err)
	}

	// Detect the programming language.
	lang := enry.GetLanguage(fileName, decodedContents)
	if lang == enry.OtherLanguage {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedLanguage, lang)
	}

	// Detect the language encoding.
	config, ok := LanguagesConfig[lang]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedLanguage, lang)
	}

	return New(bytes.NewReader(decodedContents), config), nil
}

// New returns a new CommentScanner that scans code returned by r with the given Config.
func New(r io.Reader, c *Config) *CommentScanner {
	return &CommentScanner{
		config: c,
		reader: runeio.NewReader(bufio.NewReader(r)),

		// Starting state
		state: &stateCode{},
		line:  1, // NOTE: lines are 1 indexed
	}
}

// CommentScanner is a generic code comment scanner.
type CommentScanner struct {
	reader *runeio.RuneReader
	config *Config

	// state is the current state-machine state.
	state state

	// atLineStart indicates whether the next character is at the start of the
	// line.
	atLineStart bool

	// line is the current line in the input.
	line int

	// next is the next comment to be returned by Next.
	next *Comment

	// err is the error returned by Err.
	err error
}

// Config returns the scanners configuration.
func (s *CommentScanner) Config() *Config {
	return s.config
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
		case *stateLineCommentOrString:
			var hasComment bool
			hasComment, s.state, s.err = s.processLineCommentOrString(st)
			if hasComment {
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
		lcIndex, m, err := s.lineMatch()
		if err != nil {
			return st, err
		}

		mmIndex, mm, err := s.multiLineMatch()
		if err != nil {
			return st, err
		}

		// Check for line comments.
		if m != nil {
			// If both line comments and multi-line comments match, chose the
			// one with the longest start sequence.
			if mm == nil || len(m.Start) >= len(mm.Start) {
				for i, stringStart := range s.config.Strings {
					if string(stringStart.Start) == string(m.Start) {
						return &stateLineCommentOrString{
							index: i,
						}, nil
					}
				}

				return &stateLineComment{
					index: lcIndex,
				}, nil
			}
		}

		// Check for multi-line comments.
		if mm != nil {
			if !mm.AtLineStart || s.atLineStart {
				return &stateMultilineComment{
					line:  s.line,
					index: mmIndex,
				}, nil
			}
		}

		// Check for strings.
		for i, strs := range s.config.Strings {
			eq, err := s.peekEqual(strs.Start)
			if err != nil {
				return st, err
			}
			if eq {
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

func (s *CommentScanner) lineMatch() (int, *LineCommentConfig, error) {
	// Check for line comment
	for i, m := range s.config.LineComments {
		eq, err := s.peekEqual(m.Start)
		if err != nil {
			return 0, nil, err
		}
		if eq {
			return i, &m, nil
		}
	}
	return 0, nil, nil
}

func (s *CommentScanner) multiLineMatch() (int, *MultilineCommentConfig, error) {
	// Check for multiline comment
	for i, mlConfig := range s.config.MultilineComments {
		if eq, err := s.peekEqual(mlConfig.Start); err == nil && eq {
			return i, &mlConfig, nil
		} else if err != nil {
			return 0, nil, err
		}
	}
	return 0, nil, nil
}

// processString processes strings and returns the next state.
func (s *CommentScanner) processString(st *stateString) (state, error) {
	// Discard the string start characters.
	if _, err := s.reader.Discard(len(s.config.Strings[st.index].Start)); err != nil {
		return st, fmt.Errorf("parsing string: %w", err)
	}

	for {
		// Handle escaped characters.
		escaped, err := s.config.Strings[st.index].EscapeFunc(s, s.config.Strings[st.index].End)
		// There may still be characters to process so continue if we get EOF.
		if err != nil && !errors.Is(err, io.EOF) {
			return st, err
		}
		if len(escaped) > 0 {
			// Skip the escaped characters.
			if _, err := s.reader.Discard(len(escaped)); err != nil {
				return st, fmt.Errorf("parsing string: %w", err)
			}
		} else {
			// Look for the end of the string.
			stringEnd, err := s.peekEqual(s.config.Strings[st.index].End)
			if err != nil {
				return st, fmt.Errorf("parsing string: %w", err)
			}
			if stringEnd {
				if _, err := s.reader.Discard(len(s.config.Strings[st.index].End)); err != nil {
					return st, fmt.Errorf("parsing string: %w", err)
				}
				return &stateCode{}, nil
			}

			if _, err := s.nextRune(); err != nil {
				return st, fmt.Errorf("parsing string: %w", err)
			}
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
				Text:       b.String(),
				Line:       s.line,
				Multiline:  false,
				LineConfig: &s.config.LineComments[st.index],
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

// processLineCommentOrString processes strings or line comments when they have
// the same start character. e.g. Vim Script.
func (s *CommentScanner) processLineCommentOrString(st *stateLineCommentOrString) (bool, state, error) {
	// Discard the string start characters.
	if _, err := s.reader.Discard(len(s.config.Strings[st.index].Start)); err != nil {
		return false, st, fmt.Errorf("parsing string: %w", err)
	}

	// b is used to build the line comment text.
	var b strings.Builder
	// Add the opening to the builder since we want it in the output if this is a comment.
	b.WriteString(string(s.config.Strings[st.index].Start))
	for {
		lineEnd, err := s.isLineEnd()
		if err != nil {
			return false, st, err
		}

		// If we get to the end of the line without the string being closed,
		// then it must be a comment. Languages where line comments and strings
		// share the same character cannot implement multi-line strings.
		if lineEnd {
			s.next = &Comment{
				Text:       b.String(),
				Line:       s.line,
				Multiline:  false,
				LineConfig: &s.config.LineComments[st.index],
			}
			return true, &stateCode{}, nil
		}

		// Handle escaped characters.
		escaped, err := s.config.Strings[st.index].EscapeFunc(s, s.config.Strings[st.index].End)
		// There may still be characters to process so continue.
		if err != nil && !errors.Is(err, io.EOF) {
			return false, st, err
		}
		if len(escaped) > 0 {
			// Skip the escaped characters.
			if _, discardErr := s.reader.Discard(len(escaped)); discardErr != nil {
				return false, st, fmt.Errorf("parsing string: %w", discardErr)
			}

			// Write the escaped characters in case this is a comment.
			_, err = b.WriteString(string(escaped))
			if err != nil {
				return false, st, fmt.Errorf("writing runes %q: %w", escaped, err)
			}
			continue
		}

		// Look for the end of the string.
		stringEnd, err := s.peekEqual(s.config.Strings[st.index].End)
		if err != nil {
			return false, st, fmt.Errorf("parsing string: %w", err)
		}
		if stringEnd {
			if _, discardErr := s.reader.Discard(len(s.config.Strings[st.index].End)); discardErr != nil {
				return false, st, fmt.Errorf("parsing string: %w", discardErr)
			}
			return false, &stateCode{}, nil
		}

		rn, err := s.nextRune()
		if err != nil {
			return false, st, fmt.Errorf("parsing string: %w", err)
		}

		_, err = b.WriteRune(rn)
		if err != nil {
			return false, st, fmt.Errorf("writing rune %q: %w", rn, err)
		}
	}
}

// processMultilineComment processes multi-line comments and returns the next state.
func (s *CommentScanner) processMultilineComment(st *stateMultilineComment) (state, error) {
	mm := s.config.MultilineComments[st.index]

	// Discard the opening since we don't want to parse it. It could be the same as the closing.
	if _, errDiscard := s.reader.Discard(len(mm.Start)); errDiscard != nil {
		return st, fmt.Errorf("parsing code: %w", errDiscard)
	}

	var b strings.Builder
	var nestingDepth int

	// Add the opening to the builder since we want it in the output.
	b.WriteString(string(mm.Start))
	for {
		if mm.Nested {
			// Look for a nested comment start.
			mlStart, err := s.peekEqual(mm.Start)
			if err != nil {
				return st, err
			}
			if mlStart && (!mm.AtLineStart || s.atLineStart) {
				if _, errDiscard := s.reader.Discard(len(mm.Start)); errDiscard != nil {
					return st, fmt.Errorf("parsing multi-line comment: %w", errDiscard)
				}
				// Add the start to the builder.
				b.WriteString(string(mm.Start))
				nestingDepth++
				continue
			}
		}

		// Look for the end of the comment.
		mlEnd, err := s.peekEqual(mm.End)
		if err != nil {
			return st, err
		}
		if mlEnd && (!mm.AtLineStart || s.atLineStart) {
			if _, errDiscard := s.reader.Discard(len(mm.End)); errDiscard != nil {
				return st, fmt.Errorf("parsing multi-line comment: %w", errDiscard)
			}
			// Add the ending to the builder.
			b.WriteString(string(mm.End))
			if nestingDepth == 0 {
				s.next = &Comment{
					Text:            b.String(),
					Line:            st.line,
					Multiline:       true,
					MultilineConfig: &s.config.MultilineComments[st.index],
				}
				return &stateCode{}, nil
			}
			if mm.Nested {
				nestingDepth--
			}
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
	rn, _, err := s.reader.ReadRune()
	if err != nil {
		return rn, fmt.Errorf("reading rune: %w", err)
	}
	if rn == '\n' {
		s.line++
		s.atLineStart = true
	} else {
		s.atLineStart = false
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
	r, err := s.reader.Peek(len(val))
	if err != nil {
		return false, fmt.Errorf("reading rune: %w", err)
	}
	return slices.Equal(r, val), nil
}

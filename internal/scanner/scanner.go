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

// Package scanner implements a generic comment scanner for programming
// languages based on a configuration that describes the language.
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
	"unicode"

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
	// ErrUnsupportedLanguage indicates that the detected language is not
	// supported.
	ErrUnsupportedLanguage = errors.New("unsupported language")

	// ErrBinaryFile indicates that the file is a binary file.
	ErrBinaryFile = errors.New("binary file")
)

// StringConfig is a configuration for a string literal.
type StringConfig struct {
	// Start is the starting sequence for the string literal.
	Start []rune

	// End is the ending sequence for the string literal.
	End []rune

	// EscapeFunc is a function that checks for escaped string characters.
	EscapeFunc EscapeFunc
}

// LineCommentConfig is a configuration for a line comment.
type LineCommentConfig struct {
	// Start is the starting sequence for the line comment.
	Start []rune

	// AtLineStart indicates that the comment must start at the beginning of a
	// line.
	AtLineStart bool
}

// MultilineCommentConfig is a configuration for a multi-line comment.
type MultilineCommentConfig struct {
	// Start is the starting sequence for the multiline comment.
	Start []rune

	// End is the ending sequence for the multiline comment.
	End []rune

	// AtFirstColumn indicates that the multiline comment must start at the
	// first column of a line.
	AtFirstColumn bool

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
func FromFile(f *os.File, lang, charset string) (*CommentScanner, error) {
	rawContents, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", f.Name(), err)
	}

	return FromBytes(f.Name(), rawContents, lang, charset)
}

// FromBytes returns an appropriate CommentScanner for the given contents. The
// language is auto-detected and a relevant configuration is used to initialize the scanner.
func FromBytes(fileName string, rawContents []byte, lang, charset string) (*CommentScanner, error) {
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
	if lang != "" {
		lang, _ = enry.GetLanguageByAlias(lang)
	} else {
		lang = enry.GetLanguage(fileName, decodedContents)
	}

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

		atLineStart:   true,
		atFirstColumn: true,

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

	// atFirstColumn indicates whether the next character is at the first
	// column of a line.
	atFirstColumn bool

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
//
//nolint:ireturn,nolintlint // returning interface required for state machine.
func (s *CommentScanner) processCode(state *stateCode) (state, error) {
	for {
		// Check for line comment
		lcIndex, m, err := s.lineMatch()
		if err != nil {
			return state, err
		}

		lcLen := 0
		if m != nil {
			lcLen = len(m.Start)
		}

		mmIndex, mm, err := s.multilineMatch()
		if err != nil {
			return state, err
		}

		mmLen := 0
		if mm != nil {
			mmLen = len(mm.Start)
		}

		sIndex, strs, err := s.stringMatch()
		if err != nil {
			return state, err
		}

		strsLen := 0
		if strs != nil {
			strsLen = len(strs.Start)
		}

		// Generally the longest match gets precedence. However, if two matches
		// are the same length (e.g. the same characters), some special
		// processing may need to be done.
		maxLen := max(lcLen, mmLen, strsLen)
		if maxLen == 0 {
			// There was no match, process the next rune.
			if _, err := s.nextRune(); err != nil {
				return state, err
			}

			continue
		}

		switch {
		case mmLen == maxLen && lcLen < maxLen && strsLen < maxLen:
			// multiline comment
			return &stateMultilineComment{
				line:  s.line,
				index: mmIndex,
			}, nil
		case lcLen == maxLen && mmLen < maxLen && strsLen < maxLen:
			// line comment
			return &stateLineComment{
				index: lcIndex,
			}, nil
		case strsLen == maxLen && lcLen < maxLen && mmLen < maxLen:
			// string
			return &stateString{
				index: sIndex,
			}, nil
		case mmLen == maxLen && lcLen == maxLen && strsLen < maxLen:
			// multiline comment and line comment
			// TODO(#1722): Handle when multi-line and line comments have
			// 				the same start sequence
			return &stateMultilineComment{
				line:  s.line,
				index: mmIndex,
			}, nil
		case mmLen == maxLen && strsLen == maxLen && lcLen < maxLen:
			// multiline comment and string
			// TODO(#1723): Handle when multi-line and line comments have
			// 				the same start sequence
			return &stateMultilineComment{
				line:  s.line,
				index: mmIndex,
			}, nil
		case lcLen == maxLen && strsLen == maxLen && mmLen < maxLen:
			// line comment and string
			return &stateLineCommentOrString{
				lcIndex: lcIndex,
				sIndex:  sIndex,
			}, nil
		default:
			// multiline comment and line comment and string
			// TODO(#1723): Handle when multi-line and line comments have
			// 				the same start sequence
			return &stateMultilineComment{
				line:  s.line,
				index: mmIndex,
			}, nil
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

		if eq && (!m.AtLineStart || s.atLineStart) {
			return i, &m, nil
		}
	}

	return 0, nil, nil
}

func (s *CommentScanner) multilineMatch() (int, *MultilineCommentConfig, error) {
	// Check for multiline comment
	for i, mlConfig := range s.config.MultilineComments {
		if eq, err := s.peekEqual(mlConfig.Start); err == nil && eq {
			if !mlConfig.AtFirstColumn || s.atFirstColumn {
				return i, &mlConfig, nil
			}
		} else if err != nil {
			return 0, nil, err
		}
	}

	return 0, nil, nil
}

func (s *CommentScanner) stringMatch() (int, *StringConfig, error) {
	for i, strs := range s.config.Strings {
		eq, err := s.peekEqual(strs.Start)
		if err != nil {
			return i, &strs, err
		}

		if eq {
			return i, &strs, nil
		}
	}

	return 0, nil, nil
}

// processString processes strings and returns the next state.
//
//nolint:ireturn,nolintlint // returning interface required for state machine.
func (s *CommentScanner) processString(state *stateString) (state, error) {
	// Discard the string start characters.
	if _, err := s.reader.Discard(len(s.config.Strings[state.index].Start)); err != nil {
		return state, fmt.Errorf("parsing string: %w", err)
	}

	for {
		// Handle escaped characters.
		escaped, err := s.config.Strings[state.index].EscapeFunc(s, s.config.Strings[state.index].End)
		// There may still be characters to process so continue if we get EOF.
		if err != nil && !errors.Is(err, io.EOF) {
			return state, err
		}

		if len(escaped) > 0 {
			// Skip the escaped characters.
			if _, err := s.reader.Discard(len(escaped)); err != nil {
				return state, fmt.Errorf("parsing string: %w", err)
			}
		} else {
			// Look for the end of the string.
			stringEnd, err := s.peekEqual(s.config.Strings[state.index].End)
			if err != nil {
				return state, fmt.Errorf("parsing string: %w", err)
			}

			if stringEnd {
				if _, err := s.reader.Discard(len(s.config.Strings[state.index].End)); err != nil {
					return state, fmt.Errorf("parsing string: %w", err)
				}

				return &stateCode{}, nil
			}

			if _, err := s.nextRune(); err != nil {
				return state, fmt.Errorf("parsing string: %w", err)
			}
		}
	}
}

// processLineComment processes line comments and returns the next state.
//
//nolint:ireturn,nolintlint // returning interface required for state machine.
func (s *CommentScanner) processLineComment(state *stateLineComment) (state, error) {
	var b strings.Builder

	for {
		lineEnd, err := s.isLineEnd()
		if err != nil {
			return state, err
		}

		if lineEnd {
			s.next = &Comment{
				Text:       b.String(),
				Line:       s.line,
				Multiline:  false,
				LineConfig: &s.config.LineComments[state.index],
			}

			return &stateCode{}, nil
		}

		rn, err := s.nextRune()
		if err != nil {
			return state, err
		}

		_, err = b.WriteRune(rn)
		if err != nil {
			return state, fmt.Errorf("writing rune %q: %w", rn, err)
		}
	}
}

// processLineCommentOrString processes strings or line comments when they have
// the same start character. e.g. Vim Script.
//
//nolint:ireturn,nolintlint // returning interface required for state machine.
func (s *CommentScanner) processLineCommentOrString(state *stateLineCommentOrString) (bool, state, error) {
	// Discard the string start characters.
	if _, err := s.reader.Discard(len(s.config.Strings[state.sIndex].Start)); err != nil {
		return false, state, fmt.Errorf("parsing string: %w", err)
	}

	// commentTxt is used to build the line comment text.
	var commentTxt strings.Builder
	// Add the opening to the builder since we want it in the output if this is a comment.
	commentTxt.WriteString(string(s.config.Strings[state.sIndex].Start))

	for {
		lineEnd, err := s.isLineEnd()
		if err != nil {
			return false, state, err
		}

		// If we get to the end of the line without the string being closed,
		// then it must be a comment. Languages where line comments and strings
		// share the same character cannot implement multi-line strings.
		if lineEnd {
			s.next = &Comment{
				Text:       commentTxt.String(),
				Line:       s.line,
				Multiline:  false,
				LineConfig: &s.config.LineComments[state.lcIndex],
			}

			return true, &stateCode{}, nil
		}

		// Handle escaped characters.
		escaped, err := s.config.Strings[state.sIndex].EscapeFunc(
			s,
			s.config.Strings[state.sIndex].End,
		)
		// There may still be characters to process so continue.
		if err != nil && !errors.Is(err, io.EOF) {
			return false, state, err
		}

		if len(escaped) > 0 {
			// Skip the escaped characters.
			if _, discardErr := s.reader.Discard(len(escaped)); discardErr != nil {
				return false, state, fmt.Errorf("parsing string: %w", discardErr)
			}

			// Write the escaped characters in case this is a comment.
			_, err = commentTxt.WriteString(string(escaped))
			if err != nil {
				return false, state, fmt.Errorf("writing runes %q: %w", escaped, err)
			}

			continue
		}

		// Look for the end of the string.
		stringEnd, err := s.peekEqual(s.config.Strings[state.sIndex].End)
		if err != nil {
			return false, state, fmt.Errorf("parsing string: %w", err)
		}

		if stringEnd {
			if _, discardErr := s.reader.Discard(len(s.config.Strings[state.sIndex].End)); discardErr != nil {
				return false, state, fmt.Errorf("parsing string: %w", discardErr)
			}

			return false, &stateCode{}, nil
		}

		rn, err := s.nextRune()
		if err != nil {
			return false, state, fmt.Errorf("parsing string: %w", err)
		}

		_, err = commentTxt.WriteRune(rn)
		if err != nil {
			return false, state, fmt.Errorf("writing rune %q: %w", rn, err)
		}
	}
}

// processMultilineComment processes multi-line comments and returns the next state.
//
//nolint:ireturn,nolintlint // returning interface required for state machine.
func (s *CommentScanner) processMultilineComment(state *stateMultilineComment) (state, error) {
	mlConfig := s.config.MultilineComments[state.index]

	// Discard the opening since we don't want to parse it. It could be the same as the closing.
	if _, errDiscard := s.reader.Discard(len(mlConfig.Start)); errDiscard != nil {
		return state, fmt.Errorf("parsing code: %w", errDiscard)
	}

	var commentTxt strings.Builder

	var nestingDepth int

	// Add the opening to the builder since we want it in the output.
	commentTxt.WriteString(string(mlConfig.Start))

	for {
		if mlConfig.Nested {
			// Look for a nested comment start.
			mlStart, err := s.peekEqual(mlConfig.Start)
			if err != nil {
				return state, err
			}

			if mlStart && (!mlConfig.AtFirstColumn || s.atFirstColumn) {
				if _, errDiscard := s.reader.Discard(len(mlConfig.Start)); errDiscard != nil {
					return state, fmt.Errorf("parsing multi-line comment: %w", errDiscard)
				}
				// Add the start to the builder.
				commentTxt.WriteString(string(mlConfig.Start))

				nestingDepth++

				continue
			}
		}

		// Look for the end of the comment.
		mlEnd, err := s.peekEqual(mlConfig.End)
		if err != nil {
			return state, err
		}

		if mlEnd && (!mlConfig.AtFirstColumn || s.atFirstColumn) {
			if _, errDiscard := s.reader.Discard(len(mlConfig.End)); errDiscard != nil {
				return state, fmt.Errorf("parsing multi-line comment: %w", errDiscard)
			}
			// Add the ending to the builder.
			commentTxt.WriteString(string(mlConfig.End))

			if nestingDepth == 0 {
				s.next = &Comment{
					Text:            commentTxt.String(),
					Line:            state.line,
					Multiline:       true,
					MultilineConfig: &s.config.MultilineComments[state.index],
				}

				return &stateCode{}, nil
			}

			if mlConfig.Nested {
				nestingDepth--
			}
		}

		rn, err := s.nextRune()
		if err != nil {
			return state, err
		}

		_, err = commentTxt.WriteRune(rn)
		if err != nil {
			return state, fmt.Errorf("writing rune %q: %w", rn, err)
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
		s.atFirstColumn = true
		s.atLineStart = true
	} else {
		s.atFirstColumn = false
		if !unicode.IsSpace(rn) {
			s.atLineStart = false
		}
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

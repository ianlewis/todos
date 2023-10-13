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
	// embed is required to be imported to use go:embed.
	_ "embed"

	"gopkg.in/yaml.v1"
)

//go:embed languages.yml
var languagesRaw []byte

// MultilineCommentConfig describes multi-line comments.
type MultilineCommentConfig struct {
	Start       string `yaml:"start,omitempty"`
	End         string `yaml:"end,omitempty"`
	AtLineStart bool   `yaml:"at_line_start,omitempty"`
}

var (
	// BackslashEscape indicates characters can be escaped by a backslash.
	BackslashEscape = "backslash"
	// NoEscape indicates characters cannot be escaped.
	NoEscape = "none"
	// DoubleEscape indicates that strings can be escaped with double characters.
	DoubleEscape = "double"
)

// StringConfig is config describing types of string.
type StringConfig struct {
	Start  string `yaml:"start,omitempty"`
	End    string `yaml:"end,omitempty"`
	Escape string `yaml:"escape,omitempty"`
}

// Config is configuration for a language comment scanner.
type Config struct {
	LineCommentStart []string               `yaml:"line_comment_start,omitempty"`
	MultilineComment MultilineCommentConfig `yaml:"multiline_comment,omitempty"`
	Strings          []StringConfig         `yaml:"strings,omitempty"`
}

type escapeFunc func(s *CommentScanner, st *stateString) (bool, error)

var escapeFuncs = map[string]escapeFunc{
	BackslashEscape: backslashEscape,
	NoEscape:        noEscape,
	DoubleEscape:    doubleEscape,
}

func noEscape(_ *CommentScanner, _ *stateString) (bool, error) {
	return false, nil
}

func backslashEscape(s *CommentScanner, st *stateString) (bool, error) {
	return s.peekEqual(append([]rune{'\\'}, s.config.Strings[st.index].End...))
}

func doubleEscape(s *CommentScanner, st *stateString) (bool, error) {
	b := append([]rune{}, s.config.Strings[st.index].End...)
	b = append(b, s.config.Strings[st.index].End...)
	return s.peekEqual(b)
}

// LanguagesConfig is a map of language names to their configuration. Keys are
// language names defined in the linguist library.
var LanguagesConfig map[string]*Config

//nolint:gochecknoinits // init needed to load embedded config.
func init() {
	// TODO(#460): Generate Go code rather than loading YAML at runtime.
	if err := yaml.Unmarshal(languagesRaw, &LanguagesConfig); err != nil {
		// NOTE: This shouldn't happen.
		panic(err)
	}
}

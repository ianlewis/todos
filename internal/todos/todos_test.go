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
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ianlewis/todos/internal/scanner"
)

type testScanner struct {
	index    int
	err      error
	comments []*scanner.Comment
}

func (s *testScanner) Config() *scanner.Config {
	return &scanner.Config{
		LineComments: []scanner.LineCommentConfig{
			{
				Start: []rune("//"),
			},
		},
		MultilineComments: []scanner.MultilineCommentConfig{
			{
				Start: []rune("/*"),
			},
		},
	}
}

func (s *testScanner) Scan() bool {
	if s.err != nil {
		return false
	}

	s.index++

	return s.index <= len(s.comments)
}

func (s *testScanner) Next() *scanner.Comment {
	if s.err != nil {
		return nil
	}

	return s.comments[s.index-1]
}

func (s *testScanner) Err() error {
	return s.err
}

func TestTODOScanner(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		s        CommentScanner
		config   *Config
		expected []*TODO
		errCheck func(error)
	}{
		"default_types.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// TODO",
						Line: 5,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text:      "/* TODO */",
						Line:      6,
						Multiline: true,
						MultilineConfig: &scanner.MultilineCommentConfig{
							Start: []rune("/*"),
							End:   []rune("*/"),
						},
					},
					{
						Text: "// godoc",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// FIXME",
						Line: 10,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "// TODO",
					Line:        5,
					CommentLine: 5,
				},
				{
					Type:        "TODO",
					Text:        "/* TODO */",
					Line:        6,
					CommentLine: 6,
				},
				{
					Type:        "FIXME",
					Text:        "// FIXME",
					Line:        10,
					CommentLine: 10,
				},
			},
		},

		"line_comments_basic.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// TODO",
						Line: 5,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// godoc ",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "// TODO",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
		"line_comments_message.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// TODO: foo",
						Line: 5,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// godoc ",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "// TODO: foo",
					Message:     "foo",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
		"line_comments_bug.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// TODO(github.com/foo/bar/issues/1)",
						Line: 5,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// godoc ",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "// TODO(github.com/foo/bar/issues/1)",
					Label:       "github.com/foo/bar/issues/1",
					Line:        5,
					CommentLine: 5,
				},
			},
		},

		"line_comments_bug_message.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// TODO(github.com/foo/bar/issues/1): foo",
						Line: 5,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// godoc ",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "// TODO(github.com/foo/bar/issues/1): foo",
					Message:     "foo",
					Label:       "github.com/foo/bar/issues/1",
					Line:        5,
					CommentLine: 5,
				},
			},
		},

		"line_comments_message_dash.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// TODO - foo",
						Line: 5,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// godoc ",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "// TODO - foo",
					Message:     "foo",
					Label:       "",
					Line:        5,
					CommentLine: 5,
				},
			},
		},

		"line_comments_label_message_dash.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// TODO(github.com/foo/bar/issues/1) - foo",
						Line: 5,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// godoc ",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "// TODO(github.com/foo/bar/issues/1) - foo",
					Message:     "foo",
					Label:       "github.com/foo/bar/issues/1",
					Line:        5,
					CommentLine: 5,
				},
			},
		},

		"line_comments_label_message_slash.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// TODO(github.com/foo/bar/issues/1) / foo",
						Line: 5,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// godoc ",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "// TODO(github.com/foo/bar/issues/1) / foo",
					Message:     "foo",
					Label:       "github.com/foo/bar/issues/1",
					Line:        5,
					CommentLine: 5,
				},
			},
		},

		"line_comments_label_message_multislash.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// TODO(github.com/foo/bar/issues/1) // foo",
						Line: 5,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// godoc ",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "// TODO(github.com/foo/bar/issues/1) // foo",
					Message:     "foo",
					Label:       "github.com/foo/bar/issues/1",
					Line:        5,
					CommentLine: 5,
				},
			},
		},

		"line_comments_label_message_nodelim.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// TODO(github.com/foo/bar/issues/1) foo",
						Line: 5,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// godoc ",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "// TODO(github.com/foo/bar/issues/1) foo",
					Message:     "foo",
					Label:       "github.com/foo/bar/issues/1",
					Line:        5,
					CommentLine: 5,
				},
			},
		},

		"multiline_comments_basic.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text:      "/*\nfoo\nTODO\n*/",
						Line:      5,
						Multiline: true,
						MultilineConfig: &scanner.MultilineCommentConfig{
							Start: []rune("/*"),
							End:   []rune("*/"),
						},
					},
					{
						Text: "// godoc ",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "TODO",
					Line:        7,
					CommentLine: 5,
				},
			},
		},
		"multiline_comments_single_line_message.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text:      "/* TODO: foo */",
						Line:      5,
						Multiline: true,
						MultilineConfig: &scanner.MultilineCommentConfig{
							Start: []rune("/*"),
							End:   []rune("*/"),
						},
					},
					{
						Text: "// godoc ",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "/* TODO: foo */",
					Message:     "foo",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
		"multiline_comments_message.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text:      "/*\nfoo\nTODO: foo\n*/",
						Line:      5,
						Multiline: true,
						MultilineConfig: &scanner.MultilineCommentConfig{
							Start: []rune("/*"),
							End:   []rune("*/"),
						},
					},
					{
						Text: "// godoc ",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "TODO: foo",
					Message:     "foo",
					Line:        7,
					CommentLine: 5,
				},
			},
		},
		"multiline_comments_bug.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text:      "/*\nfoo\nTODO(github.com/foo/bar/issues1)\n*/",
						Line:      5,
						Multiline: true,
						MultilineConfig: &scanner.MultilineCommentConfig{
							Start: []rune("/*"),
							End:   []rune("*/"),
						},
					},
					{
						Text: "// godoc ",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "TODO(github.com/foo/bar/issues1)",
					Label:       "github.com/foo/bar/issues1",
					Line:        7,
					CommentLine: 5,
				},
			},
		},
		"multiline_comments_bug_message.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text:      "/*\nfoo\nTODO(github.com/foo/bar/issues1): foo\n*/",
						Line:      5,
						Multiline: true,
						MultilineConfig: &scanner.MultilineCommentConfig{
							Start: []rune("/*"),
							End:   []rune("*/"),
						},
					},
					{
						Text: "// godoc ",
						Line: 9,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "TODO(github.com/foo/bar/issues1): foo",
					Label:       "github.com/foo/bar/issues1",
					Message:     "foo",
					Line:        7,
					CommentLine: 5,
				},
			},
		},
		"multiline_comments_no_todo.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text:      "/*\nfoo\nbar\n*/",
						Line:      5,
						Multiline: true,
						MultilineConfig: &scanner.MultilineCommentConfig{
							Start: []rune("/*"),
							End:   []rune("*/"),
						},
					},
					{
						Text: "// godoc ",
						Line: 9,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: nil,
		},

		"multiline_comments_multiple_todos.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// TODO: before",
						Line: 2,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text:      "/*\nfoo\nTODO(github.com/foo/bar/issues1): foo\nTODO: second task\n */",
						Line:      5,
						Multiline: true,
						MultilineConfig: &scanner.MultilineCommentConfig{
							Start: []rune("/*"),
							End:   []rune("*/"),
						},
					},
					{
						Text: "// godoc ",
						Line: 9,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text: "// TODO: after",
						Line: 10,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "// TODO: before",
					Label:       "",
					Message:     "before",
					Line:        2,
					CommentLine: 2,
				},
				{
					Type:        "TODO",
					Text:        "TODO(github.com/foo/bar/issues1): foo",
					Label:       "github.com/foo/bar/issues1",
					Message:     "foo",
					Line:        7,
					CommentLine: 5,
				},
				{
					Type:        "TODO",
					Text:        "TODO: second task",
					Label:       "",
					Message:     "second task",
					Line:        8,
					CommentLine: 5,
				},
				{
					Type:        "TODO",
					Text:        "// TODO: after",
					Label:       "",
					Message:     "after",
					Line:        10,
					CommentLine: 10,
				},
			},
		},
		"multiline_comments_javadoc.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text:      "/**\n * foo\n * TODO(github.com/foo/bar/issues1): foo\n */",
						Line:      5,
						Multiline: true,

						MultilineConfig: &scanner.MultilineCommentConfig{
							Start: []rune("/*"),
							End:   []rune("*/"),
						},
					},
					{
						Text: "// javadoc ",
						Line: 9,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "* TODO(github.com/foo/bar/issues1): foo",
					Label:       "github.com/foo/bar/issues1",
					Message:     "foo",
					Line:        7,
					CommentLine: 5,
				},
			},
		},
		"line_comment_leading_text.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment, TODO",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: nil,
		},
		"special_case_todo_naked.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "//TODO",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "//TODO",
					Label:       "",
					Message:     "",
					Line:        1,
					CommentLine: 1,
				},
			},
		},
		"special_case_todo_with_message.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "//TODO Add some useful code here.",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "//TODO Add some useful code here.",
					Label:       "",
					Message:     "Add some useful code here.",
					Line:        1,
					CommentLine: 1,
				},
			},
		},
		"special_case_todo_with_label.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "//TODO(github.com/foo/bar/issues/1) Add some useful code here.",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "//TODO(github.com/foo/bar/issues/1) Add some useful code here.",
					Label:       "github.com/foo/bar/issues/1",
					Message:     "Add some useful code here.",
					Line:        1,
					CommentLine: 1,
				},
			},
		},
		"special_case_todo_no_message.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "//TODO(github.com/foo/bar/issues/1)",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "//TODO(github.com/foo/bar/issues/1)",
					Label:       "github.com/foo/bar/issues/1",
					Message:     "",
					Line:        1,
					CommentLine: 1,
				},
			},
		},
		"multiline_comments_leading_whitespace.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
					{
						Text:      "/*\nfoo\n\t\t\tTODO(github.com/foo/bar/issues1): foo\n*/",
						Line:      5,
						Multiline: true,
						MultilineConfig: &scanner.MultilineCommentConfig{
							Start: []rune("/*"),
							End:   []rune("*/"),
						},
					},
					{
						Text: "// godoc ",
						Line: 7,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "TODO(github.com/foo/bar/issues1): foo",
					Label:       "github.com/foo/bar/issues1",
					Message:     "foo",
					Line:        7,
					CommentLine: 5,
				},
			},
		},
		"extra_line_comment_prefix.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "//// TODO: comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "//// TODO: comment",
					Label:       "",
					Message:     "comment",
					Line:        1,
					CommentLine: 1,
				},
			},
		},
		"at_prefix.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// @TODO: comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "// @TODO: comment",
					Label:       "",
					Message:     "comment",
					Line:        1,
					CommentLine: 1,
				},
			},
		},
		"at_prefix_alt.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "//@TODO comment",
						Line: 1,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "//@TODO comment",
					Label:       "",
					Message:     "comment",
					Line:        1,
					CommentLine: 1,
				},
			},
		},
		"at_prefix_multiline.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text:      "/**\n * @TODO: comment\n */",
						Line:      1,
						Multiline: true,
						MultilineConfig: &scanner.MultilineCommentConfig{
							Start: []rune("/*"),
							End:   []rune("*/"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "* @TODO: comment",
					Label:       "",
					Message:     "comment",
					Line:        2,
					CommentLine: 1,
				},
			},
		},
		// Regression test for issue #1520
		// Ensure that the TODOScanner continues scanning after finding a
		// multi-line comment with no TODOs in it.
		"regression_1520.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text:      "/**\n no match here\n */",
						Line:      1,
						Multiline: true,
						MultilineConfig: &scanner.MultilineCommentConfig{
							Start: []rune("/*"),
							End:   []rune("*/"),
						},
					},
					{
						Text: "// TODO: match",
						Line: 5,
						LineConfig: &scanner.LineCommentConfig{
							Start: []rune("//"),
						},
					},
				},
			},
			config: &Config{
				Types: []string{"TODO"},
			},
			expected: []*TODO{
				{
					Type:        "TODO",
					Text:        "// TODO: match",
					Label:       "",
					Message:     "match",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := NewTODOScanner(tc.s, tc.config)

			var found []*TODO

			for s.Scan() {
				found = append(found, s.Next())
			}

			got, want := found, tc.expected
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("unexpected todos (-want +got):\n%s", diff)
			}

			err := s.Err()
			if tc.errCheck == nil && err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if tc.errCheck != nil {
				tc.errCheck(err)
			}
		})
	}
}

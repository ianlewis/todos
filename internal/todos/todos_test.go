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
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ianlewis/todos/internal/scanner"
)

type testScanner struct {
	index    int
	err      error
	comments []*scanner.Comment
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

type todoSlice []*TODO

func (s todoSlice) String() string {
	if s == nil {
		return "nil"
	}
	var items []string
	for _, t := range s {
		items = append(items, fmt.Sprintf("%#v", t))
	}
	return "[" + strings.Join(items, ", ") + "]"
}

func TestTODOScanner(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		s        CommentScanner
		config   *Config
		expected []*TODO
		errCheck func(error)
	}{
		"line_comments.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
					},
					{
						Text: "// TODO: foo",
						Line: 5,
					},
					{
						Text: "// godoc ",
						Line: 7,
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
					Line:        5,
					CommentLine: 5,
				},
			},
		},
		"multi_line_comments.go": {
			s: &testScanner{
				comments: []*scanner.Comment{
					{
						Text: "// package comment",
						Line: 1,
					},
					{
						Text: "/*\nfoo\nTODO: foo\n*/",
						Line: 5,
					},
					{
						Text: "// godoc ",
						Line: 7,
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
					Line:        7,
					CommentLine: 5,
				},
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := NewTODOScanner(tc.s, tc.config)
			var found []*TODO
			for s.Scan() {
				found = append(found, s.Next())
			}

			if got, want := found, tc.expected; !reflect.DeepEqual(got, want) {
				t.Errorf("unexpected todos, got: %s, want: %s", todoSlice(got), todoSlice(want))
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

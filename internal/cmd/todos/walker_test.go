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

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ianlewis/todos/internal/testutils"
	"github.com/ianlewis/todos/internal/todos"
)

func newFixture(files map[string]string, types, paths []string) *fixture {
	f := &fixture{types: types, paths: paths}
	f.wd = testutils.Must(os.Getwd())

	f.dir = testutils.Must(os.MkdirTemp("", "code"))
	cleanup, cancel := testutils.WithCancel(func() {
		f.cleanup()
	}, nil)
	defer cleanup()
	testutils.Check(os.Chdir(f.dir))

	for path, src := range files {
		fullPath := filepath.Join(f.dir, path)
		testutils.Check(os.MkdirAll(filepath.Dir(fullPath), 0700))
		testutils.Check(os.WriteFile(fullPath, []byte(src), 0600))
	}

	cancel()
	return f
}

type fixture struct {
	wd    string
	dir   string
	types []string
	paths []string
	out   []todoOpt
	err   []error
}

func (f *fixture) walker() *TODOWalker {
	paths := []string{"."}
	if len(f.paths) > 0 {
		paths = f.paths
	}

	return &TODOWalker{
		outFunc: f.outFunc,
		errFunc: f.errFunc,
		todoConfig: &todos.Config{
			Types: f.types,
		},
		paths: paths,
	}
}

func (f *fixture) outFunc(o todoOpt) {
	f.out = append(f.out, o)
}

func (f *fixture) errFunc(err error) {
	f.err = append(f.err, err)
}

func (f *fixture) cleanup() {
	testutils.Check(os.Chdir(f.wd))
	testutils.Check(os.RemoveAll(f.dir))
}

var testCases = []struct {
	name     string
	files    map[string]string
	types    []string
	paths    []string
	expected []todoOpt
	err      bool
}{
	{
		name: "single file traverse path",
		files: map[string]string{
			"line_comments.go": `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		expected: []todoOpt{
			{
				fileName: "line_comments.go",
				todo: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
	{
		name: "multi-file traverse path",
		files: map[string]string{
			"line_comments.go": `
				package foo

				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
			"multi_line.go": `
				package foo

				/*
				This is a comment.
				TODO: Some other task.
				*/
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		expected: []todoOpt{
			{
				fileName: "line_comments.go",
				todo: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        7,
					CommentLine: 7,
				},
			},
			{
				fileName: "multi_line.go",
				todo: &todos.TODO{
					Type: "TODO",
					// TODO: leading whitespace should be stripped.
					Text:        "\t\t\t\tTODO: Some other task.",
					Line:        6,
					CommentLine: 4,
				},
			},
		},
	},
	{
		name: "multi-file sub-directory",
		files: map[string]string{
			"line_comments.go": `
				package foo

				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
			"sub-dir/multi_line.go": `
				package foo

				/*
				This is a comment.
				TODO: Some other task.
				*/
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		expected: []todoOpt{
			{
				fileName: "line_comments.go",
				todo: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        7,
					CommentLine: 7,
				},
			},
			{
				fileName: "sub-dir/multi_line.go",
				todo: &todos.TODO{
					Type: "TODO",
					// TODO: leading whitespace should be stripped.
					Text:        "\t\t\t\tTODO: Some other task.",
					Line:        6,
					CommentLine: 4,
				},
			},
		},
	},

	{
		name: "single file",
		files: map[string]string{
			"not_scanned.go": `package foo
				// TODO: not read.
				func NotRead() {
					return
				}`,

			"line_comments.go": `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		paths: []string{"line_comments.go"},
		expected: []todoOpt{
			{
				fileName: "line_comments.go",
				todo: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
}

func TestTODOWalker(t *testing.T) {
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(tc.files, tc.types, tc.paths)
			defer f.cleanup()

			walker := f.walker()
			if got, want := walker.Walk(), tc.err; got != want {
				t.Errorf("unexpected error code, got: %v, want: %v", got, want)
			}

			got, want := f.out, tc.expected
			if diff := cmp.Diff(want, got, cmp.AllowUnexported(todoOpt{})); diff != "" {
				t.Errorf("unexpected output (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTODOWalker_PathNotExists(t *testing.T) {
	notExistsPath := "/does/not/exist"
	if _, err := os.Stat(notExistsPath); !errors.Is(err, os.ErrNotExist) {
		panic(fmt.Sprintf("assertion failure: %s exists.", notExistsPath))
	}

	files := map[string]string{
		"line_comments.go": `package foo
			// package comment

			// TODO is a function.
			// TODO: some task.
			func TODO() {
				return // Random comment
			}`,
	}

	f := newFixture(files, []string{"TODO"}, nil)
	defer f.cleanup()

	walker := f.walker()
	walker.paths = append(walker.paths, notExistsPath)
	if got, want := walker.Walk(), true; got != want {
		t.Errorf("unexpected error code, got: %v, want: %v", got, want)
	}

	if got, want := len(f.err), 1; got != want {
		t.Errorf("unexpected # of errors, got: %v, want: %v\n%v", got, want, f.err)
	}
	if got, want := f.err[0], os.ErrNotExist; !errors.Is(got, os.ErrNotExist) {
		t.Errorf("unexpected error, got: %v, want: %v", got, want)
	}

	got, want := f.out, []todoOpt{
		{
			fileName: "line_comments.go",
			todo: &todos.TODO{
				Type:        "TODO",
				Text:        "// TODO: some task.",
				Line:        5,
				CommentLine: 5,
			},
		},
	}
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(todoOpt{})); diff != "" {
		t.Errorf("unexpected output (-want +got):\n%s", diff)
	}
}

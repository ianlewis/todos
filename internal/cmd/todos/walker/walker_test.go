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

package walker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ianlewis/todos/internal/cmd/todos/options"
	"github.com/ianlewis/todos/internal/testutils"
	"github.com/ianlewis/todos/internal/todos"
)

func newFixture(files map[string]string, types, paths []string, hidden, vendored, docs bool) *fixture {
	f := &fixture{}
	f.wd = testutils.Must(os.Getwd())

	f.dir = testutils.Must(os.MkdirTemp("", "code"))
	cleanup, cancel := testutils.WithCancel(func() {
		f.cleanup()
	}, nil)
	defer cleanup()
	testutils.Check(os.Chdir(f.dir))

	for path, src := range files {
		fullPath := filepath.Join(f.dir, path)
		testutils.Check(os.MkdirAll(filepath.Dir(fullPath), 0o700))
		testutils.Check(os.WriteFile(fullPath, []byte(src), 0o600))
	}

	if len(paths) == 0 {
		paths = []string{"."}
	}

	f.walker = New(&options.Options{
		Output:    f.outFunc,
		Error:     f.errFunc,
		TODOTypes: types,
		Paths:     paths,

		IncludeHidden:   hidden,
		IncludeVendored: vendored,
		IncludeDocs:     docs,
	})

	cancel()
	return f
}

type fixture struct {
	walker *TODOWalker
	wd     string
	dir    string
	out    []options.TODOOpt
	err    []error
}

func (f *fixture) outFunc(o options.TODOOpt) {
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
	name  string
	files map[string]string
	types []string
	paths []string

	hidden   bool
	vendored bool
	docs     bool

	expected []options.TODOOpt
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
		expected: []options.TODOOpt{
			{
				FileName: "line_comments.go",
				TODO: &todos.TODO{
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
		expected: []options.TODOOpt{
			{
				FileName: "line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        7,
					CommentLine: 7,
				},
			},
			{
				FileName: "multi_line.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "TODO: Some other task.",
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
			filepath.Join("sub-dir", "multi_line.go"): `
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
		expected: []options.TODOOpt{
			{
				FileName: "line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        7,
					CommentLine: 7,
				},
			},
			{
				FileName: filepath.Join("sub-dir", "multi_line.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "TODO: Some other task.",
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
		expected: []options.TODOOpt{
			{
				FileName: "line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
	{
		name: "hidden file skipped",
		files: map[string]string{
			// NOTE: Files starting with '.' should be hidden on all platforms.
			".line_comments.go": `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types:    []string{"TODO"},
		expected: nil,
	},
	{
		name: "hidden file specified",
		files: map[string]string{
			// NOTE: Files starting with '.' should be hidden on all platforms.
			".line_comments.go": `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		paths: []string{".line_comments.go"},
		expected: []options.TODOOpt{
			{
				FileName: ".line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
	{
		name: "hidden file processed",
		files: map[string]string{
			// NOTE: Files starting with '.' should be hidden on all platforms.
			".line_comments.go": `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		// NOTE: Include hidden files.
		hidden: true,
		expected: []options.TODOOpt{
			{
				FileName: ".line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
	{
		name: "vendored file skipped",
		files: map[string]string{
			filepath.Join("vendor", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types:    []string{"TODO"},
		expected: nil,
	},
	{
		name: "vendored file specified",
		files: map[string]string{
			filepath.Join("vendor", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		paths: []string{filepath.Join("vendor", "line_comments.go")},
		expected: []options.TODOOpt{
			{
				FileName: filepath.Join("vendor", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
	{
		name: "vendored file processed",
		files: map[string]string{
			filepath.Join("vendor", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		// NOTE: Include vendored files.
		vendored: true,
		expected: []options.TODOOpt{
			{
				FileName: filepath.Join("vendor", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
	{
		name: "docs file skipped",
		files: map[string]string{
			filepath.Join("docs", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types:    []string{"TODO"},
		expected: nil,
	},
	{
		name: "docs file specified",
		files: map[string]string{
			filepath.Join("docs", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		paths: []string{filepath.Join("docs", "line_comments.go")},
		expected: []options.TODOOpt{
			{
				FileName: filepath.Join("docs", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
	{
		name: "docs file processed",
		files: map[string]string{
			filepath.Join("docs", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		// NOTE: Include docs files.
		docs: true,
		expected: []options.TODOOpt{
			{
				FileName: filepath.Join("docs", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
	{
		name: "hidden dir skipped",
		files: map[string]string{
			// NOTE: Files starting with '.' should be hidden on all platforms.
			filepath.Join(".somepath", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types:    []string{"TODO"},
		expected: nil,
	},
	{
		name: "hidden dir specified",
		files: map[string]string{
			// NOTE: Files starting with '.' should be hidden on all platforms.
			filepath.Join(".somepath", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		paths: []string{filepath.Join(".somepath")},
		expected: []options.TODOOpt{
			{
				FileName: filepath.Join(".somepath", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
	{
		name: "hidden dir processed",
		files: map[string]string{
			// NOTE: Files starting with '.' should be hidden on all platforms.
			filepath.Join(".somepath", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		// NOTE: Include hidden files.
		hidden: true,
		expected: []options.TODOOpt{
			{
				FileName: filepath.Join(".somepath", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
	{
		name: "vendored dir skipped",
		files: map[string]string{
			filepath.Join("vendor", "pkgname", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types:    []string{"TODO"},
		expected: nil,
	},
	{
		name: "vendored dir specified",
		files: map[string]string{
			filepath.Join("vendor", "pkgname", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		paths: []string{filepath.Join("vendor", "pkgname")},
		expected: []options.TODOOpt{
			{
				FileName: filepath.Join("vendor", "pkgname", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
	{
		name: "vendored dir processed",
		files: map[string]string{
			filepath.Join("vendor", "pkgname", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		// NOTE: Include vendored files.
		vendored: true,
		expected: []options.TODOOpt{
			{
				FileName: filepath.Join("vendor", "pkgname", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
	{
		name: "docs dir skipped",
		files: map[string]string{
			filepath.Join("docs", "somedir", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types:    []string{"TODO"},
		expected: nil,
	},
	{
		name: "docs dir specified",
		files: map[string]string{
			filepath.Join("docs", "somedir", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		paths: []string{filepath.Join("docs", "somedir")},
		expected: []options.TODOOpt{
			{
				FileName: filepath.Join("docs", "somedir", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
	{
		name: "docs dir processed",
		files: map[string]string{
			filepath.Join("docs", "somedir", "line_comments.go"): `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		// NOTE: Include docs files.
		docs: true,
		expected: []options.TODOOpt{
			{
				FileName: filepath.Join("docs", "somedir", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
}

//nolint:paralleltest // fixture uses Chdir and cannot be run in parallel.
func TestTODOWalker(t *testing.T) {
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(tc.files, tc.types, tc.paths, tc.hidden, tc.vendored, tc.docs)
			defer f.cleanup()

			if got, want := f.walker.Walk(), tc.err; got != want {
				t.Errorf("unexpected error code, got: %v, want: %v", got, want)
			}

			got, want := f.out, tc.expected
			if diff := cmp.Diff(want, got, cmp.AllowUnexported(options.TODOOpt{})); diff != "" {
				t.Errorf("unexpected output (-want +got):\n%s", diff)
			}
		})
	}
}

//nolint:paralleltest // fixture uses Chdir and cannot be run in parallel.
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

	f := newFixture(files, []string{"TODO"}, nil, false, false, false)
	defer f.cleanup()

	walker := f.walker
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

	got, want := f.out, []options.TODOOpt{
		{
			FileName: "line_comments.go",
			TODO: &todos.TODO{
				Type:        "TODO",
				Text:        "// TODO: some task.",
				Line:        5,
				CommentLine: 5,
			},
		},
	}
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(options.TODOOpt{})); diff != "" {
		t.Errorf("unexpected output (-want +got):\n%s", diff)
	}
}

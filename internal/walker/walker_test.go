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
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ianlewis/todos/internal/todos"
)

var testCases = []struct {
	name  string
	files map[string]string
	types []string
	paths []string

	hidden   bool
	vendored bool

	expected []*TODORef
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
		expected: []*TODORef{
			{
				FileName: "line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Message:     "some task.",
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
		expected: []*TODORef{
			{
				FileName: "line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Message:     "some task.",
					Line:        7,
					CommentLine: 7,
				},
			},
			{
				FileName: "multi_line.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "TODO: Some other task.",
					Message:     "Some other task.",
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
		expected: []*TODORef{
			{
				FileName: "line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Message:     "some task.",
					Line:        7,
					CommentLine: 7,
				},
			},
			{
				FileName: filepath.Join("sub-dir", "multi_line.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "TODO: Some other task.",
					Message:     "Some other task.",
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
		expected: []*TODORef{
			{
				FileName: "line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Message:     "some task.",
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
		expected: []*TODORef{
			{
				FileName: ".line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Message:     "some task.",
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
		expected: []*TODORef{
			{
				FileName: ".line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Message:     "some task.",
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
		expected: []*TODORef{
			{
				FileName: filepath.Join("vendor", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Message:     "some task.",
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
		expected: []*TODORef{
			{
				FileName: filepath.Join("vendor", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Message:     "some task.",
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
		expected: []*TODORef{
			{
				FileName: filepath.Join(".somepath", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Message:     "some task.",
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
		expected: []*TODORef{
			{
				FileName: filepath.Join(".somepath", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Message:     "some task.",
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
		expected: []*TODORef{
			{
				FileName: filepath.Join("vendor", "pkgname", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Message:     "some task.",
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
		expected: []*TODORef{
			{
				FileName: filepath.Join("vendor", "pkgname", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Message:     "some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
		},
	},
	{
		name: "single file traverse path multiple todos",
		files: map[string]string{
			"line_comments.go": `package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					// TODO: some other task.
					return // Random comment
				}`,
		},
		types: []string{"TODO"},
		expected: []*TODORef{
			{
				FileName: "line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Message:     "some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
			{
				FileName: "line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some other task.",
					Message:     "some other task.",
					Line:        7,
					CommentLine: 7,
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
			f := newFixture(tc.files, tc.types, tc.paths, tc.hidden, tc.vendored)
			defer f.cleanup()

			if got, want := f.walker.Walk(), tc.err; got != want {
				t.Errorf("unexpected error code, got: %v, want: %v", got, want)
			}

			got, want := f.out, tc.expected
			if diff := cmp.Diff(want, got, cmp.AllowUnexported(TODORef{})); diff != "" {
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

	f := newFixture(files, []string{"TODO"}, nil, false, false)
	defer f.cleanup()

	walker := f.walker
	walker.options.Paths = append(walker.options.Paths, notExistsPath)
	if got, want := walker.Walk(), true; got != want {
		t.Errorf("unexpected error code, got: %v, want: %v", got, want)
	}

	if got, want := len(f.err), 1; got != want {
		t.Errorf("unexpected # of errors, got: %v, want: %v\n%v", got, want, f.err)
	}
	if got, want := f.err[0], os.ErrNotExist; !errors.Is(got, os.ErrNotExist) {
		t.Errorf("unexpected error, got: %v, want: %v", got, want)
	}

	got, want := f.out, []*TODORef{
		{
			FileName: "line_comments.go",
			TODO: &todos.TODO{
				Type:        "TODO",
				Text:        "// TODO: some task.",
				Message:     "some task.",
				Line:        5,
				CommentLine: 5,
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected output (-want +got):\n%s", diff)
	}
}

func TestTODOWalker_DefaultOptions(t *testing.T) {
	t.Parallel()

	walker := New(nil)
	got, want := walker.options.Config.Types, []string{"TODO"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected config (-want +got):\n%s", diff)
	}
}

//nolint:paralleltest // fixture uses Chdir and cannot be run in parallel.
func TestTODOWalker_StopEarly(t *testing.T) {
	files := map[string]string{
		"line_comments.go": `package foo
			// package comment

			// TODO is a function.
			// TODO: some task.
			func TODO() {
				// TODO: some other task.
				return // Random comment
			}`,
	}

	f := newFixture(files, []string{"TODO"}, nil, false, false)
	defer f.cleanup()

	// Override the handler to cause it to stop early.
	f.walker.options.TODOFunc = func(r *TODORef) error {
		if err := f.outFunc(r); err != nil {
			return err
		}
		return fs.SkipAll
	}

	if got, want := f.walker.Walk(), false; got != want {
		t.Errorf("unexpected error code, got: %v, want: %v", got, want)
	}

	// NOTE: there are two TODOs in the file but we only get one because we
	// stopped early.
	got, want := f.out, []*TODORef{
		{
			FileName: "line_comments.go",
			TODO: &todos.TODO{
				Type:        "TODO",
				Text:        "// TODO: some task.",
				Message:     "some task.",
				Line:        5,
				CommentLine: 5,
			},
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected output (-want +got):\n%s", diff)
	}
}

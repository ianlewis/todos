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

	"github.com/gobwas/glob"
	"github.com/google/go-cmp/cmp"

	"github.com/ianlewis/todos/internal/testutils"
	"github.com/ianlewis/todos/internal/todos"
)

var testCases = []struct {
	name  string
	files []*testutils.File
	opts  *Options

	expected []*TODORef
	err      bool
}{
	{
		name: "single file traverse path",
		files: []*testutils.File{
			{
				Path: "line_comments.go",
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
		},
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
		files: []*testutils.File{
			{
				Path: "line_comments.go",
				Contents: []byte(`
				package foo

				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
			{
				Path: "multi_line.go",
				Contents: []byte(`
				package foo

				/*
				This is a comment.
				TODO: Some other task.
				*/
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
		},
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
		files: []*testutils.File{
			{
				Path: "line_comments.go",
				Contents: []byte(`
				package foo

				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
			{
				Path: filepath.Join("sub-dir", "multi_line.go"),
				Contents: []byte(`
				package foo

				/*
				This is a comment.
				TODO: Some other task.
				*/
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
		},
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
		files: []*testutils.File{
			{
				Path: "not_scanned.go",
				Contents: []byte(`package foo
				// TODO: not read.
				func NotRead() {
					return
				}`),
				Mode: 0o600,
			},
			{
				Path: "line_comments.go",
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
			Paths:   []string{"line_comments.go"},
		},
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
		files: []*testutils.File{
			{
				// NOTE: Files starting with '.' should be hidden on all platforms.
				Path: ".line_comments.go",
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
		},
		expected: nil,
	},
	{
		name: "hidden file specified",
		files: []*testutils.File{
			{
				// NOTE: Files starting with '.' should be hidden on all platforms.
				Path: ".line_comments.go",
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
			Paths:   []string{".line_comments.go"},
		},
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
		files: []*testutils.File{
			{
				// NOTE: Files starting with '.' should be hidden on all platforms.
				Path: ".line_comments.go",
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
			// NOTE: Include hidden files.
			IncludeHidden: true,
		},
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
		files: []*testutils.File{
			{
				Path: filepath.Join("vendor", "line_comments.go"),
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
		},
		expected: nil,
	},
	{
		name: "vendored file specified",
		files: []*testutils.File{
			{
				Path: filepath.Join("vendor", "line_comments.go"),
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
			Paths:   []string{filepath.Join("vendor", "line_comments.go")},
		},
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
		files: []*testutils.File{
			{
				Path: filepath.Join("vendor", "line_comments.go"),
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset:         "UTF-8",
			IncludeVendored: true,
			Paths:           []string{filepath.Join("vendor", "line_comments.go")},
		},
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
		files: []*testutils.File{
			{
				// NOTE: Files starting with '.' should be hidden on all platforms.
				Path: filepath.Join(".somepath", "line_comments.go"),
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
		},
		expected: nil,
	},
	{
		name: "hidden dir specified",
		files: []*testutils.File{
			{
				// NOTE: Files starting with '.' should be hidden on all platforms.
				Path: filepath.Join(".somepath", "line_comments.go"),
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
			Paths:   []string{filepath.Join(".somepath")},
		},
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
		files: []*testutils.File{
			{
				// NOTE: Files starting with '.' should be hidden on all platforms.
				Path: filepath.Join(".somepath", "line_comments.go"),
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
			// NOTE: Include hidden files.
			IncludeHidden: true,
		},
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
		files: []*testutils.File{
			{
				Path: filepath.Join("vendor", "pkgname", "line_comments.go"),
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
		},
		expected: nil,
	},
	{
		name: "vendored dir specified",
		files: []*testutils.File{
			{
				Path: filepath.Join("vendor", "pkgname", "line_comments.go"),
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
			Paths:   []string{filepath.Join("vendor", "pkgname")},
		},
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
		files: []*testutils.File{
			{
				Path: filepath.Join("vendor", "pkgname", "line_comments.go"),
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
			// NOTE: Include vendored files.
			IncludeVendored: true,
		},
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
		files: []*testutils.File{
			{
				Path: "line_comments.go",
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					// TODO: some other task.
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
		},
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
	{
		name: "exclude file",
		files: []*testutils.File{
			{
				Path: "line_comments.go",
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
			{
				Path: "excluded.go",
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset:      "UTF-8",
			ExcludeGlobs: []glob.Glob{glob.MustCompile("excluded.*")},
		},
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
		name: "exclude dir",
		files: []*testutils.File{
			{
				Path: filepath.Join("src", "line_comments.go"),
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
			{
				Path: filepath.Join("excluded", "more_line_comments.go"),
				Contents: []byte(`package foo
				// package comment

				// TODO is a function.
				// TODO: some task.
				func TODO() {
					return // Random comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset:         "UTF-8",
			ExcludeDirGlobs: []glob.Glob{glob.MustCompile("exclude?")},
		},
		expected: []*TODORef{
			{
				FileName: filepath.Join("src", "line_comments.go"),
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
}

type fixture struct {
	dir *testutils.TempDir
	wd  string
	out []*TODORef
	err []error
}

func (f *fixture) cleanup() {
	testutils.Check(os.Chdir(f.wd))
	f.dir.Cleanup()
}

func newFixture(files []*testutils.File, opts *Options) (*fixture, *TODOWalker) {
	dir := testutils.NewTempDir(files)
	cleanup, cancel := testutils.WithCancel(func() {
		dir.Cleanup()
	}, nil)
	defer cleanup()

	wd := testutils.Must(os.Getwd())

	testutils.Check(os.Chdir(dir.Dir()))

	if len(opts.Paths) == 0 {
		opts.Paths = []string{"."}
	}

	f := &fixture{
		dir: dir,
		wd:  wd,
	}

	todoFunc := opts.TODOFunc
	opts.TODOFunc = func(r *TODORef) error {
		f.out = append(f.out, r)
		if todoFunc != nil {
			return todoFunc(r)
		}
		return nil
	}

	errFunc := opts.ErrorFunc
	opts.ErrorFunc = func(err error) error {
		f.err = append(f.err, err)
		if errFunc != nil {
			return errFunc(err)
		}
		return nil
	}

	cancel()
	return f, New(opts)
}

//nolint:paralleltest // fixture uses Chdir and cannot be run in parallel.
func TestTODOWalker(t *testing.T) {
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			f, w := newFixture(tc.files, tc.opts)
			defer f.cleanup()

			if got, want := w.Walk(), tc.err; got != want {
				t.Errorf("unexpected error code, got: %v, want: %v\nw.err: %v", got, want, w.err)
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

	files := []*testutils.File{
		{
			Path: "line_comments.go",
			Contents: []byte(`package foo
			// package comment

			// TODO is a function.
			// TODO: some task.
			func TODO() {
				return // Random comment
			}`),
			Mode: 0o600,
		},
	}

	opts := &Options{
		Config: &todos.Config{
			Types: []string{"TODO"},
		},
		Charset: "UTF-8",
		Paths:   []string{"line_comments.go", notExistsPath},
	}

	f, w := newFixture(files, opts)
	defer f.cleanup()

	if got, want := w.Walk(), true; got != want {
		t.Errorf("unexpected error code, got: %v, want: %v\nw.err: %v", got, want, w.err)
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
	file := &testutils.File{
		Path: "line_comments.go",
		Contents: []byte(`package foo
			// package comment

			// TODO is a function.
			// TODO: some task.
			func TODO() {
				// TODO: some other task.
				return // Random comment
			}`),
		Mode: 0o600,
	}

	opts := &Options{
		Config: &todos.Config{
			Types: []string{"TODO"},
		},
		Charset: "UTF-8",
		// Override the handler to cause it to stop early.
		TODOFunc: func(r *TODORef) error {
			return fs.SkipAll
		},
	}

	f, w := newFixture([]*testutils.File{file}, opts)
	defer f.cleanup()

	if got, want := w.Walk(), false; got != want {
		t.Errorf("unexpected error code, got: %v, want: %v\nw.err: %v", got, want, w.err)
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

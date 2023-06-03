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
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ianlewis/todos/internal/testutils"
	"github.com/ianlewis/todos/internal/todos"
)

func newFixture(files map[string]string, types []string) *fixture {
	f := &fixture{types: types}
	f.wd = testutils.Must(os.Getwd())

	f.dir = testutils.Must(os.MkdirTemp("", "code"))
	cleanup, cancel := testutils.WithCancel(func() {
		f.cleanup()
	}, nil)
	defer cleanup()
	testutils.Check(os.Chdir(f.dir))

	for path, src := range files {
		fullPath := filepath.Join(f.dir, path)
		testutils.Check(os.MkdirAll(filepath.Dir(fullPath), 0600))
		testutils.Check(os.WriteFile(fullPath, []byte(src), 0600))
	}

	cancel()
	return f
}

type fixture struct {
	wd    string
	dir   string
	types []string
	out   []todoOpt
	err   []error
}

func (f *fixture) walker() *TODOWalker {
	return &TODOWalker{
		outFunc: f.outFunc,
		errFunc: f.errFunc,
		todoConfig: &todos.Config{
			Types: f.types,
		},
		paths: []string{"."},
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
	expected []todoOpt
	err      bool
}{
	{
		name: "line comments",
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
}

func TestWalker(t *testing.T) {
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(tc.files, tc.types)
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

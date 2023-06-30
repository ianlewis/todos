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
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/urfave/cli/v2"

	"github.com/ianlewis/todos/internal/testutils"
	"github.com/ianlewis/todos/internal/todos"
	"github.com/ianlewis/todos/internal/walker"
)

func newContext(app *cli.App, args []string) *cli.Context {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	for _, f := range app.Flags {
		if err := f.Apply(fs); err != nil {
			panic(err)
		}
	}
	if err := fs.Parse(args); err != nil {
		panic(err)
	}
	return cli.NewContext(app, fs, nil)
}

func Test_TODOsApp_version(t *testing.T) {
	t.Parallel()

	app := newTODOsApp()
	var b strings.Builder
	app.Writer = &b
	c := newContext(app, []string{"--version"})
	if err := app.Action(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	versionTitle := app.Name + " devel"
	if !strings.HasPrefix(b.String(), versionTitle) {
		t.Fatalf("expected %q in output: \n%q", versionTitle, b.String())
	}
}

func Test_TODOsApp_Walk(t *testing.T) {
	t.Parallel()

	dir := testutils.Must(os.MkdirTemp("", "code"))
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	files := map[string]string{
		"foo.go": "// TODO: foo",
		"bar.go": "// TODO: bar",
	}

	for path, src := range files {
		fullPath := filepath.Join(dir, path)
		testutils.Check(os.MkdirAll(filepath.Dir(fullPath), 0o700))
		testutils.Check(os.WriteFile(fullPath, []byte(src), 0o600))
	}

	app := newTODOsApp()
	var b strings.Builder
	app.Writer = &b
	c := newContext(app, []string{dir})
	if err := app.Action(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for path := range files {
		if !strings.Contains(b.String(), path) {
			t.Fatalf("expected %q in output: \n%q", path, b.String())
		}
	}
}

//nolint:paralleltest // modifies cli.OsExiter
func Test_TODOsApp_ExitErrHandler_ErrWalk(t *testing.T) {
	oldExiter := cli.OsExiter
	var exitCode *int
	cli.OsExiter = func(c int) {
		exitCode = &c
	}
	defer func() {
		cli.OsExiter = oldExiter
	}()

	app := newTODOsApp()
	var b strings.Builder
	app.ErrWriter = &b
	c := newContext(app, nil)
	app.ExitErrHandler(c, ErrWalk)

	if strings.Contains(b.String(), ErrWalk.Error()) {
		t.Fatalf("unexpected %q in output: \n%q", ErrWalk.Error(), b.String())
	}

	if exitCode == nil {
		t.Fatalf("unexpected exit code, want: %v, got: %v", ExitCodeWalkError, exitCode)
	}
	if diff := cmp.Diff(ExitCodeWalkError, *exitCode); diff != "" {
		t.Errorf("unexpected exit code (-want, +got): \n%s", diff)
	}
}

func Test_outCLI(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		ref      *walker.TODORef
		expected string
	}{
		"nil": {
			ref:      nil,
			expected: "",
		},
		"TODO": {
			ref: &walker.TODORef{
				FileName: "foo.go",
				TODO: &todos.TODO{
					Type: "TODO",
					Line: 16,
					Text: "// TODO: this is a message",
				},
			},
			expected: "foo.go:16:// TODO: this is a message\n",
		},
		"FIXME": {
			ref: &walker.TODORef{
				FileName: "foo.go",
				TODO: &todos.TODO{
					Type: "FIXME",
					Line: 16,
					Text: "// FIXME: this is a message",
				},
			},
			expected: "foo.go:16:// FIXME: this is a message\n",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var w strings.Builder
			h := outCLI(&w)
			err := h(tc.ref)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.expected, w.String()); diff != "" {
				t.Errorf("unexpected output (-want, +got): \n%s", diff)
			}
		})
	}
}

func Test_outGithub(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		ref      *walker.TODORef
		expected string
	}{
		"nil": {
			ref:      nil,
			expected: "",
		},
		"notice": {
			ref: &walker.TODORef{
				FileName: "foo.go",
				TODO: &todos.TODO{
					Type: "NOTE",
					Line: 16,
					Text: "// NOTE: this is a message",
				},
			},
			expected: "::notice file=foo.go,line=16::// NOTE: this is a message\n",
		},
		"TODO warning": {
			ref: &walker.TODORef{
				FileName: "foo.go",
				TODO: &todos.TODO{
					Type: "TODO",
					Line: 16,
					Text: "// TODO: this is a message",
				},
			},
			expected: "::warning file=foo.go,line=16::// TODO: this is a message\n",
		},
		"FIXME error": {
			ref: &walker.TODORef{
				FileName: "foo.go",
				TODO: &todos.TODO{
					Type: "FIXME",
					Line: 16,
					Text: "// FIXME: this is a message",
				},
			},
			expected: "::error file=foo.go,line=16::// FIXME: this is a message\n",
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var w strings.Builder
			h := outGithub(&w)
			err := h(tc.ref)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.expected, w.String()); diff != "" {
				t.Errorf("unexpected output (-want, +got): \n%s", diff)
			}
		})
	}
}

func Test_walkerOptionsFromContext(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		args  []string
		flags []cli.Flag

		expected *walker.Options
		err      error
	}{
		"no flags no args": {
			args:  nil,
			flags: nil,
			expected: &walker.Options{
				Config:        &todos.Config{},
				IncludeHidden: true,
				Paths:         []string{"."},
			},
		},
		"output github": {
			args: []string{"--output=github"},
			flags: []cli.Flag{
				&cli.StringFlag{
					Name: "output",
				},
			},
			// NOTE: Doesn't actually check TODOFunc.
			expected: &walker.Options{
				Config:        &todos.Config{},
				IncludeHidden: true,
				Paths:         []string{"."},
			},
		},
		"invalid output": {
			args: []string{"--output=foo"},
			flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "output",
					Value: "default",
				},
			},
			err: ErrFlagParse,
		},
		"todo-types": {
			args: []string{"--todo-types=TODO,FIXME"},
			flags: []cli.Flag{
				&cli.StringFlag{
					Name: "todo-types",
				},
			},
			expected: &walker.Options{
				Config: &todos.Config{
					Types: []string{"TODO", "FIXME"},
				},
				IncludeHidden: true,
				Paths:         []string{"."},
			},
		},
		"exclude-hidden": {
			args: []string{"--exclude-hidden"},
			flags: []cli.Flag{
				&cli.BoolFlag{
					Name: "exclude-hidden",
				},
			},
			expected: &walker.Options{
				Config:        &todos.Config{},
				IncludeHidden: false,
				Paths:         []string{"."},
			},
		},
		"include-vcs": {
			args: []string{"--include-vcs"},
			flags: []cli.Flag{
				&cli.BoolFlag{
					Name: "include-vcs",
				},
			},
			expected: &walker.Options{
				Config:        &todos.Config{},
				IncludeHidden: true,
				IncludeVCS:    true,
				Paths:         []string{"."},
			},
		},
		"include-vendored": {
			args: []string{"--include-vendored"},
			flags: []cli.Flag{
				&cli.BoolFlag{
					Name: "include-vendored",
				},
			},
			expected: &walker.Options{
				Config:          &todos.Config{},
				IncludeHidden:   true,
				IncludeVendored: true,
				Paths:           []string{"."},
			},
		},
		"paths": {
			args: []string{"/path/to/code"},
			expected: &walker.Options{
				Config:        &todos.Config{},
				IncludeHidden: true,
				Paths:         []string{"/path/to/code"},
			},
		},
		"multiple-paths": {
			args: []string{"/path/to/code", "/other/path"},
			expected: &walker.Options{
				Config:        &todos.Config{},
				IncludeHidden: true,
				Paths:         []string{"/path/to/code", "/other/path"},
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			app := &cli.App{
				Flags: tc.flags,
			}
			c := newContext(app, tc.args)

			o, err := walkerOptionsFromContext(c)

			if diff := cmp.Diff(tc.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("unexpected error (-want, +got): \n%s", diff)
			}
			if err == nil {
				// NOTE: Do not consider the ErrorFunc for comparison.
				if diff := cmp.Diff(tc.expected, o, cmpopts.IgnoreFields(walker.Options{}, "TODOFunc", "ErrorFunc")); diff != "" {
					t.Errorf("unexpected options (-want, +got): \n%s", diff)
				}
			}
		})
	}
}

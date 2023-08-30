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
	"bytes"
	"encoding/json"
	"flag"
	"strings"
	"testing"

	"github.com/gobwas/glob"
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

func Test_TODOsApp_help(t *testing.T) {
	t.Parallel()

	app := newTODOsApp()
	var b strings.Builder
	app.Writer = &b
	if err := app.Run([]string{"todos", "--help"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prefix := "NAME:"
	if !strings.HasPrefix(b.String(), prefix) {
		t.Fatalf("expected %q in output: \n%q", prefix, b.String())
	}
}

func Test_TODOsApp_help_arg(t *testing.T) {
	t.Parallel()

	app := newTODOsApp()
	var b strings.Builder
	app.Writer = &b
	// NOTE: somearg should be ignored.
	if err := app.Run([]string{"todos", "--help", "somearg"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prefix := "NAME:"
	if !strings.HasPrefix(b.String(), prefix) {
		t.Fatalf("expected %q in output: \n%q", prefix, b.String())
	}
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

	files := []*testutils.File{
		{
			Path:     "foo.go",
			Contents: []byte("// TODO: foo"),
			Mode:     0o600,
		},
		{
			Path:     "bar.go",
			Contents: []byte("// TODO: bar"),
			Mode:     0o600,
		},
	}

	d := testutils.NewTempDir(files)
	defer d.Cleanup()

	app := newTODOsApp()
	var b strings.Builder
	app.Writer = &b
	c := newContext(app, []string{d.Dir()})
	if err := app.Action(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := range files {
		if !strings.Contains(b.String(), files[i].Path) {
			t.Fatalf("expected %q in output: \n%q", files[i].Path, b.String())
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

func Test_outJSON(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		ref      *walker.TODORef
		expected *outTODO
	}{
		"nil": {
			ref:      nil,
			expected: nil,
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
			expected: &outTODO{
				Path: "foo.go",
				Type: "NOTE",
				Line: 16,
				Text: "// NOTE: this is a message",
			},
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
			expected: &outTODO{
				Path: "foo.go",
				Type: "TODO",
				Line: 16,
				Text: "// TODO: this is a message",
			},
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
			expected: &outTODO{
				Path: "foo.go",
				Type: "FIXME",
				Line: 16,
				Text: "// FIXME: this is a message",
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var w bytes.Buffer
			h := outJSON(&w)
			err := h(tc.ref)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.expected == nil {
				if diff := cmp.Diff("", w.String()); diff != "" {
					t.Errorf("unexpected output (-want, +got): \n%s", diff)
				}
				return
			}

			out := &outTODO{}
			if err := json.Unmarshal(w.Bytes(), &out); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.expected, out); diff != "" {
				t.Errorf("unexpected output (-want, +got): \n%s", diff)
			}
		})
	}
}

func Test_walkerOptionsFromContext(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		args []string

		expected *walker.Options
		err      error
	}{
		"no args": {
			args: nil,
			expected: &walker.Options{
				Config: &todos.Config{
					Types: todos.DefaultTypes,
				},
				Charset:       defaultCharset,
				IncludeHidden: true,
				Paths:         []string{"."},
			},
		},
		"output github": {
			args: []string{"--output=github"},
			// NOTE: Doesn't actually check TODOFunc.
			expected: &walker.Options{
				Config: &todos.Config{
					Types: todos.DefaultTypes,
				},
				Charset:       defaultCharset,
				IncludeHidden: true,
				Paths:         []string{"."},
			},
		},
		"invalid output": {
			args: []string{"--output=foo"},
			err:  ErrFlagParse,
		},
		"todo-types": {
			args: []string{"--todo-types=TODO,FIXME"},
			expected: &walker.Options{
				Config: &todos.Config{
					Types: []string{"TODO", "FIXME"},
				},
				Charset:       defaultCharset,
				IncludeHidden: true,
				Paths:         []string{"."},
			},
		},
		"exclude-hidden": {
			args: []string{"--exclude-hidden"},
			expected: &walker.Options{
				Config: &todos.Config{
					Types: todos.DefaultTypes,
				},
				Charset:       defaultCharset,
				IncludeHidden: false,
				Paths:         []string{"."},
			},
		},
		"include-vcs": {
			args: []string{"--include-vcs"},
			expected: &walker.Options{
				Config: &todos.Config{
					Types: todos.DefaultTypes,
				},
				Charset:       defaultCharset,
				IncludeHidden: true,
				IncludeVCS:    true,
				Paths:         []string{"."},
			},
		},
		"include-vendored": {
			args: []string{"--include-vendored"},
			expected: &walker.Options{
				Config: &todos.Config{
					Types: todos.DefaultTypes,
				},
				Charset:         defaultCharset,
				IncludeHidden:   true,
				IncludeVendored: true,
				Paths:           []string{"."},
			},
		},
		"paths": {
			args: []string{"/path/to/code"},
			expected: &walker.Options{
				Config: &todos.Config{
					Types: todos.DefaultTypes,
				},
				Charset:       defaultCharset,
				IncludeHidden: true,
				Paths:         []string{"/path/to/code"},
			},
		},
		"multiple-paths": {
			args: []string{"/path/to/code", "/other/path"},
			expected: &walker.Options{
				Config: &todos.Config{
					Types: todos.DefaultTypes,
				},
				Charset:       defaultCharset,
				IncludeHidden: true,
				Paths:         []string{"/path/to/code", "/other/path"},
			},
		},
		"exclude-multiple": {
			args: []string{"--exclude=exclude.*", "--exclude=foo"},
			expected: &walker.Options{
				Config: &todos.Config{
					Types: todos.DefaultTypes,
				},
				Charset:       defaultCharset,
				IncludeHidden: true,
				ExcludeGlobs:  []glob.Glob{glob.MustCompile("exclude.*"), glob.MustCompile("foo")},
				Paths:         []string{"."},
			},
		},
		"exclude-dir-multiple": {
			args: []string{"--exclude-dir=exclude?", "--exclude-dir=foo"},
			expected: &walker.Options{
				Config: &todos.Config{
					Types: todos.DefaultTypes,
				},
				Charset:         defaultCharset,
				IncludeHidden:   true,
				ExcludeDirGlobs: []glob.Glob{glob.MustCompile("exclude?"), glob.MustCompile("foo")},
				Paths:           []string{"."},
			},
		},
		"charset": {
			args: []string{"--charset=UTF-16"},
			expected: &walker.Options{
				Config: &todos.Config{
					Types: todos.DefaultTypes,
				},
				Charset:       "UTF-16",
				IncludeHidden: true,
				Paths:         []string{"."},
			},
		},
		"detect charset": {
			args: []string{"--charset=detect"},
			expected: &walker.Options{
				Config: &todos.Config{
					Types: todos.DefaultTypes,
				},
				Charset:       "detect",
				IncludeHidden: true,
				Paths:         []string{"."},
			},
		},
		"invalid charset": {
			args: []string{"--charset=invalid"},
			err:  ErrFlagParse,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			app := newTODOsApp()
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

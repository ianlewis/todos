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
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/urfave/cli/v2"

	"github.com/ianlewis/todos/internal/todos"
	"github.com/ianlewis/todos/internal/walker"
)

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

			app := cli.App{
				Flags: tc.flags,
			}
			app.Setup()
			fs := flag.NewFlagSet("", flag.ContinueOnError)
			for _, f := range tc.flags {
				if err := f.Apply(fs); err != nil {
					panic(err)
				}
			}
			if err := fs.Parse(tc.args); err != nil {
				panic(err)
			}
			c := cli.NewContext(&app, fs, nil)

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

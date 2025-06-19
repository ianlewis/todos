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
	"strings"
	"testing"

	"github.com/gobwas/glob"
	"github.com/google/go-cmp/cmp"

	"github.com/ianlewis/todos/internal/testutils"
	"github.com/ianlewis/todos/internal/todos"
)

type testCase struct {
	name string

	files []*testutils.File
	links []*testutils.Symlink
	opts  *Options

	expected []*TODORef
	err      bool
}

//nolint:gochecknoglobals // allow global table-driven tests.
var testCases = []testCase{
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
		name: "generated file skipped",
		files: []*testutils.File{
			{
				Path: filepath.Join("package-lock.json"),
				Contents: []byte(`{
				  "name": "example", // TODO: name
				  "version": "1.0.0",
				  "lockfileVersion": 1,
				  "dependencies": {}
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},

			// Ensure it's not skipped for another reason.
			IncludeHidden:   true,
			IncludeVendored: true,

			Charset: "UTF-8",
		},
		expected: nil,
	},
	{
		name: "generated file specified",
		files: []*testutils.File{
			{
				Path: filepath.Join("package-lock.json"),
				Contents: []byte(`{
				  "name": "example", // TODO: name
				  "version": "1.0.0",
				  "lockfileVersion": 1,
				  "dependencies": {}
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},

			Charset: "UTF-8",
			Paths:   []string{filepath.Join("package-lock.json")},
		},
		expected: []*TODORef{
			{
				FileName: filepath.Join("package-lock.json"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: name",
					Message:     "name",
					Line:        2,
					CommentLine: 2,
				},
			},
		},
	},
	{
		name: "generated file processed",
		files: []*testutils.File{
			{
				Path: filepath.Join("package-lock.json"),
				Contents: []byte(`{
				  "name": "example", // TODO: name
				  "version": "1.0.0",
				  "lockfileVersion": 1,
				  "dependencies": {}
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset:          "UTF-8",
			IncludeGenerated: true,
		},
		expected: []*TODORef{
			{
				FileName: filepath.Join("package-lock.json"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: name",
					Message:     "name",
					Line:        2,
					CommentLine: 2,
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
		name: "vendor directory specified",
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
			Paths:   []string{"vendor"},
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
	{
		name: "todo last line",
		files: []*testutils.File{
			{
				Path: filepath.Join("src", "line_comments.go"),
				Contents: []byte(`package foo
				// package comment

				func TODO() {
					return // Random comment
				}
				// TODO: last line`),
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
				FileName: filepath.Join("src", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: last line",
					Message:     "last line",
					Line:        7,
					CommentLine: 7,
				},
			},
		},
	},
	{
		name: "label glob filter",
		files: []*testutils.File{
			{
				Path: "line_comments.go",
				Contents: []byte(`package foo
				// package comment

				// TODO: no label
				// TODO(no-match): no match

				// TODO is a function.
				// TODO(todo-label): some task.
				func TODO() {
					return // TODO(other-label): Return comment
				}`),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			LabelGlobs: []glob.Glob{
				glob.MustCompile("todo-*"),
				glob.MustCompile("other-*"),
			},
			Charset: "UTF-8",
		},
		expected: []*TODORef{
			{
				FileName: "line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO(todo-label): some task.",
					Message:     "some task.",
					Label:       "todo-label",
					Line:        8,
					CommentLine: 8,
				},
			},
			{
				FileName: "line_comments.go",
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO(other-label): Return comment",
					Message:     "Return comment",
					Label:       "other-label",
					Line:        10,
					CommentLine: 10,
				},
			},
		},
	},
	{
		name: "gitignore file skipped",
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
				Path:     ".gitignore",
				Contents: []byte(`line_comments.go`),
				Mode:     0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			IgnoreFileNames: []string{".gitignore"},
			Charset:         "UTF-8",
		},
		expected: nil,
	},
	{
		name: "gitignore file specified",
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
				Path:     ".gitignore",
				Contents: []byte(`line_comments.go`),
				Mode:     0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset:         "UTF-8",
			IgnoreFileNames: []string{".gitignore"},
			Paths:           []string{"line_comments.go"},
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
		name: "gitignore file included",
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
				Path: ".gitignore",
				Contents: []byte(
					strings.Join([]string{
						`*.go`,
						`!line_comments.go`,
					}, "\n")),
				Mode: 0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset:         "UTF-8",
			IgnoreFileNames: []string{".gitignore"},
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
		name: "gitignore dir skipped",
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
				Path:     ".gitignore",
				Contents: []byte(`src`),
				Mode:     0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			IgnoreFileNames: []string{".gitignore"},
			Charset:         "UTF-8",
		},
		expected: nil,
	},
	{
		name: "gitignore dir specified",
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
				Path:     ".gitignore",
				Contents: []byte(`src`),
				Mode:     0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset:         "UTF-8",
			IgnoreFileNames: []string{".gitignore"},
			Paths:           []string{"src"},
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
	{
		name: "gitignore subdir file skipped",
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
				Path:     filepath.Join("src", ".gitignore"),
				Contents: []byte(`line_comments.go`),
				Mode:     0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			IgnoreFileNames: []string{".gitignore"},
			Charset:         "UTF-8",
		},
		expected: nil,
		// expected: []*TODORef{},
	},
	{
		name: "binary files are ignored",
		files: []*testutils.File{
			{
				Path:     "binary.exe",
				Contents: []byte{1, 2, 0, 3},
				Mode:     0o600,
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
		name: "unsupported files are ignored",
		files: []*testutils.File{
			{
				Path:     "unsupported_lang.coq",
				Contents: []byte{},
				Mode:     0o600,
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
		name: "unsupported files generate error if specified",
		files: []*testutils.File{
			{
				Path:     "unsupported_lang.coq",
				Contents: []byte{},
				Mode:     0o600,
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			Charset: "UTF-8",
			Paths:   []string{"unsupported_lang.coq"},
		},
		expected: nil,
		err:      true,
	},
}

// symlinkTestCases contains test cases for symlinked files and directories.
//
//nolint:gochecknoglobals // allow global table-driven tests.
var symlinkTestCases = []testCase{
	{
		name: "code is symlinked",
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
		links: []*testutils.Symlink{
			{
				Path:   "symlink.go",
				Target: "line_comments.go",
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			FollowSymlinks: true,
			Charset:        "UTF-8",
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
				FileName: "symlink.go",
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
		name: "code is symlinked no follow",
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
		links: []*testutils.Symlink{
			{
				Path:   "symlink.go",
				Target: "line_comments.go",
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			FollowSymlinks: false, // Do not follow symlinks.
			Charset:        "UTF-8",
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
		name: "symlinked dir",
		files: []*testutils.File{
			{
				Path: filepath.Join("sub-dir", "line_comments.go"),
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
		links: []*testutils.Symlink{
			{
				Path:   "sym-dir",
				Target: "sub-dir",
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			FollowSymlinks: true,
			Charset:        "UTF-8",
		},
		expected: []*TODORef{
			{
				FileName: filepath.Join("sub-dir", "line_comments.go"),
				TODO: &todos.TODO{
					Type:        "TODO",
					Text:        "// TODO: some task.",
					Message:     "some task.",
					Line:        5,
					CommentLine: 5,
				},
			},
			{
				FileName: filepath.Join("sym-dir", "line_comments.go"),
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
		name: "symlinked dir no follow",
		files: []*testutils.File{
			{
				Path: filepath.Join("sub-dir", "line_comments.go"),
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
		links: []*testutils.Symlink{
			{
				Path:   "sym-dir",
				Target: "sub-dir",
			},
		},
		opts: &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
			FollowSymlinks: false, // Do not follow symlinks.
			Charset:        "UTF-8",
		},
		expected: []*TODORef{
			{
				FileName: filepath.Join("sub-dir", "line_comments.go"),
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

type blameTestCase struct {
	testCase

	author string
	email  string
}

//nolint:gochecknoglobals // allow global table-driven tests.
var blameTestCases = []blameTestCase{}

//nolint:gochecknoinits // init used to initialize test cases.
func init() {
	author := "John Doe"
	email := "john@doe.com"

	for _, tc := range testCases {
		opts := Options{}
		if tc.opts != nil {
			opts = *tc.opts
		}

		opts.Blame = true

		var expected []*TODORef
		for _, e := range tc.expected {
			// Copy
			e2 := *e
			e2.GitUser = &GitUser{
				Name:  author,
				Email: email,
			}
			expected = append(expected, &e2)
		}

		blameTestCases = append(blameTestCases, blameTestCase{
			testCase: testCase{
				name: tc.name,

				files:    tc.files,
				links:    tc.links,
				opts:     &opts,
				expected: expected,
				err:      tc.err,
			},

			author: author,
			email:  email,
		})
	}
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

func newFixture(files []*testutils.File, links []*testutils.Symlink, opts *Options) (*fixture, *TODOWalker) {
	dir := testutils.NewTempDir(files, links)

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

type repoFixture struct {
	tmpDir *testutils.TempDir
	repo   *testutils.TestRepo
	wd     string
	out    []*TODORef
	err    []error
}

func (f *repoFixture) cleanup() {
	testutils.Check(os.Chdir(f.wd))
	f.tmpDir.Cleanup()
}

// newRepoFixture creates a fixture with a temporary directory with a git
// repository created at the root. Files are created and committed into the git
// repository.
func newRepoFixture(author, email string,
	files []*testutils.File, links []*testutils.Symlink,
	opts *Options,
) (*repoFixture, *TODOWalker) {
	return newDirRepoFixture(author, email, ".", files, nil, links, nil, opts)
}

// newDirRepoFixture creates a fixture with a temporary directory and a git
// repository inside of it. dirFiles are created relative to the temporary
// directory root and are not committed to the git repository. repoSubPath
// specifies the repository's directory relative to the temporary directory root.
// repoFiles are created relative to the repository root and committed to the
// repository. The fixture sets the working directory to the temporary
// directory.
func newDirRepoFixture(
	author, email, repoSubPath string,
	repoFiles, dirFiles []*testutils.File,
	repoLinks, dirLinks []*testutils.Symlink,
	opts *Options,
) (*repoFixture, *TODOWalker) {
	tmpDir := testutils.NewTempDir(dirFiles, dirLinks)

	repo := testutils.NewTestRepo(
		filepath.Join(tmpDir.Dir(), repoSubPath),
		author, email,
		repoFiles, repoLinks,
	)

	cleanup, cancel := testutils.WithCancel(func() {
		tmpDir.Cleanup()
	}, nil)
	defer cleanup()

	wd := testutils.Must(os.Getwd())

	testutils.Check(os.Chdir(tmpDir.Dir()))

	if len(opts.Paths) == 0 {
		opts.Paths = []string{"."}
	}

	fixture := &repoFixture{
		tmpDir: tmpDir,
		repo:   repo,
		wd:     wd,
	}

	todoFunc := opts.TODOFunc
	opts.TODOFunc = func(r *TODORef) error {
		fixture.out = append(fixture.out, r)

		if todoFunc != nil {
			return todoFunc(r)
		}

		return nil
	}

	errFunc := opts.ErrorFunc
	opts.ErrorFunc = func(err error) error {
		fixture.err = append(fixture.err, err)

		if errFunc != nil {
			return errFunc(err)
		}

		return nil
	}

	opts.Blame = true

	cancel()

	return fixture, New(opts)
}

//nolint:paralleltest // fixture uses Chdir and cannot be run in parallel.
func TestTODOWalker(t *testing.T) {
	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			f, w := newFixture(tc.files, tc.links, tc.opts)
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
func TestTODOWalker_symlink(t *testing.T) {
	for i := range symlinkTestCases {
		tc := symlinkTestCases[i]

		t.Run(tc.name, func(t *testing.T) {
			f, w := newFixture(tc.files, tc.links, tc.opts)
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
func TestTODOWalker_git(t *testing.T) {
	for _, tc := range blameTestCases {
		t.Run(tc.name, func(t *testing.T) {
			f, w := newRepoFixture(tc.author, tc.email, tc.files, tc.links, tc.opts)
			defer f.cleanup()

			if got, want := w.Walk(), tc.err; got != want {
				t.Errorf("unexpected error code, got: %v, want: %v\nw.err: %v", got, want, w.err)
			}

			for i := range tc.expected {
				tc.expected[i].GitUser = &GitUser{
					Name:  tc.author,
					Email: tc.email,
				}
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

	f, w := newFixture(files, nil, opts)
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
		TODOFunc: func(_ *TODORef) error {
			return fs.SkipAll
		},
	}

	f, w := newFixture([]*testutils.File{file}, nil, opts)
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

//nolint:paralleltest // fixture uses Chdir and cannot be run in parallel.
func TestTODOWalker_gitSubDir(t *testing.T) {
	dirFiles := []*testutils.File{
		{
			Path: "line_comments.rb",
			Contents: []byte(`# file comment

			# TODO: some task.
			def foo()
				# Random comment
				x = "\"# Random comment x"
				y = '\'# Random comment y'
				return x + y
			end`),
			Mode: 0o600,
		},
	}

	repoSubPath := "path/to/repo"
	author := "John Doe"
	email := "john@doe.com"
	repoFiles := []*testutils.File{
		{
			Path: "repo_file.go",
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
	}

	expected := []*TODORef{
		{
			FileName: "line_comments.rb",
			TODO: &todos.TODO{
				Type:        "TODO",
				Text:        "# TODO: some task.",
				Message:     "some task.",
				Line:        3,
				CommentLine: 3,
			},
		},
		{
			FileName: filepath.Join("path", "to", "repo", "repo_file.go"),
			TODO: &todos.TODO{
				Type:        "TODO",
				Text:        "// TODO: some task.",
				Message:     "some task.",
				Line:        5,
				CommentLine: 5,
			},
			GitUser: &GitUser{
				Name:  author,
				Email: email,
			},
		},
	}

	f, w := newDirRepoFixture(author, email, repoSubPath, repoFiles, dirFiles, nil, nil, opts)

	if got, want := w.Walk(), false; got != want {
		t.Errorf("unexpected error code, got: %v, want: %v\nw.err: %v", got, want, w.err)
	}

	got, want := f.out, expected
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(TODORef{})); diff != "" {
		t.Errorf("unexpected output (-want +got):\n%s", diff)
	}
}

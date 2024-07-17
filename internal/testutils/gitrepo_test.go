// Copyright 2024 Google LLC
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

package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTempRepo(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		author      string
		email       string
		files       []*File
		expectPanic bool
	}{
		"no files": {
			author: "John Doe",
			email:  "john@doe.com",
		},
		"single file": {
			author: "John Doe",
			email:  "john@doe.com",
			files: []*File{
				{
					Path:     "testfile.txt",
					Contents: []byte("foo"),
					Mode:     0o600,
				},
			},
		},
		"single file with sub-dir": {
			author: "John Doe",
			email:  "john@doe.com",
			files: []*File{
				{
					Path:     "testdir/testfile.txt",
					Contents: []byte("foo"),
					Mode:     0o600,
				},
			},
		},
		"bad file": {
			author: "John Doe",
			email:  "john@doe.com",
			files: []*File{
				{
					Path:     "../../../../../../../testfile.txt",
					Contents: []byte("foo"),
					Mode:     0o600,
				},
			},
			expectPanic: true,
		},
		"multi-file": {
			author: "John Doe",
			email:  "john@doe.com",
			files: []*File{
				{
					Path:     "testfile.txt",
					Contents: []byte("foo"),
					Mode:     0o600,
				},
				{
					Path:     "anotherfile.txt",
					Contents: []byte("bar"),
					Mode:     0o600,
				},
			},
		},
		"multi-file with sub-dir": {
			author: "John Doe",
			email:  "john@doe.com",
			files: []*File{
				{
					Path:     "testdir/testfile.txt",
					Contents: []byte("foo"),
					Mode:     0o600,
				},
				{
					Path:     "testdir/otherfile.txt",
					Contents: []byte("bar"),
					Mode:     0o600,
				},
			},
		},
		"multi-file with multi-sub-dir": {
			author: "John Doe",
			email:  "john@doe.com",
			files: []*File{
				{
					Path:     "testdir/testfile.txt",
					Contents: []byte("foo"),
					Mode:     0o600,
				},
				{
					Path:     "otherdir/otherfile.txt",
					Contents: []byte("bar"),
					Mode:     0o600,
				},
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				if r := recover(); r != nil && !tc.expectPanic {
					t.Fatalf("unexpected panic: %v", r)
				}
			}()

			d := NewTempRepo(tc.author, tc.email, tc.files)
			baseDir := d.Dir()
			defer func() {
				_ = os.RemoveAll(baseDir)
			}()

			// Check that the temporary directory exists.
			dirStat, err := os.Stat(baseDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !dirStat.IsDir() {
				t.Fatalf("baseDir %q not a directory", baseDir)
			}

			// Check the .git/ directory exists.
			gitDir := filepath.Join(baseDir, ".git")
			gitDirStat, err := os.Stat(gitDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !gitDirStat.IsDir() {
				t.Fatalf("gitDir %q not a directory", gitDir)
			}

			for _, f := range tc.files {
				fullPath := filepath.Join(baseDir, f.Path)
				info, err := os.Stat(fullPath)
				if err != nil {
					t.Fatalf("os.Stat: %v", err)
				}

				if info.IsDir() {
					t.Fatalf("unexpected directory: %q", fullPath)
				}

				if got, want := info.Mode(), f.Mode; !compareMode(got, want) {
					t.Errorf("unexpected mode, got: %v, want: %v", got, want)
				}

				b, err := os.ReadFile(fullPath)
				if err != nil {
					t.Fatalf("os.ReadFile: %q", fullPath)
				}

				got, want := b, f.Contents
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("unexpected contents: (-want, +got): %s", diff)
				}
			}

			d.Cleanup()
			if _, err := os.Stat(baseDir); !os.IsNotExist(err) {
				t.Fatalf("expected not exist error: %v", err)
			}
		})
	}
}

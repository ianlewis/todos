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

package testutils

import (
	"os"
	"path/filepath"
	"testing"
)

// TempDir is a temporary directory which is set up with a directory structure
// with some files for testing.
type TempDir struct {
	dir string
}

// File is a test file.
type File struct {
	// Path is the path to the file.
	Path string

	// Contents are the file contents.
	Contents []byte

	// Mode is the file permissions mode.
	Mode os.FileMode
}

// Symlink is a test symbolic link.
type Symlink struct {
	// Path is the path to the symbolic link.
	Path string

	// Target is path to the target of the symbolic link relative to the
	// symbolic link.
	Target string
}

// NewTempDir creates a new TempDir. This creates a new temporary directory and
// fills it with the files given. Intermediate directories are created
// automatically with 0700 permissions. This function panics if an error occurs
// when creating the files.
func NewTempDir(t *testing.T, files []*File, links []*Symlink) *TempDir {
	t.Helper()

	d := &TempDir{}

	d.dir = t.TempDir()

	const readWriteExec = os.FileMode(0o700)

	for _, file := range files {
		fullPath := filepath.Join(d.dir, file.Path)
		Check(t, os.MkdirAll(filepath.Dir(fullPath), readWriteExec))
		Check(t, os.WriteFile(fullPath, file.Contents, file.Mode))
	}

	for _, link := range links {
		fullPath := filepath.Join(d.dir, link.Path)
		Check(t, os.MkdirAll(filepath.Dir(fullPath), readWriteExec))
		Check(t, os.Symlink(link.Target, fullPath))
	}

	return d
}

// Dir returns the path to the directory.
func (d *TempDir) Dir() string {
	return d.dir
}

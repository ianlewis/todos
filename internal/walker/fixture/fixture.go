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

package fixture

import (
	"os"
	"path/filepath"

	"github.com/ianlewis/todos/internal/testutils"
	"github.com/ianlewis/todos/internal/walker"
)

// Fixture is a fixture for a TODOWalker which allows walking over a directory
// set up for testing.
type Fixture struct {
	opts *walker.Options
	wd   string
	dir  string
	out  []*walker.TODORef
	err  []error
}

// File is a test file.
type File struct {
	// Path is the path to the file.
	Path string

	// Contents are the file contents.
	Contents []byte

	// Mode is the file permissions mode.
	Mode os.FileMode

	// IsDir specifies if it's a directory.
	IsDir bool
}

// New creates a new Fixture. This creates a new temporary directory and fills
// it with the files given. It then changes the current working directory to
// temporary directory so the walker can walk that directory using the relative
// paths. Intermediate directories are created automatically with 0700 permissions.
func New(opts *walker.Options, files []File) (*Fixture, *walker.TODOWalker) {
	f := &Fixture{}
	f.wd = testutils.Must(os.Getwd())

	f.dir = testutils.Must(os.MkdirTemp("", "code"))
	cleanup, cancel := testutils.WithCancel(func() {
		f.Cleanup()
	}, nil)
	defer cleanup()
	testutils.Check(os.Chdir(f.dir))

	for _, file := range files {
		fullPath := filepath.Join(f.dir, file.Path)
		if !file.IsDir {
			testutils.Check(os.MkdirAll(filepath.Dir(fullPath), 0o700))
			testutils.Check(os.WriteFile(fullPath, file.Contents, file.Mode))
		} else {
			testutils.Check(os.MkdirAll(fullPath, file.Mode))
		}
	}

	if len(opts.Paths) == 0 {
		opts.Paths = []string{"."}
	}

	f.opts = opts
	opts.TODOFunc = f.outFunc
	opts.ErrorFunc = f.errFunc
	walker := walker.New(opts)

	cancel()
	return f, walker
}

func (f *Fixture) outFunc(r *walker.TODORef) error {
	f.out = append(f.out, r)
	if f.opts.TODOFunc != nil {
		return f.opts.TODOFunc(r)
	}
	return nil
}

func (f *Fixture) errFunc(err error) error {
	f.err = append(f.err, err)
	if f.opts.ErrorFunc != nil {
		return f.opts.ErrorFunc(err)
	}
	return nil
}

// TODOs returns the TODORefs encountered.
func (f *Fixture) TODOs() []*walker.TODORef {
	return f.out
}

// Err returns the errors encountered in order.
func (f *Fixture) Err() []error {
	return f.err
}

// Cleanup deletes the test directory.
func (f *Fixture) Cleanup() {
	testutils.Check(os.Chdir(f.wd))
	testutils.Check(os.RemoveAll(f.dir))
}

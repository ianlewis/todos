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
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// TempRepo is a repo created in a temporary directory.
type TempRepo struct {
	dir  string
	repo *git.Repository
}

// NewTempRepo creates a new TempRepo with the given files committed into a
// single commit in the repo.
func NewTempRepo(author, email string, files []*File) *TempRepo {
	r := &TempRepo{}
	r.dir = Must(os.MkdirTemp("", "testutils"))

	r.repo = Must(git.PlainInit(r.dir, false))

	w := Must(r.repo.Worktree())

	if len(files) > 0 {
		for _, f := range files {
			fullPath := filepath.Join(r.dir, f.Path)

			// Create necessary sub-directories.
			Check(os.MkdirAll(filepath.Dir(fullPath), 0o700))

			// Write the file
			Check(os.WriteFile(fullPath, f.Contents, f.Mode))

			// git add <file>
			_ = Must(w.Add(f.Path))
		}

		// git commit <file>
		_ = Must(w.Commit("test commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  author,
				Email: email,
				When:  time.Now(),
			},
		}))
	}

	return r
}

// Dir returns the path to the repo root directory.
func (r *TempRepo) Dir() string {
	return r.dir
}

// Repository returns the git repository.
func (r *TempRepo) Repository() *git.Repository {
	return r.repo
}

// Cleanup deletes the test repository.
func (r *TempRepo) Cleanup() {
	Check(os.RemoveAll(r.dir))
}

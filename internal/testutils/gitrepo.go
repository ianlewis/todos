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
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// TestRepo is a repo created in a temporary directory.
type TestRepo struct {
	dir  string
	repo *git.Repository
}

// NewTestRepo creates a new TestRepo with the given files committed into a
// single commit in the repo.
func NewTestRepo(t *testing.T, dir, author, email string, files []*File, links []*Symlink) *TestRepo {
	t.Helper()

	testRepo := &TestRepo{
		dir: dir,
	}
	repoDir, err := git.PlainInit(testRepo.dir, false)
	testRepo.repo = Must(t, repoDir, err)

	worktree, err := testRepo.repo.Worktree()
	Check(t, err)

	const readWriteExec = os.FileMode(0o700)

	if len(files) > 0 {
		for _, f := range files {
			fullPath := filepath.Join(testRepo.dir, f.Path)

			// Create necessary sub-directories.
			Check(t, os.MkdirAll(filepath.Dir(fullPath), readWriteExec))

			// Write the file
			Check(t, os.WriteFile(fullPath, f.Contents, f.Mode))

			// git add <file>
			_, err := worktree.Add(f.Path)
			Check(t, err)
		}

		// git commit <file>
		_, err := worktree.Commit("add files", &git.CommitOptions{
			Author: &object.Signature{
				Name:  author,
				Email: email,
				When:  time.Now(),
			},
		})
		Check(t, err)
	}

	if len(links) > 0 {
		for _, link := range links {
			fullPath := filepath.Join(testRepo.dir, link.Path)

			// Create necessary sub-directories.
			Check(t, os.MkdirAll(filepath.Dir(fullPath), readWriteExec))

			// Create the symbolic link.
			Check(t, os.Symlink(link.Target, fullPath))

			// git add <file>
			_, err := worktree.Add(link.Path)
			Check(t, err)
		}

		// git commit <file>
		_, err := worktree.Commit("add links", &git.CommitOptions{
			Author: &object.Signature{
				Name:  author,
				Email: email,
				When:  time.Now(),
			},
		})
		Check(t, err)
	}

	return testRepo
}

// Dir returns the path to the repo root directory.
func (r *TestRepo) Dir() string {
	return r.dir
}

// Repository returns the git repository.
func (r *TestRepo) Repository() *git.Repository {
	return r.repo
}

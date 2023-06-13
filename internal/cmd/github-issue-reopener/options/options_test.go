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

package options

import (
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ianlewis/todos/internal/testutils"
)

func merge(dst, src map[string]string) map[string]string {
	newDst := make(map[string]string, len(dst))
	for k, v := range dst {
		newDst[k] = v
	}

	for k, v := range src {
		newDst[k] = v
	}
	return newDst
}

//nolint:paralleltest // fixture uses Getenv/Setenv and cannot be run in parallel.
func TestNew(t *testing.T) {
	defaultOwner := "ianlewis"
	defaultRepo := "todos"
	defaultSha := "2d558f8cea415d4635bbf9c2ac484f9680f707e7"
	githubEnv := map[string]string{
		"GITHUB_SHA":        defaultSha,
		"GITHUB_REPOSITORY": defaultOwner + "/" + defaultRepo,
	}
	defaultPaths := []string{"."}

	testCases := map[string]struct {
		args     []string
		env      map[string]string
		token    string
		expected *Options
		err      error
	}{
		"default": {
			args: []string{"todos"},
			env:  githubEnv,
			expected: &Options{
				RepoOwner: defaultOwner,
				RepoName:  defaultRepo,
				Sha:       defaultSha,
				Paths:     defaultPaths,
			},
		},
		"invalid repo": {
			args: []string{
				"todos",
				"--repo=test",
			},
			env: githubEnv,
			err: ErrFlagParse,
		},
		"valid repo": {
			args: []string{
				"todos",
				"--repo=ianlewis/go-stardict",
			},
			env: githubEnv,
			expected: &Options{
				RepoOwner: "ianlewis",
				RepoName:  "go-stardict",
				Sha:       defaultSha,
				Paths:     defaultPaths,
			},
		},
		"github_token": {
			args: []string{"todos"},
			env: merge(githubEnv, map[string]string{
				"GITHUB_TOKEN": "test token",
			}),
			expected: &Options{
				RepoOwner: defaultOwner,
				RepoName:  defaultRepo,
				Sha:       defaultSha,
				Token:     "test token",
				Paths:     defaultPaths,
			},
		},
		"token file": {
			args: []string{"todos"},
			env:  githubEnv,
			// NOTE: --token-file flag added automatically.
			token: "test token from file",
			expected: &Options{
				RepoOwner: defaultOwner,
				RepoName:  defaultRepo,
				Sha:       defaultSha,
				Token:     "test token from file",
				Paths:     defaultPaths,
			},
		},
		"dry-run": {
			args: []string{"todos", "--dry-run"},
			env:  githubEnv,
			expected: &Options{
				DryRun:    true,
				RepoOwner: defaultOwner,
				RepoName:  defaultRepo,
				Sha:       defaultSha,
				Paths:     defaultPaths,
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			if tc.token != "" {
				f := testutils.Must(os.CreateTemp("", "TestNew"))
				_ = testutils.Must(io.WriteString(f, tc.token))
				tc.args = append(tc.args, "--token-file", f.Name())
				defer os.Remove(f.Name())
			}

			opts, err := New(tc.args)
			if diff := cmp.Diff(tc.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("unexpected error (-want +got):\n%s", diff)
			}

			got, want := opts, tc.expected
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("unexpected Options (-want +got):\n%s", diff)
			}
		})
	}
}

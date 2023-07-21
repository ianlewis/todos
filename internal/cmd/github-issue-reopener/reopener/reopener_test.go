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

package reopener

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-github/v52/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"

	"github.com/ianlewis/todos/internal/testutils"
	"github.com/ianlewis/todos/internal/todos"
	"github.com/ianlewis/todos/internal/walker"
)

func Test_labelMatch(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		label string
		match []string
	}{
		"no match": {
			label: "No Match",
			match: nil,
		},
		"http url match": {
			label: "http://github.com/ianlewis/todos/issues/123",
			match: []string{
				"http://github.com/ianlewis/todos/issues/123",
				"http://github.com/ianlewis/todos/issues/",
				"http://",
				"ianlewis",
				"todos",
				"123",
			},
		},
		"https url match": {
			label: "https://github.com/ianlewis/todos/issues/123",
			match: []string{
				"https://github.com/ianlewis/todos/issues/123",
				"https://github.com/ianlewis/todos/issues/",
				"https://",
				"ianlewis",
				"todos",
				"123",
			},
		},
		"no scheme url match": {
			label: "github.com/ianlewis/todos/issues/123",
			match: []string{
				"github.com/ianlewis/todos/issues/123",
				"github.com/ianlewis/todos/issues/",
				"",
				"ianlewis",
				"todos",
				"123",
			},
		},
		"# match": {
			label: "#123",
			match: []string{
				"#123",
				"#",
				"",
				"",
				"",
				"123",
			},
		},
		"number only match": {
			label: "123",
			match: []string{
				"123",
				"",
				"",
				"",
				"",
				"123",
			},
		},
		"whitespace": {
			label: "   123    ",
			match: []string{
				"   123    ",
				"",
				"",
				"",
				"",
				"123",
			},
		},
		"invalid url": {
			label: "github.com/ianlewis/todos/123",
			match: nil,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, want := labelMatch.FindStringSubmatch(tc.label), tc.match
			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatalf("unexpected match (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_handleTODO(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		owner    string
		repo     string
		ref      []*walker.TODORef
		expected map[int]*IssueRef
		err      error
	}{
		"test.go": {
			ref: []*walker.TODORef{
				{
					FileName: "test.go",
					TODO: &todos.TODO{
						Type:        "TODO",
						Text:        "// TODO(#123): Foo",
						Label:       "#123",
						Message:     "Foo",
						Line:        5,
						CommentLine: 5,
					},
				},
			},
			expected: map[int]*IssueRef{
				123: {
					Number: 123,
					TODOs: []*walker.TODORef{
						{
							FileName: "test.go",
							TODO: &todos.TODO{
								Type:        "TODO",
								Text:        "// TODO(#123): Foo",
								Label:       "#123",
								Message:     "Foo",
								Line:        5,
								CommentLine: 5,
							},
						},
					},
				},
			},
		},
		"url_label.go": {
			owner: "ianlewis",
			repo:  "todos",
			ref: []*walker.TODORef{
				{
					FileName: "test.go",
					TODO: &todos.TODO{
						Type:        "TODO",
						Text:        "// TODO(github.com/ianlewis/todos/issues/123): Foo",
						Label:       "github.com/ianlewis/todos/issues/123",
						Message:     "Foo",
						Line:        5,
						CommentLine: 5,
					},
				},
			},
			expected: map[int]*IssueRef{
				123: {
					Number: 123,
					TODOs: []*walker.TODORef{
						{
							FileName: "test.go",
							TODO: &todos.TODO{
								Type:        "TODO",
								Text:        "// TODO(github.com/ianlewis/todos/issues/123): Foo",
								Label:       "github.com/ianlewis/todos/issues/123",
								Message:     "Foo",
								Line:        5,
								CommentLine: 5,
							},
						},
					},
				},
			},
		},
		"multi-references": {
			owner: "ianlewis",
			repo:  "todos",
			ref: []*walker.TODORef{
				{
					FileName: "test.go",
					TODO: &todos.TODO{
						Type:        "TODO",
						Text:        "// TODO(#123): Foo",
						Label:       "#123",
						Message:     "Foo",
						Line:        5,
						CommentLine: 5,
					},
				},
				{
					FileName: "other.go",
					TODO: &todos.TODO{
						Type:        "TODO",
						Text:        "// TODO(github.com/ianlewis/todos/123): Bar",
						Label:       "github.com/ianlewis/todos/issues/123",
						Message:     "Bar",
						Line:        5,
						CommentLine: 5,
					},
				},
			},
			expected: map[int]*IssueRef{
				123: {
					Number: 123,
					TODOs: []*walker.TODORef{
						{
							FileName: "test.go",
							TODO: &todos.TODO{
								Type:        "TODO",
								Text:        "// TODO(#123): Foo",
								Label:       "#123",
								Message:     "Foo",
								Line:        5,
								CommentLine: 5,
							},
						},
						{
							FileName: "other.go",
							TODO: &todos.TODO{
								Type:        "TODO",
								Text:        "// TODO(github.com/ianlewis/todos/123): Bar",
								Label:       "github.com/ianlewis/todos/issues/123",
								Message:     "Bar",
								Line:        5,
								CommentLine: 5,
							},
						},
					},
				},
			},
		},
		"foo_label.go": {
			ref: []*walker.TODORef{
				{
					FileName: "test.go",
					TODO: &todos.TODO{
						Type:        "TODO",
						Text:        "// TODO(foo): Foo",
						Label:       "",
						Message:     "Foo",
						Line:        5,
						CommentLine: 5,
					},
				},
			},
			expected: nil,
		},
		"different_repo.go": {
			owner: "ianlewis",
			repo:  "todos",
			ref: []*walker.TODORef{
				{
					FileName: "test.go",
					TODO: &todos.TODO{
						Type:        "TODO",
						Text:        "// TODO(github.com/ianlewis/otherrepo/issues/123): Foo",
						Label:       "github.com/otherowner/todos/issues/123",
						Message:     "Foo",
						Line:        5,
						CommentLine: 5,
					},
				},
			},
			expected: nil,
		},
		"different_owner.go": {
			owner: "ianlewis",
			repo:  "todos",
			ref: []*walker.TODORef{
				{
					FileName: "test.go",
					TODO: &todos.TODO{
						Type:        "TODO",
						Text:        "// TODO(github.com/otherowner/todos/issues/123): Foo",
						Label:       "github.com/otherowner/todos/issues/123",
						Message:     "Foo",
						Line:        5,
						CommentLine: 5,
					},
				},
			},
			expected: nil,
		},
		"issue_no_bad_format.go": {
			owner: "ianlewis",
			repo:  "todos",
			ref: []*walker.TODORef{
				{
					FileName: "test.go",
					TODO: &todos.TODO{
						Type:        "TODO",
						Text:        "// TODO(github.com/ianlewis/todos/issues/foo): Foo",
						Label:       "github.com/otherowner/todos/issues/foo",
						Message:     "Foo",
						Line:        5,
						CommentLine: 5,
					},
				},
			},
			expected: nil,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r := New(context.Background(), &Options{
				RepoOwner: tc.owner,
				RepoName:  tc.repo,
			})
			for i := range tc.ref {
				err := r.handleTODO(tc.ref[i])
				if diff := cmp.Diff(tc.err, err, cmpopts.EquateErrors()); diff != "" {
					t.Fatalf("unexpected error (-want +got):\n%s", diff)
				}
			}

			if diff := cmp.Diff(tc.expected, r.refs.cache); diff != "" {
				t.Fatalf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_handleErr(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		err      error
		rerr     error
		expected error
	}{
		"nil": {
			err:      nil,
			rerr:     nil,
			expected: nil,
		},
		"EOF": {
			err:      io.EOF,
			rerr:     io.EOF,
			expected: nil,
		},
		"canceled": {
			err:      context.Canceled,
			rerr:     context.Canceled,
			expected: context.Canceled,
		},
		"deadline": {
			err:      context.DeadlineExceeded,
			rerr:     context.DeadlineExceeded,
			expected: context.DeadlineExceeded,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r := New(context.Background(), &Options{})
			err := r.handleErr(tc.err)
			if diff := cmp.Diff(tc.rerr, r.err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("unexpected error (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.expected, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("unexpected error (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_loadIssue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		issueNumber int
		owner       string
		repo        string
		cached      *github.Issue
		remote      *github.Issue

		expected *github.Issue
		err      error
	}{
		"cached no remote": {
			issueNumber: 123,
			owner:       "ianlewis",
			repo:        "todos",
			cached: &github.Issue{
				Number: testutils.AsPtr(123),
				Title:  testutils.AsPtr("bug: this is a bug"),
				State:  testutils.AsPtr("open"),
			},

			expected: &github.Issue{
				Number: testutils.AsPtr(123),
				Title:  testutils.AsPtr("bug: this is a bug"),
				State:  testutils.AsPtr("open"),
			},
		},
		"cached w/ remote": {
			issueNumber: 123,
			owner:       "ianlewis",
			repo:        "todos",
			cached: &github.Issue{
				Number: testutils.AsPtr(123),
				Title:  testutils.AsPtr("bug: this is a bug"),
				State:  testutils.AsPtr("open"),
			},
			remote: &github.Issue{
				Number: testutils.AsPtr(123),
				Title:  testutils.AsPtr("bug: different title"),
				State:  testutils.AsPtr("closed"),
			},
			expected: &github.Issue{
				Number: testutils.AsPtr(123),
				Title:  testutils.AsPtr("bug: this is a bug"),
				State:  testutils.AsPtr("open"),
			},
		},
		"not cached": {
			issueNumber: 123,
			owner:       "ianlewis",
			repo:        "todos",
			remote: &github.Issue{
				Number: testutils.AsPtr(123),
				Title:  testutils.AsPtr("bug: this is a bug"),
				State:  testutils.AsPtr("open"),
			},

			expected: &github.Issue{
				Number: testutils.AsPtr(123),
				Title:  testutils.AsPtr("bug: this is a bug"),
				State:  testutils.AsPtr("open"),
			},
		},
		"not cached no remote": {
			issueNumber: 123,
			owner:       "ianlewis",
			repo:        "todos",
			err:         ErrAPI,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r := New(context.Background(), &Options{
				RepoOwner: tc.owner,
				RepoName:  tc.repo,
				client: mock.NewMockedHTTPClient(
					mock.WithRequestMatchHandler(
						mock.GetReposIssuesByOwnerByRepoByIssueNumber,
						http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							if tc.remote != nil {
								pathMatch := regexp.MustCompile("/repos/([a-zA-Z0-9]+)/([a-zA-Z0-9]+)/issues/([0-9]+)")
								parts := pathMatch.FindStringSubmatch(r.URL.Path)
								if len(parts) != 4 {
									http.NotFound(w, r)
									return
								}
								repoOwner := parts[1]
								repoName := parts[2]
								issueNumber, err := strconv.Atoi(parts[3])
								if err != nil {
									http.NotFound(w, r)
									return
								}

								if repoOwner == tc.owner && repoName == tc.repo && issueNumber == *tc.remote.Number {
									_ = testutils.Must(w.Write(mock.MustMarshal(tc.remote)))
									return
								}
							}

							http.NotFound(w, r)
						}),
					),
				),
			})

			if tc.cached != nil {
				r.issues.cache = map[int]*github.Issue{
					*tc.cached.Number: tc.cached,
				}
			}

			issue, err := r.loadIssue(context.Background(), tc.issueNumber)
			if diff := cmp.Diff(tc.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("unexpected error (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.expected, issue); diff != "" {
				t.Fatalf("unexpected result (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.expected, r.issues.cache[tc.issueNumber]); diff != "" {
				t.Fatalf("unexpected cached result (-want +got):\n%s", diff)
			}
		})
	}
}

func parseAndGetIssue(owner, repo string, issues []*github.Issue, r *http.Request) *github.Issue {
	pathMatch := regexp.MustCompile("/repos/([a-zA-Z0-9]+)/([a-zA-Z0-9]+)/issues/([0-9]+)")
	parts := pathMatch.FindStringSubmatch(r.URL.Path)
	if len(parts) != 4 {
		return nil
	}

	repoOwner := parts[1]
	repoName := parts[2]
	issueNumber, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil
	}

	if repoOwner != owner || repoName != repo {
		return nil
	}

	for i := range issues {
		if *issues[i].Number == issueNumber {
			return issues[i]
		}
	}

	return nil
}

//nolint:paralleltest // uses Chdir and cannot be run in parallel.
func Test_ReopenAll(t *testing.T) {
	testCases := map[string]struct {
		owner  string
		repo   string
		files  []*testutils.File
		dryRun bool
		issues []*github.Issue

		reopenErr  bool
		commentErr bool

		expectedIDs    map[int]bool
		expectedResult bool
	}{
		"reopen issue": {
			owner: "ianlewis",
			repo:  "todos",
			files: []*testutils.File{
				{
					Path: "code.go",
					Contents: []byte(`package foo
					// package comment

					// TODO is a function.
					// TODO(#321): some task.
					func TODO() {
						return // Random comment
					}`),
					Mode: 0o600,
				},
			},
			issues: []*github.Issue{
				{
					Number: testutils.AsPtr(321),
					State:  testutils.AsPtr("closed"),
					Title:  testutils.AsPtr("Test Issue"),
				},
			},
			expectedIDs:    map[int]bool{321: true},
			expectedResult: false,
		},
		"reopen issue dry-run": {
			owner:  "ianlewis",
			repo:   "todos",
			dryRun: true,
			files: []*testutils.File{
				{
					Path: "code.go",
					Contents: []byte(`package foo
					// package comment

					// TODO is a function.
					// TODO(#321): some task.
					func TODO() {
						return // Random comment
					}`),
					Mode: 0o600,
				},
			},
			issues: []*github.Issue{
				{
					Number: testutils.AsPtr(321),
					State:  testutils.AsPtr("closed"),
					Title:  testutils.AsPtr("Test Issue"),
				},
			},
			expectedIDs:    nil,
			expectedResult: false,
		},
		"already open issue": {
			owner: "ianlewis",
			repo:  "todos",
			files: []*testutils.File{
				{
					Path: "code.go",
					Contents: []byte(`package foo
					// package comment

					// TODO is a function.
					// TODO(#321): some task.
					func TODO() {
						return // Random comment
					}`),
					Mode: 0o600,
				},
			},
			issues: []*github.Issue{
				{
					Number: testutils.AsPtr(321),
					State:  testutils.AsPtr("open"),
					Title:  testutils.AsPtr("Test Issue"),
				},
			},
			expectedIDs:    nil,
			expectedResult: false,
		},
		"issue not exists": {
			owner: "ianlewis",
			repo:  "todos",
			files: []*testutils.File{
				{
					Path: "code.go",
					Contents: []byte(`package foo
					// package comment

					// TODO is a function.
					// TODO(#321): some task.
					func TODO() {
						return // Random comment
					}`),
					Mode: 0o600,
				},
			},
			issues:         nil,
			expectedIDs:    nil,
			expectedResult: true,
		},
		"reopen error": {
			owner: "ianlewis",
			repo:  "todos",
			files: []*testutils.File{
				{
					Path: "code.go",
					Contents: []byte(`package foo
					// package comment

					// TODO is a function.
					// TODO(#321): some task.
					func TODO() {
						return // Random comment
					}`),
					Mode: 0o600,
				},
			},
			reopenErr: true,
			issues: []*github.Issue{
				{
					Number: testutils.AsPtr(321),
					State:  testutils.AsPtr("closed"),
					Title:  testutils.AsPtr("Test Issue"),
				},
			},
			expectedIDs:    nil,
			expectedResult: true,
		},
		"comment error": {
			owner: "ianlewis",
			repo:  "todos",
			files: []*testutils.File{
				{
					Path: "code.go",
					Contents: []byte(`package foo
					// package comment

					// TODO is a function.
					// TODO(#321): some task.
					func TODO() {
						return // Random comment
					}`),
					Mode: 0o600,
				},
			},
			commentErr: true,
			issues: []*github.Issue{
				{
					Number: testutils.AsPtr(321),
					State:  testutils.AsPtr("closed"),
					Title:  testutils.AsPtr("Test Issue"),
				},
			},
			// NOTE: the issue will still be reopened.
			expectedIDs:    map[int]bool{321: true},
			expectedResult: true,
		},
	}

	for name, tc := range testCases {
		tc := tc

		//nolint:paralleltest // uses Chdir and cannot be run in parallel.
		t.Run(name, func(t *testing.T) {
			// Use TempDir to set up directory.
			wd := testutils.Must(os.Getwd())
			dir := testutils.NewTempDir(tc.files)
			defer dir.Cleanup()
			testutils.Check(os.Chdir(dir.Dir()))
			defer func() {
				testutils.Check(os.Chdir(wd))
			}()

			var reopenedIDs map[int]bool
			r := New(context.Background(), &Options{
				DryRun:    tc.dryRun,
				RepoOwner: tc.owner,
				RepoName:  tc.repo,
				Paths:     []string{"."},
				client: mock.NewMockedHTTPClient(
					mock.WithRequestMatchHandler(
						mock.GetReposIssuesByOwnerByRepoByIssueNumber,
						http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							issue := parseAndGetIssue(tc.owner, tc.repo, tc.issues, r)
							if issue == nil {
								http.NotFound(w, r)
								return
							}

							b, err := json.Marshal(issue)
							if err != nil {
								http.Error(w, fmt.Sprintf("%s: %v", http.StatusText(http.StatusInternalServerError), err), http.StatusInternalServerError)
								return
							}

							w.Write(b)
						}),
					),
					mock.WithRequestMatchHandler(
						mock.PatchReposIssuesByOwnerByRepoByIssueNumber,
						http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							if tc.reopenErr {
								http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
								return
							}

							issue := parseAndGetIssue(tc.owner, tc.repo, tc.issues, r)
							if issue == nil {
								http.NotFound(w, r)
								return
							}

							if reopenedIDs == nil {
								reopenedIDs = make(map[int]bool)
							}
							reopenedIDs[*issue.Number] = true
						}),
					),
					mock.WithRequestMatchHandler(
						mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
						http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							if tc.commentErr {
								http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
								return
							}

							issue := parseAndGetIssue(tc.owner, tc.repo, tc.issues, r)
							if issue == nil {
								http.NotFound(w, r)
								return
							}

							if reopenedIDs == nil {
								reopenedIDs = make(map[int]bool)
							}
							reopenedIDs[*issue.Number] = true
						}),
					),
				),
			})

			if got, want := r.ReopenAll(context.Background()), tc.expectedResult; got != want {
				t.Errorf("unexpected result, got: %v, want: %v", got, want)
			}

			got, want := reopenedIDs, tc.expectedIDs
			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatalf("unexpected reopened issues (-want +got):\n%s", diff)
			}
		})
	}
}

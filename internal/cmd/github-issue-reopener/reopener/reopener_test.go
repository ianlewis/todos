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
	"net/http"
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

func matchEqual(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}

	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}

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

			if got, want := labelMatch.FindStringSubmatch(tc.label), tc.match; !matchEqual(got, want) {
				t.Fatalf("unexpected match (-want +got):\n%s", cmp.Diff(want, got))
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
			})
			r.client = github.NewClient(mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.GetReposIssuesByOwnerByRepoByIssueNumber,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if tc.remote != nil {
							pathMatch := regexp.MustCompile("/repos/([a-zA-Z0-9]+)/([a-zA-Z0-9]+)/issues/([0-9]+)")
							parts := pathMatch.FindStringSubmatch(r.URL.Path)
							if len(parts) != 4 {
								t.Fatalf("not enough match parts for path %q: %q", r.URL.Path, parts)
							}
							repoOwner := parts[1]
							repoName := parts[2]
							issueNumber, err := strconv.Atoi(parts[3])
							if err != nil {
								t.Fatalf("bad issue number %q: %v", parts[3], err)
							}

							if repoOwner == tc.owner && repoName == tc.repo && issueNumber == *tc.remote.Number {
								_ = testutils.Must(w.Write(mock.MustMarshal(tc.remote)))
								return
							}
						}

						http.NotFound(w, r)
					}),
				),
			))

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

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
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/google/go-github/v52/github"
	"github.com/ianlewis/todos/internal/todos"
)

// labelMatch is the regexp that matches the TODO label. It can be of the
// following forms:
// - http://github.com/owner/repo/issues/1234
// - github.com/owner/repo/issues/1234
// - #1234
// - 1234
var labelMatch = regexp.MustCompile("((https?://)?github.com/(.*)/(.*)/issues/|#?)([0-9]+)")

// IssuesRef is a set of references to an issue.
type IssuesRef struct {
	// ID is the issue ID.
	ID int

	// TODOs is a list of todos referencing this issue.
	TODOs []*todos.TODO
}

// IssueReopener is reopens issues referenced by TODOs.
type IssueReopener struct {
	client *github.Client
	owner  string
	repo   string

	// issues is a local cache of issues.
	issues struct {
		sync.Mutex
		cache map[int]*github.Issue
	}
}

// New returns a new IssueReopener
func New(ctx context.Context, owner, repo, token string) (*IssueReopener, error) {
	var client *github.Client
	if token == "" {
		client = github.NewClient(nil)
	} else {
		client = github.NewTokenClient(ctx, token)
	}

	return &IssueReopener{
		client: client,
		owner:  owner,
		repo:   repo,
	}, nil
}

// Reopen the issue if the todo references an issue on the repo. If not this is a no-op.
func (r *IssueReopener) Reopen(ctx context.Context, fileName string, todo *todos.TODO) error {
	match := labelMatch.FindStringSubmatch(todo.Label)
	if len(match) == 0 {
		return nil
	}

	owner := match[2]
	repo := match[3]
	if owner != r.owner || repo != r.repo {
		return nil
	}

	id, err := strconv.Atoi(match[4])
	if err != nil {
		// issue is not a number.
		return nil
	}

	issue, err := r.loadIssue(ctx, id)
	if err != nil {
		return fmt.Errorf("loading issue: %w", err)
	}
	if issue.State == nil || *issue.State == "open" {
		return nil
	}

	comment := strings.Builder{}
	fmt.Fprintln(&comment, "There are TODOs still referencing this issue:")
	fmt.Fprintf(&comment,
		"1. [%s:%d](https://github.com/%s/%s/blob/HEAD/%s#L%d): %s\n",
		l.File, l.Line, b.owner, b.repo, l.File, l.Line, l.Comment)
	fmt.Fprintf(&comment,
		"\n\nSearch [TODO](https://github.com/%s/%s/search?q=%%22%s%%22)", b.owner, b.repo, todo.Issue)

	return nil
}

// loadIssue gets the issue by number.
func (r *IssueReopener) loadIssue(ctx context.Context, id int) (*github.Issue, error) {
	r.issues.Lock()
	defer r.issues.Unlock()

	if issue, ok := r.issues.cache[id]; ok {
		return issue, nil
	}

	issue, _, err := Get(ctx, r.owner, r.repo, id)
	if err != nil {
		return issue, err
	}
	r.issues.cache[id] = issue
	return issue, nil
}

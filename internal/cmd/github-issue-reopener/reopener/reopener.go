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
	"github.com/ianlewis/todos/internal/walker"
)

// labelMatch is the regexp that matches the TODO label. It can be of the
// following forms:
// - http://github.com/owner/repo/issues/1234
// - github.com/owner/repo/issues/1234
// - #1234
// - 1234
var labelMatch = regexp.MustCompile("((https?://)?github.com/(.*)/(.*)/issues/|#?)([0-9]+)")

// IssueRef is a set of references to an issue.
type IssueRef struct {
	// ID is the issue ID.
	ID int

	// TODOs is a list of todos referencing this issue.
	TODOs []*walker.TODORef
}

// IssueReopener is reopens issues referenced by TODOs.
type IssueReopener struct {
	client *github.Client
	owner  string
	repo   string

	dryRun bool

	// issues is a local cache of issues.
	issues struct {
		sync.Mutex
		cache map[int]*github.Issue
	}

	refs struct {
		sync.Mutex
		cache map[int]*IssueRef
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

// handleTODO implements walker.TODOHandler
func (r *IssueReopener) handleTODO(ref *walker.TODORef) error {
	match := labelMatch.FindStringSubmatch(ref.TODO.Label)
	if len(match) == 0 {
		return nil
	}

	// Check if the URL matches the owner and repo name.
	if match[2] != r.owner || match[3] != r.repo {
		return nil
	}

	id, err := strconv.Atoi(match[4])
	if err != nil {
		// issue is not a number.
		return nil
	}

	_ = r.addToRef(id, ref)
	return nil
}

// ReopenAll reopens issues for all encountered todos and returns the last error encountered.
func (r *IssueReopener) ReopenAll(ctx context.Context) error {
	r.refs.Lock()
	defer r.refs.Unlock()

	for id, issueRef := range r.refs.cache {
		issue, err := r.loadIssue(ctx, id)
		if err != nil {
			return fmt.Errorf("loading issue: %w", err)
		}
		if issue.State == nil || *issue.State == "open" {
			// The issue is still open. Do nothing.
			return nil
		}

		if r.dryRun {
			fmt.Printf("[dry-run] Reopening https://github.com/%s/%s/issues/%d\n", r.owner, r.repo, id)
			return nil
		}

		fmt.Printf("Reopening https://github.com/%s/%s/issues/%d\n", r.owner, r.repo, id)

		req := &github.IssueRequest{State: github.String("open")}
		_, _, err = r.client.Issues.Edit(ctx, r.owner, r.repo, id, req)
		if err != nil {
			return fmt.Errorf("reopening issue %d: %v", id, err)
		}

		// TODO: Add label to reopened issue.

		comment := strings.Builder{}
		fmt.Fprintln(&comment, "There are TODOs referencing this issue:")
		for i, todoRef := range issueRef.TODOs {
			fmt.Fprintf(&comment,
				"%d. [%s:%d](https://github.com/%s/%s/blob/HEAD/%s#L%d): %s\n",
				i, todoRef.FileName, todoRef.TODO.Line, r.owner, r.repo, todoRef.FileName, todoRef.TODO.Line, todoRef.TODO.Message)
		}

		cmt := &github.IssueComment{
			Body:      github.String(comment.String()),
			Reactions: &github.Reactions{Confused: github.Int(1)},
		}
		if _, _, err := r.client.Issues.CreateComment(ctx, r.owner, r.repo, id, cmt); err != nil {
			return fmt.Errorf("failed to add comment to issue %d: %v", id, err)
		}
	}

	return nil
}

// loadIssue gets the issue by number.
func (r *IssueReopener) loadIssue(ctx context.Context, id int) (*github.Issue, error) {
	r.issues.Lock()
	defer r.issues.Unlock()

	if issue, ok := r.issues.cache[id]; ok {
		return issue, nil
	}

	issue, _, err := r.client.Issues.Get(ctx, r.owner, r.repo, id)
	if err != nil {
		return issue, err
	}
	r.issues.cache[id] = issue
	return issue, nil
}

// addToRef gets the IssueRef and adds the todo to it.
func (r *IssueReopener) addToRef(id int, todo *walker.TODORef) *IssueRef {
	r.refs.Lock()
	defer r.refs.Unlock()

	ref, ok := r.refs.cache[id]
	if !ok {
		ref := &IssueRef{
			ID:    id,
			TODOs: []*walker.TODORef{todo},
		}
		r.refs.cache[id] = ref
	} else {
		ref.TODOs = append(ref.TODOs, todo)
	}

	return ref
}

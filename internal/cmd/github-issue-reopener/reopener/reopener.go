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
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/google/go-github/v52/github"

	"github.com/ianlewis/todos/internal/cmd/github-issue-reopener/util"
	"github.com/ianlewis/todos/internal/walker"
)

// labelMatch is the regexp that matches the TODO label. It can be of the
// following forms:
// - http://github.com/owner/repo/issues/1234
// - github.com/owner/repo/issues/1234
// - #1234
// - 1234
//
// TODO(#100): Support vanity issue urls (e.g. golang.org/issues/123).
var labelMatch = regexp.MustCompile(`^\s*((https?://)?github.com/(.*)/(.*)/issues/|#?)([0-9]+)\s*$`)

// ErrAPI is an error that occurs with the GitHub API.
var ErrAPI = errors.New("GitHub API error")

// IssueRef is a set of references to an issue.
type IssueRef struct {
	// Number is the issue number.
	Number int

	// TODOs is a list of todos referencing this issue.
	TODOs []*walker.TODORef
}

// Options is options for the IssueReopener.
type Options struct {
	// DryRun indicates that changes will only be printed and not actually executed.
	DryRun bool

	// Paths are the paths to walk to look for TODOs to reopen.
	Paths []string

	// RepoOwner is the repository owner.
	RepoOwner string

	// RepoName is the repository name.
	RepoName string

	// Sha of the current checkout.
	Sha string

	// Token is the GitHub Token.
	Token string
}

// IssueReopener is reopens issues referenced by TODOs.
type IssueReopener struct {
	client *github.Client

	options *Options

	// issues is a local cache of issues.
	issues struct {
		sync.Mutex
		cache map[int]*github.Issue
	}

	refs struct {
		sync.Mutex
		cache map[int]*IssueRef
	}

	// err is the last issue encountered.
	err error
}

// New returns a new IssueReopener.
func New(ctx context.Context, opts *Options) *IssueReopener {
	var client *github.Client
	if opts.Token == "" {
		client = github.NewClient(nil)
	} else {
		client = github.NewTokenClient(ctx, opts.Token)
	}

	return &IssueReopener{
		client:  client,
		options: opts,
	}
}

// ReopenAll scans the given paths for TODOs and reopens any closed issues it
// finds. It returns the last error encountered.
func (r *IssueReopener) ReopenAll(ctx context.Context) bool {
	w := walker.New(&walker.Options{
		TODOFunc:  r.handleTODO,
		ErrorFunc: r.handleErr,
		// TODO: Support TODO config.
		Config: nil,

		// TODO: Support walker config.
		IncludeHidden:   false,
		IncludeVendored: false,
		Paths:           r.options.Paths,
	})

	// NOTE: even if we encounter errors when walking the directory tree we
	// will still continue on and try to reopen any issues we did find.
	_ = w.Walk()

	r.reopenAll(ctx)

	return r.err != nil
}

// handleTODO implements walker.TODOHandler.
func (r *IssueReopener) handleTODO(ref *walker.TODORef) error {
	match := labelMatch.FindStringSubmatch(ref.TODO.Label)
	if len(match) == 0 {
		return nil
	}

	// Check if the URL matches the owner and repo name.
	if (match[3] != "" || match[4] != "") && (match[3] != r.options.RepoOwner || match[4] != r.options.RepoName) {
		return nil
	}

	number, err := strconv.Atoi(match[5])
	if err != nil {
		fmt.Printf("bad issue number %q: %v", match[5], err)
		// issue is not a number.
		return nil
	}

	_ = r.addToRef(number, ref)
	return nil
}

// handleErr implements walker.ErrorHandler.
func (r *IssueReopener) handleErr(err error) error {
	r.err = err
	fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return nil
}

// reopenAll reopens issues for all encountered todos.
func (r *IssueReopener) reopenAll(ctx context.Context) {
	r.refs.Lock()
	defer r.refs.Unlock()

	for id, issueRef := range r.refs.cache {
		issue, err := r.loadIssue(ctx, id)
		if err != nil {
			err = fmt.Errorf("loading issue %d: %w", id, err)
			if herr := r.handleErr(err); herr != nil {
				return
			}
			continue
		}
		if util.MustString(issue.State) == "open" {
			// The issue is still open. Do nothing.
			continue
		}

		if r.options.DryRun {
			fmt.Printf(
				"[dry-run] Reopening https://github.com/%s/%s/issues/%d : %s\n",
				r.options.RepoOwner,
				r.options.RepoName,
				id,
				util.MustString(issue.Title),
			)
			continue
		}

		fmt.Printf("Reopening https://github.com/%s/%s/issues/%d\n", r.options.RepoOwner, r.options.RepoName, id)

		req := &github.IssueRequest{State: github.String("open")}
		_, _, err = r.client.Issues.Edit(ctx, r.options.RepoOwner, r.options.RepoName, id, req)
		if err != nil {
			err = fmt.Errorf("%w: %w", ErrAPI, err)
			if herr := r.handleErr(err); herr != nil {
				return
			}
			continue
		}

		comment := strings.Builder{}
		fmt.Fprintln(&comment, "There are TODOs referencing this issue:")
		for i, todoRef := range issueRef.TODOs {
			fmt.Fprintf(
				&comment,
				"%d. [%s:%d](https://github.com/%s/%s/blob/%s/%s#L%d): %s\n",
				i+1,
				todoRef.FileName,
				todoRef.TODO.Line,
				r.options.RepoOwner,
				r.options.RepoName,
				r.options.Sha,
				todoRef.FileName,
				todoRef.TODO.Line,
				todoRef.TODO.Message,
			)
		}

		cmt := &github.IssueComment{
			Body:      github.String(comment.String()),
			Reactions: &github.Reactions{Confused: github.Int(1)},
		}
		if _, _, err := r.client.Issues.CreateComment(ctx, r.options.RepoOwner, r.options.RepoName, id, cmt); err != nil {
			err = fmt.Errorf("%w: %w", ErrAPI, err)
			if herr := r.handleErr(err); herr != nil {
				return
			}
			continue
		}
	}
}

// loadIssue gets the issue by number.
func (r *IssueReopener) loadIssue(ctx context.Context, number int) (*github.Issue, error) {
	r.issues.Lock()
	defer r.issues.Unlock()

	if r.issues.cache == nil {
		r.issues.cache = make(map[int]*github.Issue)
	}

	if issue, ok := r.issues.cache[number]; ok {
		return issue, nil
	}

	issue, _, err := r.client.Issues.Get(ctx, r.options.RepoOwner, r.options.RepoName, number)
	if err != nil {
		return issue, fmt.Errorf("%w: %w", ErrAPI, err)
	}
	r.issues.cache[number] = issue
	return issue, nil
}

// addToRef gets the IssueRef and adds the todo to it.
func (r *IssueReopener) addToRef(number int, todo *walker.TODORef) *IssueRef {
	r.refs.Lock()
	defer r.refs.Unlock()

	if r.refs.cache == nil {
		r.refs.cache = make(map[int]*IssueRef)
	}

	ref, ok := r.refs.cache[number]
	if !ok {
		ref = &IssueRef{
			Number: number,
			TODOs:  []*walker.TODORef{todo},
		}
		r.refs.cache[number] = ref
	} else {
		ref.TODOs = append(ref.TODOs, todo)
	}

	return ref
}

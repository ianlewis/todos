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

	"github.com/google/go-github/v52/github"
	"github.com/ianlewis/todos/internal/todos"
)

// IssueReopener is reopens issues referenced by TODOs.
type IssueReopener struct {
	client *github.Client
}

// New returns a new IssueReopener
func New(ctx context.Context, token string) (*IssueReopener, error) {
	var client *github.Client
	if token == "" {
		client = github.NewClient(nil)
	} else {
		client = github.NewTokenClient(ctx, token)
	}

	return &IssueReopener{
		client: client,
	}, nil
}

// Reopen the issue if the todo references an issue on the repo. If not this is a no-op.
func (r *IssueReopener) Reopen(todo *todos.TODO) error {
}

func parseIssue(todo *todos.TODO)

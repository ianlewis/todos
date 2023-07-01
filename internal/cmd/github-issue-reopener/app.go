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

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
	"sigs.k8s.io/release-utils/version"

	"github.com/ianlewis/todos/internal/cmd/github-issue-reopener/reopener"
	"github.com/ianlewis/todos/internal/cmd/github-issue-reopener/util"
)

const (
	// ExitCodeSuccess is successful error code.
	ExitCodeSuccess int = iota

	// ExitCodeFlagParseError is the exit code for a flag parsing error.
	ExitCodeFlagParseError

	// ExitCodeReopenError is the exit code for an error encountered when reopening issues.
	ExitCodeReopenError

	// ExitCodeUnknownError is the exit code for an unknown error.
	ExitCodeUnknownError
)

var (
	// ErrReopen is an error encountered when reopening issues.
	ErrReopen = errors.New("reopening issues")

	// ErrFlagParse is a flag parsing error.
	ErrFlagParse = errors.New("parsing flags")
)

func newGitHubIssueReopenerApp() *cli.App {
	return &cli.App{
		Name:  filepath.Base(os.Args[0]),
		Usage: "Reopen GitHub issues that are still referenced by TODOs.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:               "dry-run",
				Usage:              "Perform a dry-run. Don't take any action.",
				Value:              false,
				DisableDefaultText: true,
			},
			&cli.StringFlag{
				Name:      "token-file",
				Usage:     "File containing the GitHub token. If not provided GH_TOKEN or GITHUB_TOKEN is used.",
				TakesFile: true,
			},
			&cli.DurationFlag{
				Name:        "timeout",
				Usage:       "Timeout for the scanning the code. 0 for no timeout.",
				DefaultText: "no timeout",
				Value:       0,
			},
			&cli.StringFlag{
				Name:    "repo",
				Usage:   "The GitHub repository of the form <owner>/<name> (defaults to GITHUB_REPOSITORY)",
				Aliases: []string{"R"},
			},
			&cli.StringFlag{
				Name:  "sha",
				Usage: "The SHA digest of the current checkout (defaults to GITHUB_SHA)",
			},
			&cli.BoolFlag{
				Name:               "version",
				Usage:              "Print the version and exit",
				Aliases:            []string{"v"},
				DisableDefaultText: true,
			},
		},
		ArgsUsage:       "[PATH]...",
		Copyright:       "Google LLC",
		HideHelpCommand: true,
		Action: func(c *cli.Context) error {
			if c.Bool("version") {
				versionInfo := version.GetVersionInfo()
				fmt.Fprintf(c.App.Writer, `%s %s
Copyright (c) Google LLC

%s`, c.App.Name, versionInfo.GitVersion, versionInfo.String())
				return nil
			}

			ctx := context.Background()
			timeout := c.Duration("timeout")
			if timeout != 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}
			opts, err := reopenerOptionsFromContext(c)
			if err != nil {
				return err
			}
			r := reopener.New(ctx, opts)
			if r.ReopenAll(ctx) {
				return ErrReopen
			}

			return nil
		},
		ExitErrHandler: func(c *cli.Context, err error) {
		},
	}
}

var gitShaMatch = regexp.MustCompile(`^[0-9a-f]{7,40}$`)

func reopenerOptionsFromContext(c *cli.Context) (*reopener.Options, error) {
	o := reopener.Options{}

	// DryRun
	o.DryRun = c.Bool("dry-run")

	// Paths
	o.Paths = c.Args().Slice()
	if len(o.Paths) == 0 {
		o.Paths = []string{"."}
	}

	// RepoName
	// RepoOwner
	repo := c.String("repo")
	if repo == "" {
		repo = os.Getenv("GITHUB_REPOSITORY")
	}
	if parts := strings.SplitN(repo, "/", 2); len(parts) == 2 {
		o.RepoOwner = parts[0]
		o.RepoName = parts[1]
	} else {
		return nil, fmt.Errorf("%w: invalid repo: %q", ErrFlagParse, repo)
	}

	// Sha
	o.Sha = c.String("sha")
	if o.Sha == "" {
		o.Sha = os.Getenv("GITHUB_SHA")
	}
	if !gitShaMatch.MatchString(o.Sha) {
		return nil, fmt.Errorf("%w: invalid git digest", ErrFlagParse)
	}

	// Token
	o.Token = util.FirstString(os.Getenv("GH_TOKEN"), os.Getenv("GITHUB_TOKEN"))
	tokenFile := c.String("token-file")
	if tokenFile != "" {
		bytes, err := os.ReadFile(tokenFile)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFlagParse, err)
		}
		o.Token = string(bytes)
	}

	return &o, nil
}

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
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"sigs.k8s.io/release-utils/version"
)

var gitShaMatch = regexp.MustCompile(`^[0-9a-f]{7,40}$`)

// ErrFlagParse is a flag parsing error.
var ErrFlagParse = errors.New("parsing flags")

// Options are the command line options.
type Options struct {
	// DryRun indicates that changes will only be printed and not actually executed.
	DryRun bool

	// RepoOwner is the repository owner.
	RepoOwner string

	// RepoName is the repository name.
	RepoName string

	// Sha of the current checkout.
	Sha string

	// Token is the GitHub Token.
	Token string

	// Version indicates the command should print version info and exit.
	Version bool

	// Help indicates the command should print the help and exit.
	Help bool

	// Paths are the paths to walk to look for TODOs to revive.
	Paths []string
}

// New parses the given command-line args and returns a new options instance.
func New(args []string) (*Options, error) {
	var o Options

	// Set defaults from the environment.
	repo := os.Getenv("GITHUB_REPOSITORY")
	o.Sha = os.Getenv("GITHUB_SHA")

	var tokenFile string

	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.BoolVar(&o.Help, "help", false, "print help and exit")
	fs.BoolVar(&o.Help, "h", false, "print help and exit")
	fs.BoolVar(&o.Version, "version", false, "print version information and exit")
	fs.StringVar(&repo, "repo", repo, "The GitHub repository of the form <owner>/<name>")
	fs.StringVar(&o.Sha, "sha", o.Sha, "The SHA digest of the current checkout")
	fs.StringVar(&tokenFile, "token-file", "", "File containing the GitHub token")
	fs.BoolVar(&o.DryRun, "dry-run", false, "Perform a dry-run. Don't take any action.")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [OPTION]... [PATH]...\n", args[0])
		fmt.Fprintf(fs.Output(), "Try '%s --help' for more information.\n", args[0])
	}

	if err := fs.Parse(args[1:]); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFlagParse, err)
	}

	if parts := strings.SplitN(repo, "/", 2); len(parts) == 2 {
		o.RepoOwner = parts[0]
		o.RepoName = parts[1]
	} else {
		return nil, fmt.Errorf("%w: invalid repo: %q", ErrFlagParse, repo)
	}

	// Validate the git sha
	if !gitShaMatch.MatchString(o.Sha) {
		return nil, fmt.Errorf("%w: invalid git digest", ErrFlagParse)
	}

	o.Token = os.Getenv("GITHUB_TOKEN")
	if tokenFile != "" {
		bytes, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFlagParse, err)
		}
		o.Token = string(bytes)
	}

	o.Paths = fs.Args()
	if len(o.Paths) == 0 {
		o.Paths = []string{"."}
	}

	return &o, nil
}

// PrintLongUsage prints the long help for the options.
func (o *Options) PrintLongUsage() {
	fmt.Fprintf(os.Stdout, `Usage: %s [OPTION]... [PATH]...
Reopen GitHub issues that are still referenced by TODOs.

OPTIONS:
  -h, --help                  Print help and exit.
  --repo=OWNER/REPO           GitHub Repository. Defaults to GITHUB_REPOSITORY.
  --sha=SHA1                  Git digest of current checkout. Defaults to GITHUB_SHA.
  --dry-run                   Perform a dry-run. Don't take any action.
  --token-file=FILE           File containing the GitHub token.
  --version                   Print version information and exit.
`, os.Args[0])
}

// PrintVersion prints version information.
func (o *Options) PrintVersion() {
	versionInfo := version.GetVersionInfo()

	fmt.Fprintf(os.Stdout, `%s %s
Copyright (c) Google LLC
License Apache License Version 2.0

%s`, os.Args[0], versionInfo.GitVersion, versionInfo.String())
}

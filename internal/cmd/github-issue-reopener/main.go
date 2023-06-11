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

	"github.com/ianlewis/todos/internal/cmd/github-issue-reopener/options"
	"github.com/ianlewis/todos/internal/cmd/github-issue-reopener/reopener"
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

// ErrReopen is an error encountered when reopening issues.
var ErrReopen = errors.New("reopening issues")

// PrintError prints an error for the command.
func PrintError(cmd string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", cmd, err)
}

// ExitCode return an exit code for the given error.
func ExitCode(err error) int {
	switch {
	case errors.Is(err, options.ErrFlagParse):
		return ExitCodeFlagParseError
	case errors.Is(err, ErrReopen):
		return ExitCodeReopenError
	case err != nil:
		return ExitCodeUnknownError
	default:
		return ExitCodeSuccess
	}
}

// Exit prints the error and exits the application with the right exit code.
func Exit(err error) {
	// NOTE: Reopen errors return an exit code but do not print the error as it
	// has presumably already been handled.
	if err != nil && !errors.Is(err, ErrReopen) {
		PrintError(os.Args[0], err)
	}
	os.Exit(ExitCode(err))
}

// Run runs the the command.
func Run(opts *options.Options) error {
	if opts.Help {
		opts.PrintLongUsage()
		return nil
	}

	if opts.Version {
		opts.PrintVersion()
		return nil
	}

	// TODO: Support timeouts etc.
	ctx := context.Background()
	r := reopener.New(ctx, opts)
	if r.ReopenAll(ctx) {
		return ErrReopen
	}
	return nil
}

func main() {
	opts, err := options.New(os.Args)
	if err != nil {
		Exit(err)
	}
	Exit(Run(opts))
}

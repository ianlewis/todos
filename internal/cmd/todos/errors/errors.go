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

package errors

import (
	"errors"
	"fmt"
	"os"
)

const (
	// ExitCodeSuccess is successful error code.
	ExitCodeSuccess int = iota

	// ExitCodeFlagParseError is the exit code for a flag parsing error.
	ExitCodeFlagParseError

	// ExitCodeWalkError is the exit code for a file walking error.
	ExitCodeWalkError

	// ExitCodeUnknownError is the exit code for an unknown error.
	ExitCodeUnknownError
)

var (
	// ErrFlagParse is a flag parsing error.
	ErrFlagParse = errors.New("parsing flags")

	// ErrWalk is a file recursing error.
	ErrWalk = errors.New("walking")
)

// PrintError prints an error for the command.
func PrintError(cmd string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", cmd, err)
}

// ExitCode return an exit code for the given error.
func ExitCode(err error) int {
	switch {
	case errors.Is(err, ErrFlagParse):
		return ExitCodeFlagParseError
	case errors.Is(err, ErrWalk):
		return ExitCodeWalkError
	case err != nil:
		return ExitCodeUnknownError
	default:
		return ExitCodeSuccess
	}
}

// Exit prints the error and exits the application with the right exit code.
func Exit(err error) {
	// NOTE: Walk errors return an exit code but do not print the error as it
	// has presumably already been handled.
	if err != nil && !errors.Is(err, ErrWalk) {
		PrintError(os.Args[0], err)
	}
	os.Exit(ExitCode(err))
}

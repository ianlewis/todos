// Copyright 2025 Google LLC
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
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/ianlewis/todos/internal/utils"
)

const (
	// ExitCodeSuccess is successful error code when TODOs are not found.
	ExitCodeSuccess int = iota

	// ExitCodeTODOsFound is the exit code when TODOs are found.
	ExitCodeTODOsFound

	// ExitCodeFlagParseError is the exit code for a flag parsing error.
	ExitCodeFlagParseError

	// ExitCodeWalkError is the exit code for a file walking error.
	ExitCodeWalkError

	// ExitCodeUnknownError is the exit code for an unknown error.
	ExitCodeUnknownError
)

var (
	// ErrTODOsFound is the error that occurs when TODOs are found.
	// ErrTODOsFound has no error message.
	ErrTODOsFound = cli.Exit("", ExitCodeTODOsFound)

	// ErrFlagParse is a flag parsing error.
	ErrFlagParse = cli.Exit("parsing flags", ExitCodeFlagParseError)

	// ErrWalk is a file recursing error.
	ErrWalk = cli.Exit("error scanning files", ExitCodeWalkError)

	// ErrUnknown is an unknown error.
	ErrUnknown = cli.Exit("unexpected error", ExitCodeUnknownError)
)

// joinedError is an interface that matches errors that wrap multiple errors
// via [errors.Join].
type joinedError interface {
	Unwrap() []error
}

// Unwrap unwraps errors including those wrapping multiple errors.
func Unwrap(err error) []error {
	unwrapped := errors.Unwrap(err)
	if unwrapped != nil {
		return []error{unwrapped}
	}
	if err, ok := err.(joinedError); ok {
		return err.Unwrap()
	}
	return nil
}

// traverseErr calls func on the given err and any wrapped errors.
func traverseErr(err error, cb func(err error)) {
	if err == nil {
		return
	}

	cb(err)
	for _, err := range Unwrap(err) {
		traverseErr(err, cb)
	}
}

type exitCoder struct {
	msg      string
	wrapped  error
	exitCode int
}

func (err *exitCoder) Error() string {
	return err.msg
}

func (err *exitCoder) Unwrap() error {
	return err.wrapped
}

func (err *exitCoder) ExitCode() int {
	return err.exitCode
}

// wrapExitCoder wraps an ExitCoder with another ExitCoder with different error message.
func wrapExitCoder(msg string, err cli.ExitCoder) cli.ExitCoder {
	return &exitCoder{
		msg:      msg,
		wrapped:  err,
		exitCode: err.ExitCode(),
	}
}

// ExitErrHandler handles error that occur when running the app. It calls
// [cli.HandleExitCoder] but checks for wrapped errors that implement
// [cli.ExitCoder].
func ExitErrHandler(c *cli.Context, err error) {
	if err == nil {
		return
	}

	// NOTE: Walk errors return an exit code but do not print the error as it
	// has presumably already been handled by the walker.
	if errors.Is(err, ErrWalk) {
		cli.OsExiter(ErrWalk.ExitCode())
		return
	}

	traverseErr(err, func(unwrapped error) {
		//nolint:errorlint // errors are already being unwrapped.
		if errExitCoder, ok := unwrapped.(cli.ExitCoder); ok {
			// NOTE: We intentionally use the original err for the error message here.
			var msg string
			if err.Error() != "" {
				msg = fmt.Sprintf("%s: %v", c.App.Name, err)
			}
			// NOTE: cli.HandleExitCoder does nothing if the error is not a
			//       cli.ExitCoder even if it wraps cli.ExitCoder errors. The
			//       error passed to cli.HandleExitCoder itself must be a
			//       cli.ExitCoder.
			cli.HandleExitCoder(wrapExitCoder(msg, errExitCoder))
		}
	})

	_ = utils.Must(fmt.Fprintf(c.App.ErrWriter, "%s: %v\n", c.App.Name, err))
	cli.OsExiter(ErrUnknown.ExitCode())
}

// OnUsageError handles usage errors by wrapping them with [ErrFlagParse].
func OnUsageError(_ *cli.Context, err error, _ bool) error {
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFlagParse, err)
	}
	return nil
}

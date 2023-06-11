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
	"os"

	todoerr "github.com/ianlewis/todos/internal/cmd/todos/errors"
	"github.com/ianlewis/todos/internal/cmd/todos/options"
	"github.com/ianlewis/todos/internal/walker"
)

func main() {
	opts, err := options.New(os.Args)
	if err != nil {
		todoerr.Exit(err)
	}
	todoerr.Exit(Run(opts))
}

// Run runs the the `todos` command.
func Run(opts *options.Options) error {
	if opts.Help {
		opts.PrintLongUsage()
		return nil
	}

	if opts.Version {
		opts.PrintVersion()
		return nil
	}

	w := walker.New(&walker.Options{
		TODOFunc:        opts.Output,
		ErrorFunc:       opts.Error,
		IncludeHidden:   opts.IncludeHidden,
		IncludeVendored: opts.IncludeVendored,
		Paths:           opts.Paths,
	})
	if w.Walk() {
		return todoerr.ErrWalk
	}

	return nil
}

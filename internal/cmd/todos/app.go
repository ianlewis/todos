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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"sigs.k8s.io/release-utils/version"

	"github.com/ianlewis/todos/internal/todos"
	"github.com/ianlewis/todos/internal/walker"
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

// newTODOsApp returns a new `todos` application.
func newTODOsApp() *cli.App {
	return &cli.App{
		Name:  "todos",
		Usage: "Search for TODOS in code.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:               "exclude-hidden",
				Usage:              "Exclude hidden files and directories",
				DisableDefaultText: true,
			},
			&cli.BoolFlag{
				Name:               "include-vendored",
				Usage:              "Include vendored directories",
				DisableDefaultText: true,
			},
			&cli.StringFlag{
				Name:  "todo-types",
				Usage: "Comma separated list of TODO types",
				Value: strings.Join(todos.DefaultTypes, ","),
			},
			&cli.StringFlag{
				Name:    "output",
				Usage:   "Output type (default, github)",
				Value:   "default",
				Aliases: []string{"o"},
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

			opts, err := walkerOptionsFromContext(c)
			if err != nil {
				return err
			}
			w := walker.New(opts)
			if w.Walk() {
				return ErrWalk
			}

			return nil
		},
		ExitErrHandler: func(c *cli.Context, err error) {
			if err == nil {
				return
			}

			// NOTE: Walk errors return an exit code but do not print the error as it
			// has presumably already been handled.
			if errors.Is(err, ErrWalk) {
				cli.OsExiter(ExitCodeWalkError)
				return
			}

			// ExitCode return an exit code for the given error.
			fmt.Fprintf(c.App.ErrWriter, "%s: %v\n", filepath.Base(os.Args[0]), err)
			if errors.Is(err, ErrFlagParse) {
				cli.OsExiter(ExitCodeFlagParseError)
				return
			}

			cli.OsExiter(ExitCodeUnknownError)
		},
	}
}

func outCLI(w io.Writer) walker.TODOHandler {
	return func(o *walker.TODORef) error {
		if o == nil {
			return nil
		}
		fmt.Fprintf(w, "%s%s%s%s%s\n",
			color.MagentaString(o.FileName),
			color.CyanString(":"),
			color.GreenString(fmt.Sprintf("%d", o.TODO.Line)),
			color.CyanString(":"),
			o.TODO.Text,
		)
		return nil
	}
}

func outGithub(w io.Writer) walker.TODOHandler {
	return func(o *walker.TODORef) error {
		if o == nil {
			return nil
		}
		typ := "notice"
		switch o.TODO.Type {
		case "TODO", "HACK", "COMBAK":
			typ = "warning"
		case "FIXME", "XXX", "BUG":
			typ = "error"
		}
		fmt.Fprintf(w, "::%s file=%s,line=%d::%s\n", typ, o.FileName, o.TODO.Line, o.TODO.Text)
		return nil
	}
}

var outTypes = map[string]func(io.Writer) walker.TODOHandler{
	// NOTE: An empty value is treated as the default value.
	"":        outCLI,
	"default": outCLI,
	"github":  outGithub,
}

func walkerOptionsFromContext(c *cli.Context) (*walker.Options, error) {
	o := walker.Options{}

	o.IncludeHidden = !c.Bool("exclude-hidden")

	outType := c.String("output")
	outFunc, ok := outTypes[outType]
	if !ok {
		return nil, fmt.Errorf("%w: invalid output type: %v", ErrFlagParse, outType)
	}

	o.TODOFunc = outFunc(c.App.Writer)
	o.ErrorFunc = func(err error) error {
		fmt.Fprintf(c.App.ErrWriter, "%s: %v\n", c.App.Name, err)
		return nil
	}

	o.Config = &todos.Config{}

	todoTypesStr := c.String("todo-types")
	if todoTypesStr != "" {
		for _, todoType := range strings.Split(todoTypesStr, ",") {
			o.Config.Types = append(o.Config.Types, strings.TrimSpace(todoType))
		}
	}

	o.Paths = c.Args().Slice()
	if len(o.Paths) == 0 {
		o.Paths = []string{"."}
	}

	return &o, nil
}
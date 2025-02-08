// Copyright 2025 Ian Lewis
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

//go:build profile

package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/urfave/cli/v2"
)

func setupProfiling(app *cli.App) {
	app.Flags = append(app.Flags,
		&cli.StringFlag{
			Name:  "profile",
			Usage: "write cpu profile to `FILE`",
			Value: "todos.prof",
		},
	)

	appAction := app.Action
	app.Action = func(c *cli.Context) error {
		profile := c.String("profile")
		if !c.Bool("help") && !c.Bool("version") && profile != "" {
			f, err := os.Create(profile)
			if err != nil {
				return fmt.Errorf("creating prof file: %w", err)
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				return fmt.Errorf("creating prof file: %w", err)
			}
		}
		defer pprof.StopCPUProfile()

		return appAction(c)
	}

	appExit := app.ExitErrHandler
	app.ExitErrHandler = func(c *cli.Context, err error) {
		// Make sure profile is stopped even if os.Exit is called.
		pprof.StopCPUProfile()
		appExit(c, err)
	}
}

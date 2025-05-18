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

// Command todos parses programming language files and prints references to any
// "TODO" comments it finds. It can act as a linter to prevent leaving work
// undone unintentially and can output in various formats.
package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/ianlewis/todos/internal/utils"
)

func main() {
	app := newTODOsApp()

	if err := app.Run(os.Args); err != nil {
		// NOTE: Errors are usually handled in the app itself and os.Exit is
		// called halting the program before we get here. However, Run could
		// return errors in some situations.
		_ = utils.Must(fmt.Fprintf(app.ErrWriter, "%s: %v\n", app.Name, err))

		cli.OsExiter(ErrUnknown.ExitCode())
	}
}

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
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/urfave/cli/v2"

	"github.com/ianlewis/todos/internal/cmd/todos/app"
	"github.com/ianlewis/todos/internal/utils"
)

func main() {
	f, err := os.Create("todos.prof")
	if err != nil {
		_ = utils.Must(fmt.Fprintf(os.Stderr, "%v", err))
		os.Exit(1)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		_ = utils.Must(fmt.Fprintf(os.Stderr, "%v", err))
		os.Exit(1)
	}
	defer pprof.StopCPUProfile()

	exit := cli.OsExiter
	cli.OsExiter = func(code int) {
		// Make sure profile is stopped even if os.Exit is called.
		pprof.StopCPUProfile()
		exit(code)
	}

	app.Main()
}

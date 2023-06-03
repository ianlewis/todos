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
	"strings"

	"github.com/fatih/color"

	"github.com/ianlewis/todos/internal/todos"
)

const (
	exitCodeSuccess int = iota
	exitCodeFlagParseError
	exitCodeWalkError
)

func printError(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], msg)
}

func outReadable(o todoOpt) {
	fmt.Printf("%s%s%s%s%s\n",
		color.MagentaString(o.fileName),
		color.CyanString(":"),
		color.GreenString(fmt.Sprintf("%d", o.todo.Line)),
		color.CyanString(":"),
		o.todo.Text,
	)
}

func outGithub(o todoOpt) {
	typ := "notice"
	switch o.todo.Type {
	case "TODO", "HACK", "COMBAK":
		typ = "warning"
	case "FIXME", "XXX", "BUG":
		typ = "error"
	}
	fmt.Printf("::%s file=%s,line=%d::%s\n", typ, o.fileName, o.todo.Line, o.todo.Text)
}

var outTypes = map[string]lineWriter{
	"default": outReadable,
	"github":  outGithub,
}

func main() {
	opts := &Options{}
	fSet := opts.FlagSet()
	if err := fSet.Parse(os.Args[1:]); err != nil {
		printError(fmt.Sprintf("parsing flags: %v", err))
		os.Exit(exitCodeFlagParseError)
	}

	outFunc, ok := outTypes[opts.Output]
	if !ok {
		printError(fmt.Sprintf("invalid output type: %q", opts.Output))
		os.Exit(exitCodeFlagParseError)
	}

	if opts.Help {
		opts.PrintLongUsage()
		os.Exit(exitCodeSuccess)
	}

	if opts.Version {
		opts.PrintVersion()
		os.Exit(exitCodeSuccess)
	}

	paths := fSet.Args()
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var todoTypes []string
	for _, todoType := range strings.Split(opts.TodoTypes, ",") {
		todoTypes = append(todoTypes, strings.TrimSpace(todoType))
	}

	walker := TODOWalker{
		outFunc: outFunc,
		errFunc: func(err error) {
			printError("%v", err)
		},
		todoConfig: &todos.Config{
			Types: todoTypes,
		},
		includeHidden:   opts.IncludeHidden,
		includeVendored: opts.IncludeVendored,
		includeDocs:     opts.IncludeDocs,
		paths:           paths,
	}

	if walker.Walk() {
		os.Exit(exitCodeWalkError)
	}
}

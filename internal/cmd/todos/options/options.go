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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"sigs.k8s.io/release-utils/version"

	todoerr "github.com/ianlewis/todos/internal/cmd/todos/errors"
	"github.com/ianlewis/todos/internal/todos"
	"github.com/ianlewis/todos/internal/walker"
)

func outReadable(o *walker.TODORef) error {
	fmt.Printf("%s%s%s%s%s\n",
		color.MagentaString(o.FileName),
		color.CyanString(":"),
		color.GreenString(fmt.Sprintf("%d", o.TODO.Line)),
		color.CyanString(":"),
		o.TODO.Text,
	)
	return nil
}

func outGithub(o *walker.TODORef) error {
	typ := "notice"
	switch o.TODO.Type {
	case "TODO", "HACK", "COMBAK":
		typ = "warning"
	case "FIXME", "XXX", "BUG":
		typ = "error"
	}
	fmt.Printf("::%s file=%s,line=%d::%s\n", typ, o.FileName, o.TODO.Line, o.TODO.Text)
	return nil
}

var outTypes = map[string]walker.TODOHandler{
	"default": outReadable,
	"github":  outGithub,
}

// Options are the command line options.
type Options struct {
	// Output writes the output for a TODO.
	Output walker.TODOHandler

	// Error writes error output.
	Error walker.ErrorHandler

	// Types is list of TODO types.
	TODOTypes []string

	// IncludeHidden indicates including hidden files & directories.
	IncludeHidden bool

	// IncludeVendored indicates including vendored directories.
	IncludeVendored bool

	// Version indicates the command should print version info and exit.
	Version bool

	// Help indicates the command should print the help and exit.
	Help bool

	// Paths are the paths to walk to look for TODOs.
	Paths []string
}

// New parses the given command-line args and returns a new options instance.
func New(args []string) (*Options, error) {
	baseCmd := filepath.Base(args[0])

	var outType string
	var todoTypes string
	var excludeHidden bool

	var o Options
	fs := flag.NewFlagSet(baseCmd, flag.ExitOnError)
	fs.BoolVar(&o.Help, "help", false, "print help and exit")
	fs.BoolVar(&o.Help, "h", false, "print help and exit")
	fs.BoolVar(&o.Version, "version", false, "print version information and exit")

	fs.BoolVar(&excludeHidden, "exclude-hidden", false, "exclude hidden files and directories")

	fs.BoolVar(&o.IncludeVendored, "include-vendored", false, "include vendored directories")
	fs.StringVar(&todoTypes, "todo-types", strings.Join(todos.DefaultTypes, ","), "comma separated list of TODO types")
	fs.StringVar(&outType, "o", "default", "output type (default, github)")
	fs.StringVar(&outType, "output", "default", "output type (default, github)")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [OPTION]... [PATH]...\n", baseCmd)
		fmt.Fprintf(fs.Output(), "Try '%s --help' for more information.\n", baseCmd)
	}

	if err := fs.Parse(args[1:]); err != nil {
		return nil, fmt.Errorf("%w: %w", todoerr.ErrFlagParse, err)
	}

	o.IncludeHidden = !excludeHidden

	outFunc, ok := outTypes[outType]
	if !ok {
		return nil, fmt.Errorf("%w: invalid output type: %v", todoerr.ErrFlagParse, o.Output)
	}
	o.Output = outFunc
	o.Error = func(err error) error {
		todoerr.PrintError(baseCmd, err)
		return nil
	}

	for _, todoType := range strings.Split(todoTypes, ",") {
		o.TODOTypes = append(o.TODOTypes, strings.TrimSpace(todoType))
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
Search for TODOS in code.

OPTIONS:
  -h, --help                  Print help and exit.
  --exclude-hidden            Exclude hidden files and directories.
  --include-vendored          Include vendored directories.
  --todo-types=TYPES          Comma separated list of TODO types.
  -o, --output=TYPE           Output type (default, github).
  --version                   Print version information and exit.
`, filepath.Base(os.Args[0]))
}

// PrintVersion prints version information.
func (o *Options) PrintVersion() {
	versionInfo := version.GetVersionInfo()

	fmt.Fprintf(os.Stdout, `%s %s
Copyright (c) Google LLC
License Apache License Version 2.0

%s`, filepath.Base(os.Args[0]), versionInfo.GitVersion, versionInfo.String())
}

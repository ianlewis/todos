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
	"flag"
	"fmt"
	"os"

	"sigs.k8s.io/release-utils/version"
)

var (
	unknown = "unknown"

	// Version is the current application version.
	Version = "devel"

	// Commit is the git commit sha.
	Commit = unknown

	// CommitDate is the date of the commit.
	CommitDate = unknown

	// TreeState is the state of the code tree when built.
	TreeState = unknown
)

// Options are the command line options.
type Options struct {
	// IncludeHidden indicates including hidden files & directories.
	IncludeHidden bool

	// IncludeDocs indicates including documentation files.
	IncludeDocs bool

	// IncludeVendored indicates including vendored directories.
	IncludeVendored bool

	// Output is the output type. Valid values are "default" or "github".
	Output string

	// Version indicates the command should print verison info and exit.
	Version bool

	// Help indicates the command should print the help and exit.
	Help bool
}

// FlagSet returns a FlagSet for the options.
func (o *Options) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("todos", flag.ExitOnError)
	fs.BoolVar(&o.Help, "help", false, "print help and exit")
	fs.BoolVar(&o.Help, "h", false, "print help and exit")
	fs.BoolVar(&o.Version, "version", false, "print version information and exit")
	fs.BoolVar(&o.IncludeHidden, "include-hidden", false, "include hidden files and directories")
	fs.BoolVar(&o.IncludeDocs, "include-docs", false, "include documentation")
	fs.BoolVar(&o.IncludeVendored, "include-vendored", false, "include vendored directories")
	fs.StringVar(&o.Output, "o", "default", "output type (default, github)")
	fs.StringVar(&o.Output, "output", "default", "output type (default, github)")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [OPTION]... [PATH]...\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "Try '%s --help' for more information.\n", os.Args[0])
	}
	return fs
}

// PrintLongUsage prints the long help for the options.
func (o *Options) PrintLongUsage() {
	fmt.Fprintf(os.Stdout, `Usage: %s [OPTION]... [PATH]...
Search for TODOS in code.

OPTIONS:
  -h, --help                  Print help and exit.
  --include-hidden            Include hidden files and directories.
  --include-docs              Include documentation.
  --include-vendored          Include vendored directories.
  -o, --output=TYPE           Output type (default, github).
  --version                   Print version information and exit.
`, os.Args[0])
}

// PrintVersion prints version information.
func (o *Options) PrintVersion() {
	versionInfo := version.GetVersionInfo()

	fmt.Fprintf(os.Stdout, `todos %s
Copyright (c) Google LLC
License Apache License Version 2.0

%s`, versionInfo.GitVersion, versionInfo.String())
}

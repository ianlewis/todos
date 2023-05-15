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
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ianlewis/linguist"

	"github.com/ianlewis/todos/internal/todos"
)

const (
	exitCodeSuccess int = iota
	exitCodeFlagParseError
	exitCodeDirWalkError
)

type todoOpt struct {
	fileName string
	todo     todos.TODO
}

type lineWriter func(todoOpt)

func outReadable(o todoOpt) {
	fmt.Printf("%s:%d:%s\n", o.fileName, o.todo.Line, o.todo.Text)
}

func outGithub(o todoOpt) {
	typ := "notice"
	switch o.todo.Type {
	case "TODO", "HACK":
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

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], msg)
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

	var exitCode int
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			printError(fmt.Sprintf("%s: %v", path, err))
			exitCode = exitCodeDirWalkError
			continue
		}

		fInfo, err := f.Stat()
		if err != nil {
			printError(fmt.Sprintf("%s: %v", path, err))
			exitCode = exitCodeDirWalkError
			continue
		}

		if fInfo.IsDir() {
			// Walk the directory
			exitCode = walkDir(path, exitCode, opts, outFunc)
		} else {
			// single file
			exitCode = scanFile(f, exitCode, outFunc)
		}

		f.Close()
	}

	os.Exit(exitCode)
}

func walkDir(path string, exitCode int, opts *Options, outFunc lineWriter) int {
	if err := fs.WalkDir(os.DirFS(path), ".", func(subPath string, d fs.DirEntry, err error) error {
		// If the path had an error then just skip it. WalkDir has likely hit the path already.
		if err != nil {
			printError(fmt.Sprintf("%s: %v", subPath, err))
			exitCode = exitCodeDirWalkError
			return nil
		}

		fullPath, err := filepath.EvalSymlinks(filepath.Join(path, subPath))
		if err != nil {
			// NOTE: If the symbolic link couldn't be evaluated just skip it.
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		f, err := os.Open(fullPath)
		if err != nil {
			printError(fmt.Sprintf("%s: %v", subPath, err))
			exitCode = exitCodeDirWalkError
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		defer f.Close()

		if d.IsDir() {
			// NOTE: If subPath is "." then this path was explicitly included.
			if subPath != "." {
				if isHidden(fullPath) && !opts.IncludeHidden {
					// Skip hidden files.
					return fs.SkipDir
				}
				if !opts.IncludeVendored && linguist.IsVendored(fullPath) {
					return fs.SkipDir
				}
				if !opts.IncludeDocs && linguist.IsDocumentation(fullPath) {
					return fs.SkipDir
				}
			}
			return nil
		}

		if isHidden(fullPath) && !opts.IncludeHidden {
			// Skip hidden files.
			return nil
		}

		exitCode = scanFile(f, exitCode, outFunc)

		return nil
	}); err != nil {
		// This shouldn't happen. Errors are all handled in the WalkDir.
		panic(err)
	}

	return exitCode
}

func isHidden(path string) bool {
	base := filepath.Base(path)
	if base == "." || base == ".." {
		return false
	}
	// TODO(github.com/ianlewis/todos/issues/7): support hidden files on windows.
	return strings.HasPrefix(base, ".")
}

func scanFile(f *os.File, exitCode int, out lineWriter) int {
	s, err := todos.CommentScannerFromFile(f)
	if err != nil {
		printError(fmt.Sprintf("%s: %v", f.Name(), err))
		return exitCodeDirWalkError
	}

	// Skip files that can't be scanned.
	if s == nil {
		return exitCode
	}
	t := todos.NewTODOScanner(s)
	for t.Scan() {
		todo := t.Next()
		out(todoOpt{
			fileName: f.Name(),
			todo:     todo,
		})
	}
	if err := t.Err(); err != nil {
		printError(fmt.Sprintf("%s: %v", f.Name(), err))
		return exitCodeDirWalkError
	}

	return exitCode
}

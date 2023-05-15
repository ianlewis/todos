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
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ianlewis/linguist"

	"github.com/ianlewis/todos/internal/todos"
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
		os.Exit(1)
	}
	outFunc, ok := outTypes[opts.Output]
	if !ok {
		printError(fmt.Sprintf("invalid output type: %q", opts.Output))
		os.Exit(1)
	}

	if opts.Help {
		opts.PrintLongUsage()
		os.Exit(0)
	}

	if opts.Version {
		opts.PrintVersion()
		os.Exit(0)
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
			exitCode = 2
			continue
		}

		fInfo, err := f.Stat()
		if err != nil {
			printError(fmt.Sprintf("%s: %v", path, err))
			exitCode = 2
			continue
		}

		if fInfo.IsDir() {
			if err := fs.WalkDir(os.DirFS(path), ".", func(subPath string, d fs.DirEntry, err error) error {
				// If the path had an error then just skip it. WalkDir has likely hit the path already.
				if err != nil {
					printError(fmt.Sprintf("%s: %v", subPath, err))
					exitCode = 2
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
					exitCode = 2
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

				scanFile(f, outFunc)

				return nil
			}); err != nil {
				// This shouldn't happen. Errors are all handled in the WalkDir.
				panic(err)
			}
		} else {
			// single file
			scanFile(f, outFunc)
		}

		f.Close()
	}

	os.Exit(exitCode)
}

func isHidden(path string) bool {
	base := filepath.Base(path)
	if base == "." || base == ".." {
		return false
	}
	// TODO(github.com/ianlewis/todos/issues/7): support hidden files on windows.
	return strings.HasPrefix(base, ".")
}

func scanFile(f *os.File, out lineWriter) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic parsing file %q\n", f.Name())
			panic(err)
		}
	}()

	s, err := todos.CommentScannerFromFile(f)
	if err != nil {
		panic(err)
	}

	// Skip files that can't be scanned.
	if s == nil {
		return
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
		panic(err)
	}
}

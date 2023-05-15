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

func main() {
	opts := &Options{}
	fSet := opts.FlagSet()
	if err := fSet.Parse(os.Args[1:]); err != nil {
		panic(err)
	}
	outFunc, ok := outTypes[opts.Output]
	if !ok {
		panic(fmt.Sprintf("invalid output type: %q", opts.Output))
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

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}

		fInfo, err := f.Stat()
		if err != nil {
			panic(err)
		}

		if fInfo.IsDir() {
			fs.WalkDir(os.DirFS(path), ".", func(subPath string, d fs.DirEntry, err error) error {
				fullPath, err := filepath.EvalSymlinks(filepath.Join(path, subPath))
				if err != nil {
					// NOTE: If the symblink couldn't be evaluated just skip it.
					return nil
				}

				f, err := os.Open(fullPath)
				if err != nil {
					panic(err)
				}
				defer f.Close()

				fInfo, err := f.Stat()
				if err != nil {
					panic(err)
				}
				if fInfo.IsDir() {
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
			})
		} else {
			// single file
			scanFile(f, outFunc)
		}

		f.Close()
	}

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

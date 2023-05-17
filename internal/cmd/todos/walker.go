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

func printError(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], msg)
}

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

func isHidden(path string) bool {
	base := filepath.Base(path)
	if base == "." || base == ".." {
		return false
	}
	// TODO(github.com/ianlewis/todos/issues/7): support hidden files on windows.
	return strings.HasPrefix(base, ".")
}

// TODOWalker walks the directory tree and scans files for TODOS.
type TODOWalker struct {
	// opts are the configuration options.
	outFunc lineWriter

	// todoConfig is the TODOScanner config.
	todoConfig *todos.Config

	// includeHidden indicates whether hidden paths should be processed. Hidden
	// paths are always processed if there are specified explicitly in `paths`.
	includeHidden bool

	// includeVendored indicates whether vendored paths should be processed. Vendored
	// paths are always processed if there are specified explicitly in `paths`.
	includeVendored bool

	// includeDocs indicates whether docs paths should be processed. Docs
	// paths are always processed if there are specified explicitly in `paths`.
	includeDocs bool

	// paths is a list of paths to walk.
	paths []string

	// The last error encountered.
	err error
}

// Walk walks the paths and scans files it finds for TODOs. It does not fail
// when it encounters errors. It instead prints an error message and returns true
// if errors were encountered.
func (w *TODOWalker) Walk() bool {
	for _, path := range w.paths {
		f, err := os.Open(path)
		if err != nil {
			printError(fmt.Sprintf("%s: %v", path, err))
			w.err = err
			continue
		}

		fInfo, err := f.Stat()
		if err != nil {
			printError(fmt.Sprintf("%s: %v", path, err))
			w.err = err
			continue
		}

		if fInfo.IsDir() {
			// Walk the directory
			w.walkDir(path)
		} else {
			// single file
			w.scanFile(f)
		}

		f.Close()
	}

	return w.err != nil
}

func (w *TODOWalker) walkDir(path string) {
	if err := fs.WalkDir(os.DirFS(path), ".", func(subPath string, d fs.DirEntry, err error) error {
		// If the path had an error then just skip it. WalkDir has likely hit the path already.
		if err != nil {
			printError(fmt.Sprintf("%s: %v", subPath, err))
			w.err = err
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
			w.err = err
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		defer f.Close()

		if d.IsDir() {
			// NOTE: If subPath is "." then this path was explicitly included.
			if subPath != "." {
				if isHidden(fullPath) && !w.includeHidden {
					// Skip hidden files.
					return fs.SkipDir
				}
				if !w.includeVendored && linguist.IsVendored(fullPath) {
					return fs.SkipDir
				}
				if !w.includeDocs && linguist.IsDocumentation(fullPath) {
					return fs.SkipDir
				}
			}
			return nil
		}

		if isHidden(fullPath) && !w.includeHidden {
			// Skip hidden files.
			return nil
		}

		w.scanFile(f)

		return nil
	}); err != nil {
		// This shouldn't happen. Errors are all handled in the WalkDir.
		panic(err)
	}
}

func (w *TODOWalker) scanFile(f *os.File) {
	s, err := todos.CommentScannerFromFile(f)
	if err != nil {
		printError(fmt.Sprintf("%s: %v", f.Name(), err))
		w.err = err
		return
	}

	// Skip files that can't be scanned.
	if s == nil {
		return
	}
	t := todos.NewTODOScanner(s, w.todoConfig)
	for t.Scan() {
		todo := t.Next()
		w.outFunc(todoOpt{
			fileName: f.Name(),
			todo:     todo,
		})
	}
	if err := t.Err(); err != nil {
		printError(fmt.Sprintf("%s: %v", f.Name(), err))
		w.err = err
		return
	}
}

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

	"github.com/fatih/color"
	"github.com/ianlewis/linguist"

	"github.com/ianlewis/todos/internal/scanner"
	"github.com/ianlewis/todos/internal/todos"
)

type todoOpt struct {
	fileName string
	todo     *todos.TODO
}

type lineWriter func(todoOpt)

// errHandler functions handle errors and return if they are fatal.
type errHandler func(error)

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

// TODOWalker walks the directory tree and scans files for TODOS.
type TODOWalker struct {
	// outFunc is for printing when todos are found.
	outFunc lineWriter

	// errHandler is for handling errors.
	errFunc errHandler

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

	// path is the currently walked path.
	path string

	// The last error encountered.
	err error
}

// Walk walks the paths and scans files it finds for TODOs. It does not fail
// when it encounters errors. It instead prints an error message and returns true
// if errors were encountered.
func (w *TODOWalker) Walk() bool {
	for _, path := range w.paths {
		w.path = path

		f, err := os.Open(path)
		if err != nil {
			w.handleErr(path, err)
			continue
		}

		fInfo, err := f.Stat()
		if err != nil {
			w.handleErr(path, err)
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
	if err := fs.WalkDir(os.DirFS(path), ".", w.walkFunc); err != nil {
		// This shouldn't happen. Errors are all handled in the WalkDir.
		panic(err)
	}
}

// walkFunc implements io.fs.WalkDirFunc.
func (w *TODOWalker) walkFunc(path string, d fs.DirEntry, err error) error {
	// If the path had an error then just skip it. WalkDir has likely hit the path already.
	if err != nil {
		w.handleErr(path, err)
		return nil
	}

	fullPath, err := filepath.EvalSymlinks(filepath.Join(w.path, path))
	if err != nil {
		// NOTE: If the symbolic link couldn't be evaluated just skip it.
		if d.IsDir() {
			return fs.SkipDir
		}
		return nil
	}

	f, err := os.Open(fullPath)
	if err != nil {
		w.handleErr(path, err)
		if d.IsDir() {
			return fs.SkipDir
		}
		return nil
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		w.handleErr(path, err)
		if d.IsDir() {
			return fs.SkipDir
		}
		return nil
	}

	// NOTE(github.com/ianlewis/todos/issues/40): d.IsDir sometimes returns false for some directories.
	if info.IsDir() {
		return w.processDir(path, fullPath)
	}
	return w.processFile(path, fullPath, f)
}

func (w *TODOWalker) processDir(path, fullPath string) error {
	// NOTE: If path is "." then this path was explicitly included.
	if path == "." {
		return nil
	}

	hdn, err := isHidden(fullPath)
	if err != nil {
		w.handleErr(path, err)
		return fs.SkipDir
	}

	if hdn && !w.includeHidden {
		// Skip hidden files.
		return fs.SkipDir
	}
	if !w.includeVendored && linguist.IsVendored(fullPath) {
		return fs.SkipDir
	}
	if !w.includeDocs && linguist.IsDocumentation(fullPath) {
		return fs.SkipDir
	}
	return nil
}

func (w *TODOWalker) processFile(path, fullPath string, f *os.File) error {
	hdn, err := isHidden(fullPath)
	if err != nil {
		w.handleErr(path, err)
		return nil
	}

	if hdn && !w.includeHidden {
		// Skip hidden files.
		return nil
	}

	w.scanFile(f)

	return nil
}

func (w *TODOWalker) scanFile(f *os.File) {
	s, err := scanner.FromFile(f)
	if err != nil {
		w.handleErr(f.Name(), err)
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
		w.handleErr(f.Name(), err)
	}
}

func (w *TODOWalker) handleErr(prefix string, err error) {
	w.err = err
	if w.errFunc != nil {
		if prefix != "" {
			err = fmt.Errorf("%s: %w", prefix, err)
		}
		w.errFunc(err)
	}
}

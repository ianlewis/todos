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

package walker

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/ianlewis/linguist"

	"github.com/ianlewis/todos/internal/scanner"
	"github.com/ianlewis/todos/internal/todos"
)

// TODORef represents a TODO in a specific file.
type TODORef struct {
	FileName string
	TODO     *todos.TODO
}

// TODOHandler handles found TODO references. It can return SkipAll or SkipDir.
type TODOHandler func(*TODORef) error

// ErrorHandler handles found TODO references. It can return SkipAll or SkipDir.
type ErrorHandler func(error) error

// Options are options for the walker.
type Options struct {
	// TODOFunc handles when TODOs are found.
	TODOFunc TODOHandler

	// ErrorFunc handles when errors are found.
	ErrorFunc ErrorHandler

	// Config is the config for scanning todos.
	Config *todos.Config

	// Charset is the character set to use when reading the files or 'detect'
	// for charset detection.
	Charset string

	// ExcludeGlobs is a list of Glob that matches excluded files.
	ExcludeGlobs []glob.Glob

	// ExcludeDirGlobs is a list of Glob that matches excluded dirs.
	ExcludeDirGlobs []glob.Glob

	// IncludeHidden indicates whether hidden paths should be processed. Hidden
	// paths are always processed if there are specified explicitly in `paths`.
	IncludeHidden bool

	// IncludeVendored indicates whether vendored paths should be processed. Vendored
	// paths are always processed if there are specified explicitly in `paths`.
	IncludeVendored bool

	// IncludeVCS indicates that VCS paths (.git, .hg, .svn, etc.) should be included.
	IncludeVCS bool

	// Paths are the paths to walk to look for TODOs.
	Paths []string
}

// New returns a new walker for the options.
func New(opts *Options) *TODOWalker {
	if opts == nil {
		opts = &Options{
			Config: &todos.Config{
				Types: []string{"TODO"},
			},
		}
	}
	return &TODOWalker{
		options: opts,
	}
}

// TODOWalker walks the directory tree and scans files for TODOS.
type TODOWalker struct {
	// options are the walker's options.
	options *Options

	// path is the currently walked path.
	path string

	// The last error encountered.
	err error
}

// Walk walks the paths and scans files it finds for TODOs. It does not fail
// when it encounters errors. It instead prints an error message and returns true
// if errors were encountered.
func (w *TODOWalker) Walk() bool {
	for _, path := range w.options.Paths {
		w.path = path

		f, err := os.Open(path)
		if err != nil {
			if herr := w.handleErr(path, err); herr != nil {
				break
			}
			continue
		}

		fInfo, err := f.Stat()
		if err != nil {
			if herr := w.handleErr(path, err); herr != nil {
				break
			}
			continue
		}

		if fInfo.IsDir() {
			// Walk the directory
			w.walkDir(path)
		} else {
			// single file
			if err := w.scanFile(f); err != nil {
				if herr := w.handleErr(path, err); herr != nil {
					break
				}
			}
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
		return w.handleErr(path, err)
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
		if herr := w.handleErr(path, err); herr != nil {
			return herr
		}
		if d.IsDir() {
			return fs.SkipDir
		}
		return nil
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		if herr := w.handleErr(path, err); herr != nil {
			return herr
		}
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

	// Exclude directories that match one of the given glob patterns.
	for _, g := range w.options.ExcludeDirGlobs {
		if g.Match(filepath.Base(fullPath)) {
			return fs.SkipDir
		}
	}

	hdn, err := isHidden(fullPath)
	if err != nil {
		if herr := w.handleErr(path, err); herr != nil {
			return herr
		}
		return fs.SkipDir
	}

	if hdn && !w.options.IncludeHidden {
		// Skip hidden files.
		return fs.SkipDir
	}

	if !w.options.IncludeVCS && isVCS(fullPath) {
		return fs.SkipDir
	}

	// NOTE: linguist only matches paths with a *nix path separators.
	// TODO(github.com/ianlewis/linguist/issues/1): Update when linguist supports Windows.
	fullPath = strings.ReplaceAll(fullPath, string(os.PathSeparator), "/")

	// NOTE: linguist only matches paths with a path separator at the end.
	if !strings.HasSuffix(fullPath, "/") {
		fullPath += "/"
	}
	if !w.options.IncludeVendored && linguist.IsVendored(fullPath) {
		return fs.SkipDir
	}
	return nil
}

func (w *TODOWalker) processFile(path, fullPath string, f *os.File) error {
	// Exclude files that match one of the given glob patterns.
	for _, g := range w.options.ExcludeGlobs {
		if g.Match(filepath.Base(fullPath)) {
			// Skip file.
			return nil
		}
	}

	hdn, err := isHidden(fullPath)
	if err != nil {
		return w.handleErr(path, err)
	}

	if hdn && !w.options.IncludeHidden {
		// Skip hidden files.
		return nil
	}

	return w.scanFile(f)
}

func (w *TODOWalker) scanFile(f *os.File) error {
	s, err := scanner.FromFile(f, w.options.Charset)
	if err != nil {
		if herr := w.handleErr(f.Name(), err); herr != nil {
			return herr
		}
	}

	// Skip files that can't be scanned.
	if s == nil {
		return nil
	}
	t := todos.NewTODOScanner(s, w.options.Config)
	for t.Scan() {
		todo := t.Next()
		if w.options.TODOFunc != nil {
			if err := w.options.TODOFunc(&TODORef{
				FileName: f.Name(),
				TODO:     todo,
			}); err != nil {
				return err
			}
		}
	}
	if err := t.Err(); err != nil {
		if herr := w.handleErr(f.Name(), err); herr != nil {
			return herr
		}
	}

	return nil
}

func (w *TODOWalker) handleErr(prefix string, err error) error {
	// If it's a skip error then just return it.
	if errors.Is(err, fs.SkipDir) || errors.Is(err, fs.SkipAll) {
		return err
	}

	w.err = err
	if w.options.ErrorFunc != nil {
		if prefix != "" {
			err = fmt.Errorf("%s: %w", prefix, err)
		}
		if herr := w.options.ErrorFunc(err); herr != nil {
			return herr
		}
	}
	return nil
}

// isVCS returns whether the path is a vcs path. Should only be called on directories.
func isVCS(path string) bool {
	basePath := filepath.Base(path)
	return basePath == ".git" || basePath == ".hg" || basePath == ".svn"
}

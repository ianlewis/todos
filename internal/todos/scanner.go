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

package todos

import (
	"bytes"
	"io"
	"os"

	"github.com/ianlewis/linguist"

	"github.com/ianlewis/todos/internal/scanner"
)

var (
	languageMap = map[string]scanner.Config{
		"C":            scanner.CConfig,
		"C++":          scanner.CPPConfig,
		"C#":           scanner.CSConfig,
		"Dockerfile":   scanner.DockerfileConfig,
		"Go":           scanner.GoConfig,
		"Go Module":    scanner.GoConfig,
		"Go Checksums": scanner.GoConfig,
		"HTML":         scanner.HTMLConfig,
		"Java":         scanner.JavaConfig,
		"JavaScript":   scanner.JavascriptConfig,
		// NOTE: Some JSON files support JS comments (e.g. tsconfig.json)
		"JSON":        scanner.JavascriptConfig,
		"Makefile":    scanner.MakefileConfig,
		"Objective-C": scanner.ObjectiveCConfig,
		"Perl":        scanner.PerlConfig,
		"PHP":         scanner.PHPConfig,
		"Python":      scanner.PythonConfig,
		"Ruby":        scanner.RubyConfig,
		"Scala":       scanner.ScalaConfig,
		"Shell":       scanner.ShellConfig,
		"Swift":       scanner.SwiftConfig,
		"TOML":        scanner.TOMLConfig,
		"TypeScript":  scanner.TypescriptConfig,
		"XML":         scanner.XMLConfig,
		"YAML":        scanner.YAMLConfig,
	}
)

// CommentScannerFromFile returns an appropriate CommentScanner for the given file.
func CommentScannerFromFile(f *os.File) (*scanner.CommentScanner, error) {
	contents, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	if linguist.ShouldIgnoreContents(contents) {
		return nil, nil
	}

	name := f.Name()
	lang := linguist.LanguageByContents(contents, linguist.LanguageHints(name))

	if config, ok := languageMap[lang]; ok {
		return scanner.New(bytes.NewReader(contents), config), nil
	}

	return nil, nil
}

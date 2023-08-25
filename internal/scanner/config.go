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

package scanner

// Config is configuration for a generic comment scanner.
type Config struct {
	LineCommentStart      []string
	MultilineCommentStart string
	MultilineCommentEnd   string
	Strings               [][2]string

	// escapeFunc returns if the scanner is currently at a escaped string.
	escapeFunc func(s *CommentScanner, st *stateString) (bool, error)
}

func noEscape(_ *CommentScanner, _ *stateString) (bool, error) {
	return false, nil
}

func backslashEscape(s *CommentScanner, st *stateString) (bool, error) {
	return s.peekEqual(append([]rune{'\\'}, s.config.Strings[st.index][1]...))
}

func doubleEscape(s *CommentScanner, st *stateString) (bool, error) {
	b := append([]rune{}, s.config.Strings[st.index][1]...)
	b = append(b, s.config.Strings[st.index][1]...)
	return s.peekEqual(b)
}

var (
	languageMap = map[string]*Config{
		"Assembly":     &AssemblyConfig,
		"C":            &CConfig,
		"C++":          &CPPConfig,
		"C#":           &CSConfig,
		"Dockerfile":   &DockerfileConfig,
		"Erlang":       &ErlangConfig,
		"Go":           &GoConfig,
		"Go Module":    &GoConfig,
		"Go Checksums": &GoConfig,
		"Haskell":      &HaskellConfig,
		"HTML":         &HTMLConfig,
		"Java":         &JavaConfig,
		"JavaScript":   &JavascriptConfig,
		// NOTE: Some JSON files support JS comments (e.g. tsconfig.json)
		"JSON":          &JavascriptConfig,
		"Lua":           &LuaConfig,
		"Makefile":      &MakefileConfig,
		"Objective-C":   &ObjectiveCConfig,
		"Perl":          &PerlConfig,
		"PHP":           &PHPConfig,
		"Python":        &PythonConfig,
		"Ruby":          &RubyConfig,
		"Scala":         &ScalaConfig,
		"Shell":         &ShellConfig,
		"Swift":         &SwiftConfig,
		"SQL":           &SQLConfig,
		"TOML":          &TOMLConfig,
		"TypeScript":    &TypescriptConfig,
		"Unix Assembly": &UnixAssemblyConfig,
		"XML":           &XMLConfig,
		"YAML":          &YAMLConfig,
	}

	// AssemblyConfig is a config for Assembly.
	AssemblyConfig = Config{
		LineCommentStart:      []string{";"},
		MultilineCommentStart: "/*",
		MultilineCommentEnd:   "*/",
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"},
		},
		escapeFunc: noEscape,
	}

	// CConfig is a config for C.
	CConfig = Config{
		LineCommentStart:      []string{"//"},
		MultilineCommentStart: "/*",
		MultilineCommentEnd:   "*/",
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"}, // character
		},
		escapeFunc: backslashEscape,
	}

	// CPPConfig is a config for C++.
	CPPConfig = CConfig

	// CSConfig is a config for C#
	// NOTE: @"" strings are handled by normal double quote handling.
	CSConfig = CConfig

	// DockerfileConfig is a config for Dockerfiles.
	DockerfileConfig = ShellConfig

	// ErlangConfig is a config for Erlang.
	ErlangConfig = Config{
		LineCommentStart: []string{"%"},
		// NOTE: Erlang does not have multi-line comments.
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"}, // Atom
		},
		escapeFunc: backslashEscape,
	}

	// GoConfig is a config for Go.
	GoConfig = Config{
		LineCommentStart:      []string{"//"},
		MultilineCommentStart: "/*",
		MultilineCommentEnd:   "*/",
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"}, // Rune.
			{"`", "`"},
		},
		escapeFunc: backslashEscape,
	}

	// HaskellConfig is a config for Haskell.
	HaskellConfig = Config{
		LineCommentStart:      []string{"--"},
		MultilineCommentStart: "{-",
		MultilineCommentEnd:   "-}",
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"}, // Character
		},
		escapeFunc: backslashEscape,
	}

	// HTMLConfig is a config for HTML.
	HTMLConfig = XMLConfig

	// JavaConfig is a config for Java.
	// NOTE: All strings are treated as multi-line.
	JavaConfig = CConfig

	// JavascriptConfig is a config for Javascript.
	JavascriptConfig = CConfig

	// LuaConfig is a config for Lua.
	LuaConfig = Config{
		LineCommentStart:      []string{"--"},
		MultilineCommentStart: "--[[",
		MultilineCommentEnd:   "--]]",
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"}, // character
		},
		escapeFunc: backslashEscape,
	}

	// MakefileConfig is a config for Makefiles.
	MakefileConfig = ShellConfig

	// ObjectiveCConfig is a config for Objective-C.
	ObjectiveCConfig = CConfig

	// PerlConfig is a config for Perl.
	PerlConfig = Config{
		LineCommentStart:      []string{"#"},
		MultilineCommentStart: "=",
		MultilineCommentEnd:   "=cut",
		Strings: [][2]string{
			// TODO(#1): Perl supports strings with any delimiter.
			{"\"", "\""},
			{"'", "'"},
		},
		escapeFunc: backslashEscape,
	}

	// PHPConfig is a config for PHP.
	PHPConfig = Config{
		LineCommentStart:      []string{"#", "//"},
		MultilineCommentStart: "/*",
		MultilineCommentEnd:   "*/",
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"}, // character
		},
		escapeFunc: backslashEscape,
	}

	// PythonConfig is a config for Python.
	// TODO(#1): Python parsing should also include python docstrings.
	PythonConfig = Config{
		LineCommentStart:      []string{"#"},
		MultilineCommentStart: "\"\"\"",
		MultilineCommentEnd:   "\"\"\"",
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"},
		},
		escapeFunc: backslashEscape,
	}

	// RConfig is a config for R.
	RConfig = Config{
		LineCommentStart: []string{"#"},
		// NOTE: R has no multi-line comments.
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"},
		},
		escapeFunc: backslashEscape,
	}

	// RubyConfig is a config for Ruby.
	RubyConfig = Config{
		LineCommentStart:      []string{"#"},
		MultilineCommentStart: "=begin",
		MultilineCommentEnd:   "=end",
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"},
			{"%{", "}"},
		},
		escapeFunc: backslashEscape,
	}

	// RustConfig is a config for Rust.
	RustConfig = CConfig

	// ScalaConfig is a config for Scala.
	ScalaConfig = JavaConfig

	// ShellConfig is a config for Shell.
	ShellConfig = Config{
		LineCommentStart: []string{"#"},
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"},
		},
		escapeFunc: backslashEscape,
	}

	// SwiftConfig is a config for Swift.
	SwiftConfig = Config{
		LineCommentStart:      []string{"//"},
		MultilineCommentStart: "/*",
		MultilineCommentEnd:   "*/",
		Strings: [][2]string{
			{"\"", "\""},
		},
		escapeFunc: backslashEscape,
	}

	// SQLConfig is a config for SQL.
	SQLConfig = Config{
		LineCommentStart:      []string{"--"},
		MultilineCommentStart: "/*",
		MultilineCommentEnd:   "*/",
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"},
		},
		escapeFunc: doubleEscape,
	}

	// TOMLConfig is a config for TOML.
	TOMLConfig = ShellConfig

	// TypescriptConfig is a config for Typescript.
	TypescriptConfig = JavascriptConfig

	// UnixAssemblyConfig is a config for Unix Assembly.
	UnixAssemblyConfig = AssemblyConfig

	// XMLConfig is a config for XML.
	XMLConfig = Config{
		MultilineCommentStart: "<!--",
		MultilineCommentEnd:   "-->",
		// For attributes.
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"},
		},
		escapeFunc: backslashEscape,
	}

	// YAMLConfig is a config for YAML.
	YAMLConfig = ShellConfig
)

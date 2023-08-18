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
}

var (
	languageMap = map[string]*Config{
		"Assembly":     &AssemblyConfig,
		"C":            &CConfig,
		"C++":          &CPPConfig,
		"C#":           &CSConfig,
		"Dockerfile":   &DockerfileConfig,
		"Go":           &GoConfig,
		"Go Module":    &GoConfig,
		"Go Checksums": &GoConfig,
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
		"TOML":          &TOMLConfig,
		"TypeScript":    &TypescriptConfig,
		"Unix Assembly": &UnixAssemblyConfig,
		"XML":           &XMLConfig,
		"YAML":          &YAMLConfig,
	}

	// AssemblyConfig is a config for Assembly.
	AssemblyConfig = Config{
		LineCommentStart: []string{";"},
		// TODO(#1): Parsing should exclude the leading '*' for multi-line comments.
		MultilineCommentStart: "/*",
		MultilineCommentEnd:   "*/",
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"},
		},
	}

	// CConfig is a config for C.
	CConfig = Config{
		LineCommentStart: []string{"//"},
		// TODO(#1): Parsing should exclude the leading '*' for multi-line comments.
		MultilineCommentStart: "/*",
		MultilineCommentEnd:   "*/",
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"}, // character
		},
	}

	// CPPConfig is a config for C++.
	CPPConfig = CConfig

	// CSConfig is a config for C#
	// NOTE: @"" strings are handled by normal double quote handling.
	CSConfig = CConfig

	// DockerfileConfig is a config for Dockerfiles.
	DockerfileConfig = ShellConfig

	// GoConfig is a config for Go.
	GoConfig = Config{
		LineCommentStart: []string{"//"},
		// TODO(#1): Parsing should exclude the leading '*' for multi-line comments.
		MultilineCommentStart: "/*",
		MultilineCommentEnd:   "*/",
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"}, // Rune.
			{"`", "`"},
		},
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
	}

	// PHPConfig is a config for PHP.
	PHPConfig = Config{
		LineCommentStart: []string{"#", "//"},
		// TODO(#1): Parsing should exclude the leading '*' for multi-line comments.
		MultilineCommentStart: "/*",
		MultilineCommentEnd:   "*/",
		Strings: [][2]string{
			{"\"", "\""},
			{"'", "'"}, // character
		},
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
	}

	// SwiftConfig is a config for Swift.
	SwiftConfig = Config{
		LineCommentStart: []string{"//"},
		// TODO(#1): Parsing should exclude the leading '*' for multi-line comments.
		MultilineCommentStart: "/*",
		MultilineCommentEnd:   "*/",
		Strings: [][2]string{
			{"\"", "\""},
		},
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
	}

	// YAMLConfig is a config for YAML.
	YAMLConfig = ShellConfig
)

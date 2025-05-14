// Copyright 2024 Google LLC
// Copyright 2025 Ian Lewis
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

// Common config.

func concat[T any](slices ...[]T) []T {
	var tLen int
	for _, s := range slices {
		tLen += len(s)
	}
	newS := make([]T, tLen)
	var i int
	for _, s := range slices {
		i += copy(newS[i:], s)
	}
	return newS
}

// TODO(#1686): Refactor language config global variables.
//
//nolint:gochecknoglobals // remove in the future.
var (
	// sh-style languages.

	// hashLineComments are sh-style line comments.
	hashLineComments = []LineCommentConfig{
		{
			Start: []rune("#"),
		},
	}

	// C-style languages.

	// cLineComments are C-style line comments.
	cLineComments = []LineCommentConfig{
		{
			Start: []rune("//"),
		},
	}

	// cBlockComments are C-style block comments.
	cBlockComments = []MultilineCommentConfig{
		{
			Start: []rune("/*"),
			End:   []rune("*/"),
		},
	}

	singleQuoteString = []StringConfig{
		{
			Start:      []rune{'\''},
			End:        []rune{'\''},
			EscapeFunc: CharEscape('\\'),
		},
	}

	doubleQuoteString = []StringConfig{
		{
			Start:      []rune{'"'},
			End:        []rune{'"'},
			EscapeFunc: CharEscape('\\'),
		},
	}

	tripleDoubleQuoteString = []StringConfig{
		{
			Start:      []rune(`"""`),
			End:        []rune(`"""`),
			EscapeFunc: CharEscape('\\'),
		},
	}

	// cStrings are C-style strings.
	cStrings = concat(
		// Strings
		doubleQuoteString,
		// Characters
		singleQuoteString,
	)

	singleQuoteStringNoEscape = []StringConfig{
		{
			Start:      []rune{'\''},
			End:        []rune{'\''},
			EscapeFunc: NoEscape,
		},
	}

	doubleQuoteStringNoEscape = []StringConfig{
		{
			Start:      []rune{'"'},
			End:        []rune{'"'},
			EscapeFunc: NoEscape,
		},
	}

	fortranStrings = concat(
		// Strings
		doubleQuoteStringNoEscape,
		// Characters
		singleQuoteStringNoEscape,
	)

	// XML-style languages.

	// xmlBlockComments are XML-style block comments.
	xmlBlockComments = []MultilineCommentConfig{
		{
			Start: []rune("<!--"),
			End:   []rune("-->"),
		},
	}

	tripleDoubleQuoteComments = []MultilineCommentConfig{
		{
			Start: []rune(`"""`),
			End:   []rune(`"""`),
		},
	}
)

// LanguagesConfig is a map of language names to their configuration for all
// languages supported by CommentScanner.
//
// TODO(#1686): Refactor LanguagesConfig global variable.
//
//nolint:gochecknoglobals // remove in the future.
var LanguagesConfig = map[string]*Config{
	"Assembly": {
		LineComments: []LineCommentConfig{
			{
				Start: []rune{';'},
			},
		},
		MultilineComments: cBlockComments,
		// NOTE: Assembly doesn't have string escape characters.
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: NoEscape,
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: NoEscape,
			},
		},
	},
	"C": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings:           cStrings,
	},
	"C#": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings:           cStrings,
	},
	"C++": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings:           cStrings,
	},
	"Clojure": {
		LineComments: []LineCommentConfig{
			{Start: []rune{';'}},
		},
		MultilineComments: nil,
		Strings:           doubleQuoteString,
	},
	"CODEOWNERS": {
		LineComments: []LineCommentConfig{
			{
				Start:       []rune("#"),
				AtLineStart: true,
			},
		},
	},
	"CoffeeScript": {
		LineComments: hashLineComments,
		MultilineComments: []MultilineCommentConfig{
			{
				Start: []rune("###"),
				End:   []rune("###"),
			},
		},
		Strings: cStrings,
	},
	"Dart": {
		LineComments: concat(
			cLineComments,
			[]LineCommentConfig{
				// Dart documentation comments.
				{Start: []rune("///")},
			},
		),
		MultilineComments: cBlockComments,
		Strings:           cStrings,
	},
	"Dockerfile": {
		LineComments:      hashLineComments,
		MultilineComments: nil,
		Strings:           cStrings,
	},
	"Elixir": {
		LineComments: hashLineComments,
		// Support function documentation.
		MultilineComments: []MultilineCommentConfig{
			{
				Start: []rune(`@moduledoc """`),
				End:   []rune(`"""`),
			},
			{
				Start: []rune(`@doc """`),
				End:   []rune(`"""`),
			},
		},
		Strings: concat(
			singleQuoteString,
			doubleQuoteString,
			tripleDoubleQuoteString,
			[]StringConfig{
				{
					Start:      []rune("'''"),
					End:        []rune("'''"),
					EscapeFunc: CharEscape('\\'),
				},
			},
		),
	},
	"Emacs Lisp": {
		LineComments: []LineCommentConfig{
			{Start: []rune{';'}},
		},
		MultilineComments: nil,
		Strings:           doubleQuoteString,
	},
	"Erlang": {
		LineComments: []LineCommentConfig{
			{Start: []rune{'%'}},
		},
		MultilineComments: nil,
		Strings:           cStrings,
	},
	"Fortran": {
		LineComments: []LineCommentConfig{
			{Start: []rune{'!'}},
		},
		MultilineComments: nil,
		Strings:           fortranStrings,
	},
	"Fortran Free Form": {
		LineComments: []LineCommentConfig{
			{Start: []rune{'!'}},
		},
		MultilineComments: nil,
		Strings:           fortranStrings,
	},
	"Go": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings: concat(
			cStrings,
			[]StringConfig{
				// Go raw strings
				{
					Start:      []rune{'`'},
					End:        []rune{'`'},
					EscapeFunc: NoEscape,
				},
			},
		),
	},
	"Go Module": {
		LineComments:      cLineComments,
		MultilineComments: nil,
		Strings: concat(
			doubleQuoteString,
			[]StringConfig{
				// NOTE: Characters are not supported.
				// Go raw strings
				{
					Start:      []rune{'`'},
					End:        []rune{'`'},
					EscapeFunc: NoEscape,
				},
			},
		),
	},
	"GraphQL": {
		LineComments:      hashLineComments,
		MultilineComments: tripleDoubleQuoteComments,
		Strings:           doubleQuoteString,
	},
	"Groovy": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings: concat(
			cStrings,
			[]StringConfig{
				{
					Start:      []rune("'''"),
					End:        []rune("'''"),
					EscapeFunc: CharEscape('\\'),
				},
			},
		),
	},
	"HCL": {
		LineComments: concat(
			hashLineComments,
			cLineComments,
		),
		Strings: doubleQuoteString,
	},
	"HTML": {
		LineComments:      nil,
		MultilineComments: xmlBlockComments,
		Strings:           cStrings,
	},
	"HTML+ERB": {
		LineComments: nil,
		MultilineComments: concat(
			xmlBlockComments,
			[]MultilineCommentConfig{
				{
					Start: []rune("<%#"),
					End:   []rune("%>"),
				},
			},
		),
		Strings: cStrings,
	},
	"Haskell": {
		LineComments: []LineCommentConfig{
			{
				Start: []rune("--"),
			},
		},
		MultilineComments: []MultilineCommentConfig{
			{
				Start: []rune("{-"),
				End:   []rune("-}"),
			},
		},
		Strings: cStrings,
	},
	"Ignore List": {
		LineComments: []LineCommentConfig{
			{
				Start:       []rune("#"),
				AtLineStart: true,
			},
		},
	},
	"JSON": {
		// NOTE: Some JSON parsers support comments.
		LineComments: []LineCommentConfig{
			{
				Start: []rune("//"),
			},
			{
				Start: []rune{'#'},
			},
		},
		MultilineComments: cBlockComments,
		Strings:           cStrings,
	},
	"Java": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings:           cStrings,
	},
	"JavaScript": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings:           cStrings,
	},
	"Julia": {
		LineComments: concat(
			hashLineComments,
		),
		MultilineComments: []MultilineCommentConfig{
			{
				Start: []rune("#="),
				End:   []rune("=#"),
				// Julia supports nested block comments.
				Nested: true,
			},
		},
		Strings: concat(
			// Raw strings
			// https://docs.julialang.org/en/v1/manual/strings/#man-raw-string-literals
			[]StringConfig{
				{
					Start:      []rune(`raw"`),
					End:        []rune(`"`),
					EscapeFunc: NoEscape,
				},
			},
			cStrings,
			// Raw doc strings
			// https://docs.julialang.org/en/v1/manual/documentation/#Advanced-Usage
			[]StringConfig{
				{
					Start:      []rune(`raw"""`),
					End:        []rune(`"""`),
					EscapeFunc: NoEscape,
				},
			},
			tripleDoubleQuoteString,
		),
	},
	"Kotlin": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings:           cStrings,
	},
	"Lua": {
		LineComments: []LineCommentConfig{
			{
				Start: []rune("--"),
			},
		},
		MultilineComments: []MultilineCommentConfig{
			{
				Start: []rune("--[["),
				End:   []rune("--]]"),
			},
		},
		Strings: cStrings,
	},
	"MATLAB": {
		LineComments: []LineCommentConfig{
			{
				Start: []rune{'%'},
			},
		},
		MultilineComments: []MultilineCommentConfig{
			{
				Start: []rune("%{"),
				End:   []rune("}%"),
			},
		},
		Strings: cStrings,
	},
	"Makefile": {
		LineComments:      hashLineComments,
		MultilineComments: nil,
		Strings:           cStrings,
	},
	"Markdown": {
		MultilineComments: xmlBlockComments,
		Strings: []StringConfig{
			// Inline code
			{
				Start:      []rune{'`'},
				End:        []rune{'`'},
				EscapeFunc: NoEscape,
			},
			// Code block
			{
				Start:      []rune("```"),
				End:        []rune("```"),
				EscapeFunc: NoEscape,
			},
		},
	},
	"Nix": {
		LineComments:      hashLineComments,
		MultilineComments: cBlockComments,
	},
	"Objective-C": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings:           cStrings,
	},
	"OCaml": {
		MultilineComments: []MultilineCommentConfig{
			{
				Start: []rune("(*"),
				End:   []rune("*)"),
				// OCaml supports nested block comments.
				Nested: true,
			},
		},
		// TODO(#1627): Support OCaml quoted string literals.
		Strings: cStrings,
	},
	"Pascal": {
		LineComments: cLineComments, // Delphi comments
		MultilineComments: []MultilineCommentConfig{
			{
				Start: []rune("(*"),
				End:   []rune("*)"),
			},
			{
				Start: []rune("{"),
				End:   []rune("}"),
			},
		},
		Strings: []StringConfig{
			// Strings
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: DoubleEscape,
			},
			// Characters
			{
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: DoubleEscape,
			},
		},
	},
	"PHP": {
		LineComments: []LineCommentConfig{
			{
				Start: []rune{'#'},
			},
			{
				Start: []rune("//"),
			},
		},
		MultilineComments: cBlockComments,
		Strings:           cStrings,
	},
	"Perl": {
		LineComments: hashLineComments,
		MultilineComments: []MultilineCommentConfig{
			{
				Start:         []rune{'='},
				End:           []rune("=cut"),
				AtFirstColumn: true,
			},
		},
		Strings: cStrings,
	},
	"PowerShell": {
		LineComments: hashLineComments,
		MultilineComments: []MultilineCommentConfig{
			{
				Start: []rune("<#"),
				End:   []rune("#>"),
			},
		},
		// NOTE:  Powershell escape character is the grave character (`)
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('`'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('`'),
			},
		},
	},
	"Puppet": {
		LineComments:      hashLineComments,
		MultilineComments: nil,
		Strings:           cStrings,
	},
	"Python": {
		LineComments:      hashLineComments,
		MultilineComments: tripleDoubleQuoteComments,
		Strings:           cStrings,
	},
	"R": {
		LineComments:      hashLineComments,
		MultilineComments: nil,
		Strings:           cStrings,
	},
	"Ruby": {
		LineComments: hashLineComments,
		MultilineComments: []MultilineCommentConfig{
			{
				Start:         []rune("=begin"),
				End:           []rune("=end"),
				AtFirstColumn: true,
			},
		},
		Strings: concat(
			cStrings,
			[]StringConfig{
				{
					Start:      []rune("%{"),
					End:        []rune{'}'},
					EscapeFunc: CharEscape('\\'),
				},
			},
		),
	},
	"Rust": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings:           cStrings,
	},
	"SQL": {
		LineComments: []LineCommentConfig{
			{
				Start: []rune("--"),
			},
		},
		MultilineComments: cBlockComments,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: DoubleEscape,
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: DoubleEscape,
			},
		},
	},
	"Scala": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings:           cStrings,
	},
	"Shell": {
		LineComments:      hashLineComments,
		MultilineComments: nil,
		Strings:           cStrings,
	},
	"Swift": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings:           doubleQuoteString,
	},
	"TOML": {
		LineComments:      hashLineComments,
		MultilineComments: nil,
		Strings:           cStrings,
	},
	"TeX": {
		LineComments: []LineCommentConfig{
			{
				Start: []rune{'%'},
			},
		},
		MultilineComments: nil,
		Strings:           nil,
	},
	"TypeScript": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings:           cStrings,
	},
	"Unix Assembly": {
		LineComments: []LineCommentConfig{
			{
				Start: []rune{';'},
			},
		},
		MultilineComments: cBlockComments,
		// NOTE: Assembly doesn't have string escape characters.
		Strings: fortranStrings,
	},
	"VBA": {
		LineComments: []LineCommentConfig{
			{
				Start: []rune{'\''},
			},
		},
		MultilineComments: nil,
		Strings:           doubleQuoteString,
	},
	"Vim Script": {
		LineComments: []LineCommentConfig{
			{
				Start: []rune{'"'},
			},
		},
		MultilineComments: nil,
		Strings:           cStrings,
	},
	"Visual Basic .NET": {
		LineComments: []LineCommentConfig{
			{
				Start: []rune{'\''},
			},
		},
		MultilineComments: nil,
		Strings:           doubleQuoteString,
	},
	"XML": {
		LineComments:      nil,
		MultilineComments: xmlBlockComments,
		Strings:           cStrings,
	},
	"XSLT": {
		LineComments:      nil,
		MultilineComments: xmlBlockComments,
		Strings:           cStrings,
	},
	"YAML": {
		LineComments:      hashLineComments,
		MultilineComments: nil,
		Strings:           cStrings,
	},
}

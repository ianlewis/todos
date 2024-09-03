// Copyright 2024 Google LLC
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
			Start:       []rune("/*"),
			End:         []rune("*/"),
			AtLineStart: false,
		},
	}

	// cStrings are C-style strings.
	cStrings = []StringConfig{
		// Strings
		{
			Start:      []rune{'"'},
			End:        []rune{'"'},
			EscapeFunc: CharEscape('\\'),
		},
		// Characters
		{
			Start:      []rune{'\''},
			End:        []rune{'\''},
			EscapeFunc: CharEscape('\\'),
		},
	}

	// XML-style languages.

	// xmlBlockComments are XML-style block comments.
	xmlBlockComments = []MultilineCommentConfig{
		{
			Start:       []rune("<!--"),
			End:         []rune("-->"),
			AtLineStart: false,
		},
	}
)

var LanguagesConfig = map[string]*Config{
	"Assembly": {
		LineComments: []LineCommentConfig{
			{
				Start: []rune{';'},
			},
		},
		MultilineComments: []MultilineCommentConfig{
			{
				Start:       []rune("/*"),
				End:         []rune("*/"),
				AtLineStart: false,
			},
		},
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
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"CoffeeScript": {
		LineComments: hashLineComments,
		MultilineComments: []MultilineCommentConfig{
			{Start: []rune("###"), End: []rune("###"), AtLineStart: false},
		},
		Strings: cStrings,
	},
	"Dockerfile": {
		LineComments:      hashLineComments,
		MultilineComments: nil,
		Strings:           cStrings,
	},
	"Elixir": {
		LineComments: hashLineComments,
		// Support function documentation.
		// TODO(#1546): Support @moduledoc
		MultilineComments: []MultilineCommentConfig{
			{
				Start:       []rune("@moduledoc \"\"\""),
				End:         []rune("\"\"\""),
				AtLineStart: false,
			},
			{
				Start:       []rune("@doc \"\"\""),
				End:         []rune("\"\"\""),
				AtLineStart: false,
			},
		},
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune("\"\"\""),
				End:        []rune("\"\"\""),
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune("'''"),
				End:        []rune("'''"),
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Emacs Lisp": {
		LineComments: []LineCommentConfig{
			{Start: []rune{';'}},
		},
		MultilineComments: nil,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			},
		},
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
	"Fortran Free Form": {
		LineComments: []LineCommentConfig{
			{Start: []rune{'!'}},
		},
		MultilineComments: nil,
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
	"Go": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			},
			{
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
			// Go raw strings
			{
				Start:      []rune{'`'},
				End:        []rune{'`'},
				EscapeFunc: NoEscape,
			},
		},
	},
	"Go Module": {
		LineComments:      cLineComments,
		MultilineComments: nil,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			},
			// NOTE: Characters are not supported.
			// Go raw strings
			{
				Start:      []rune{'`'},
				End:        []rune{'`'},
				EscapeFunc: NoEscape,
			},
		},
	},
	"Groovy": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune("'''"),
				End:        []rune("'''"),
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"HTML": {
		LineComments:      nil,
		MultilineComments: xmlBlockComments,
		Strings:           cStrings,
	},
	"HTML+ERB": {
		LineComments: nil,
		MultilineComments: []MultilineCommentConfig{
			{
				Start:       []rune("<!--"),
				End:         []rune("-->"),
				AtLineStart: false,
			},
			{
				Start:       []rune("<%#"),
				End:         []rune("%>"),
				AtLineStart: false,
			},
		},
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
				Start:       []rune("{-"),
				End:         []rune("-}"),
				AtLineStart: false,
			},
		},
		Strings: cStrings,
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
				Start:       []rune("--[["),
				End:         []rune("--]]"),
				AtLineStart: false,
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
				Start:       []rune("%{"),
				End:         []rune("}%"),
				AtLineStart: false,
			},
		},
		Strings: cStrings,
	},
	"Makefile": {
		LineComments:      hashLineComments,
		MultilineComments: nil,
		Strings:           cStrings,
	},
	"Objective-C": {
		LineComments:      cLineComments,
		MultilineComments: cBlockComments,
		Strings:           cStrings,
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
				Start:       []rune{'='},
				End:         []rune("=cut"),
				AtLineStart: true,
			},
		},
		Strings: cStrings,
	},
	"PowerShell": {
		LineComments: hashLineComments,
		MultilineComments: []MultilineCommentConfig{
			{
				Start:       []rune("<#"),
				End:         []rune("#>"),
				AtLineStart: false,
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
		LineComments: hashLineComments,
		MultilineComments: []MultilineCommentConfig{
			{
				Start:       []rune("\"\"\""),
				End:         []rune("\"\"\""),
				AtLineStart: false,
			},
		},
		Strings: cStrings,
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
				Start:       []rune("=begin"),
				End:         []rune("=end"),
				AtLineStart: true,
			},
		},
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune("%{"),
				End:        []rune{'}'},
				EscapeFunc: CharEscape('\\'),
			},
		},
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
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			},
		},
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
	"VBA": {
		LineComments: []LineCommentConfig{
			{
				Start: []rune{'\''},
			},
		},
		MultilineComments: nil,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			},
		},
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
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"XML": {
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

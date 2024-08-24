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

var LanguagesConfig = map[string]*Config{
	"Assembly": {
		LineCommentStart:            [][]rune{{';'}},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
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
		LineCommentStart:            [][]rune{[]rune("//")},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"C#": {
		LineCommentStart:            [][]rune{[]rune("//")},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"C++": {
		LineCommentStart:            [][]rune{[]rune("//")},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Clojure": {
		LineCommentStart:            [][]rune{{';'}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"CoffeeScript": {
		LineCommentStart:            [][]rune{{'#'}},
		MultilineCommentStart:       []rune("###"),
		MultilineCommentEnd:         []rune("###"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Dockerfile": {
		LineCommentStart:            [][]rune{{'#'}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Elixir": {
		LineCommentStart: [][]rune{{'#'}},
		// Support function documentation.
		// TODO(#1546): Support @moduledoc
		MultilineCommentStart:       []rune("@doc \"\"\""),
		MultilineCommentEnd:         []rune("\"\"\""),
		MultilineCommentAtLineStart: false,
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
		LineCommentStart:            [][]rune{{';'}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Erlang": {
		LineCommentStart:            [][]rune{{'%'}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Fortran": {
		LineCommentStart:            [][]rune{{'!'}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
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
		LineCommentStart:            [][]rune{{'!'}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
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
		LineCommentStart:            [][]rune{[]rune("//")},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
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
				Start:      []rune{'`'},
				End:        []rune{'`'},
				EscapeFunc: NoEscape,
			},
		},
	},
	"Go Module": {
		LineCommentStart:            [][]rune{[]rune("//")},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'`'},
				End:        []rune{'`'},
				EscapeFunc: NoEscape,
			},
		},
	},
	"Groovy": {
		LineCommentStart:            [][]rune{[]rune("//")},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
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
		LineCommentStart:            nil,
		MultilineCommentStart:       []rune("<!--"),
		MultilineCommentEnd:         []rune("--!>"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Haskell": {
		LineCommentStart:            [][]rune{[]rune("--")},
		MultilineCommentStart:       []rune("{-"),
		MultilineCommentEnd:         []rune("-}"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"JSON": {
		LineCommentStart: [][]rune{
			[]rune("//"),
			{'#'},
		},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Java": {
		LineCommentStart:            [][]rune{[]rune("//")},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"JavaScript": {
		LineCommentStart:            [][]rune{[]rune("//")},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Kotlin": {
		LineCommentStart:            [][]rune{[]rune("//")},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Lua": {
		LineCommentStart:            [][]rune{[]rune("--")},
		MultilineCommentStart:       []rune("--[["),
		MultilineCommentEnd:         []rune("--]]"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"MATLAB": {
		LineCommentStart:            [][]rune{{'%'}},
		MultilineCommentStart:       []rune("%{"),
		MultilineCommentEnd:         []rune("}%"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Makefile": {
		LineCommentStart:            [][]rune{{'#'}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Objective-C": {
		LineCommentStart:            [][]rune{[]rune("//")},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"PHP": {
		LineCommentStart: [][]rune{
			{'#'},
			[]rune("//"),
		},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Perl": {
		LineCommentStart:            [][]rune{{'#'}},
		MultilineCommentStart:       []rune{'='},
		MultilineCommentEnd:         []rune("=cut"),
		MultilineCommentAtLineStart: true,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"PowerShell": {
		LineCommentStart:            [][]rune{{'#'}},
		MultilineCommentStart:       []rune("<#"),
		MultilineCommentEnd:         []rune("#>"),
		MultilineCommentAtLineStart: false,
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
		LineCommentStart:            [][]rune{{'#'}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Python": {
		LineCommentStart:            [][]rune{{'#'}},
		MultilineCommentStart:       []rune("\"\"\""),
		MultilineCommentEnd:         []rune("\"\"\""),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"R": {
		LineCommentStart:            [][]rune{{'#'}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Ruby": {
		LineCommentStart:            [][]rune{{'#'}},
		MultilineCommentStart:       []rune("=begin"),
		MultilineCommentEnd:         []rune("=end"),
		MultilineCommentAtLineStart: true,
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
		LineCommentStart:            [][]rune{[]rune("//")},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"SQL": {
		LineCommentStart:            [][]rune{[]rune("--")},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
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
		LineCommentStart:            [][]rune{[]rune("//")},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Shell": {
		LineCommentStart:            [][]rune{{'#'}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Swift": {
		LineCommentStart:            [][]rune{[]rune("//")},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"TOML": {
		LineCommentStart:            [][]rune{{'#'}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"TeX": {
		LineCommentStart:            [][]rune{{'%'}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings:                     nil,
	},
	"TypeScript": {
		LineCommentStart:            [][]rune{[]rune("//")},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Unix Assembly": {
		LineCommentStart:            [][]rune{{';'}},
		MultilineCommentStart:       []rune("/*"),
		MultilineCommentEnd:         []rune("*/"),
		MultilineCommentAtLineStart: false,
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
		LineCommentStart:            [][]rune{{'\''}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Vim Script": {
		LineCommentStart:            [][]rune{{'"'}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"Visual Basic .NET": {
		LineCommentStart:            [][]rune{{'\''}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"XML": {
		LineCommentStart:            nil,
		MultilineCommentStart:       []rune("<!--"),
		MultilineCommentEnd:         []rune("--!>"),
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
	"YAML": {
		LineCommentStart:            [][]rune{{'#'}},
		MultilineCommentStart:       nil,
		MultilineCommentEnd:         nil,
		MultilineCommentAtLineStart: false,
		Strings: []StringConfig{
			{
				Start:      []rune{'"'},
				End:        []rune{'"'},
				EscapeFunc: CharEscape('\\'),
			}, {
				Start:      []rune{'\''},
				End:        []rune{'\''},
				EscapeFunc: CharEscape('\\'),
			},
		},
	},
}

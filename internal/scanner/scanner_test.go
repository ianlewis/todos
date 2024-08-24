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

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/text/encoding/ianaindex"

	"github.com/ianlewis/todos/internal/testutils"
)

var scannerTestCases = []*struct {
	// name is the test name. it is also the file name.
	name string

	// src is the source code.
	src string

	// config is language name of the scanner configuration to use.
	// TODO(#96): Use a *Config
	config string

	// comments are the comments expected to be found by the scanner.
	// TODO(#96): Use []*Comment and go-cmp
	comments []struct {
		// Text is the comment text.
		text string

		// line is the starting line where the comment is found.
		line int
	}
}{
	// Assembly
	{
		name: "line_comments.s",
		src: `; file comment

			; TODO is a function.
			TODO:
				ret ; Random comment`,
		config: "Unix Assembly",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "; file comment",
				line: 1,
			},
			{
				text: "; TODO is a function.",
				line: 3,
			},
			{
				text: "; Random comment",
				line: 5,
			},
		},
	},
	{
		name: "comments_in_string.s",
		src: `; file comment

			; TODO is a function.
			TODO:
				msg x "; Random comment",0
				msg y '; Random comment',0
				ret`,
		config: "Unix Assembly",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "; file comment",
				line: 1,
			},
			{
				text: "; TODO is a function.",
				line: 3,
			},
		},
	},
	{
		name: "escaped_string.s",
		src: `; file comment

			; TODO is a function.
			TODO:
				msg x "\"; Random comment",0
				msg y '\'; Random comment',0
				ret`,
		config: "Unix Assembly",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "; file comment",
				line: 1,
			},
			{
				text: "; TODO is a function.",
				line: 3,
			},
			// NOTE: Assembly doesn't seem to support escaped double or single quotes.
			{
				text: "; Random comment\",0",
				line: 5,
			},
			{
				text: "; Random comment',0",
				line: 6,
			},
		},
	},
	{
		name: "multi_line.s",
		src: `; file comment

			/*
			TODO is a function.
			*/
			TODO:
				ret ; Random comment
			/* extra comment */`,
		config: "Unix Assembly",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "; file comment",
				line: 1,
			},
			{
				text: "/*\n\t\t\tTODO is a function.\n\t\t\t*/",
				line: 3,
			},
			{
				text: "; Random comment",
				line: 7,
			},
			{
				text: "/* extra comment */",
				line: 8,
			},
		},
	},

	// Clojure
	{
		name: "line_comments.clj",
		src: `; file comment

			;; TODO is a function.
			(defn TODO [] (str "Hello") ) ; Random comment
			;;; extra comment ;;;`,
		config: "Clojure",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "; file comment",
				line: 1,
			},
			{
				text: ";; TODO is a function.",
				line: 3,
			},
			{
				text: "; Random comment",
				line: 4,
			},
			{
				text: ";;; extra comment ;;;",
				line: 5,
			},
		},
	},
	{
		name: "comments_in_string.clj",
		src: `; module comment

			; TODO is a function
			(defn TODO [] (str "; Random comment") )
			`,
		config: "Clojure",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "; module comment",
				line: 1,
			},
			{
				text: "; TODO is a function",
				line: 3,
			},
		},
	},
	{
		name: "escaped_string.clj",
		src: `; module comment

			; TODO is a function
			(defn TODO [] (str "\"; Random comment") )
			`,
		config: "Clojure",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "; module comment",
				line: 1,
			},
			{
				text: "; TODO is a function",
				line: 3,
			},
		},
	},

	// CoffeeScript
	{
		name: "line_comments.coffee",
		src: `# file comment

			###
			TODO is a function.
			###
			TODO = () ->
				return # Random comment
			### extra comment ###`,
		config: "CoffeeScript",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "###\n\t\t\tTODO is a function.\n\t\t\t###",
				line: 3,
			},
			{
				text: "# Random comment",
				line: 7,
			},
			{
				text: "### extra comment ###",
				line: 8,
			},
		},
	},
	{
		name: "comments_in_string.coffee",
		src: `# module comment

			# TODO is a function
			TODO = () ->
				console.log("# Random comment\n")
			`,
		config: "CoffeeScript",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# module comment",
				line: 1,
			},
			{
				text: "# TODO is a function",
				line: 3,
			},
		},
	},
	{
		name: "escaped_string.go",
		src: `# module comment

			# TODO is a function
			TODO = () ->
				console.log("\"# Random comment\n")
				console.log('\'# Random comment\n')
			`,
		config: "CoffeeScript",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# module comment",
				line: 1,
			},
			{
				text: "# TODO is a function",
				line: 3,
			},
		},
	},

	// Emacs Lisp
	{
		name: "line_comments.el",
		src: `; file comment

			;; TODO is a function.
			(defun TODO () "Hello") ; Random comment
			;;; extra comment ;;;`,
		config: "Emacs Lisp",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "; file comment",
				line: 1,
			},
			{
				text: ";; TODO is a function.",
				line: 3,
			},
			{
				text: "; Random comment",
				line: 4,
			},
			{
				text: ";;; extra comment ;;;",
				line: 5,
			},
		},
	},
	{
		name: "comments_in_string.el",
		src: `; module comment

			; TODO is a function
			(defun TODO () "; Random comment")
			`,
		config: "Emacs Lisp",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "; module comment",
				line: 1,
			},
			{
				text: "; TODO is a function",
				line: 3,
			},
		},
	},
	{
		name: "escaped_string.el",
		src: `; module comment

			; TODO is a function
			(defun TODO () "\"; Random comment")
			`,
		config: "Emacs Lisp",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "; module comment",
				line: 1,
			},
			{
				text: "; TODO is a function",
				line: 3,
			},
		},
	},

	// Erlang
	{
		name: "line_comments.erl",
		src: `module(foo)
			% module comment

			% TODO is a function
			TODO() ->
				io:fwrite("Hello World\n") % Random comment
			`,
		config: "Erlang",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "% module comment",
				line: 2,
			},
			{
				text: "% TODO is a function",
				line: 4,
			},
			{
				text: "% Random comment",
				line: 6,
			},
		},
	},
	{
		name: "comments_in_string.go",
		src: `module(foo)
			% module comment

			% TODO is a function
			TODO() ->
				io:fwrite("% Random comment\n")
			`,
		config: "Erlang",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "% module comment",
				line: 2,
			},
			{
				text: "% TODO is a function",
				line: 4,
			},
		},
	},
	{
		name: "escaped_string.go",
		src: `module(foo)
			% module comment

			% TODO is a function
			TODO() ->
				io:fwrite("\"% Random comment\n")
				io:fwrite('\'% Random comment\n')
			`,
		config: "Erlang",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "% module comment",
				line: 2,
			},
			{
				text: "% TODO is a function",
				line: 4,
			},
		},
	},

	// Fortran
	{
		name: "line_comments.f90",
		src: `! file comment

			! INPUT is a subroutine
			SUBROUTINE INPUT(X, Y, Z)
			REAL X,Y,Z
			PRINT *,'ENTER THREE NUMBERS => '
			READ *,X,Y,Z  ! Random comment
			RETURN
			END`,
		config: "Fortran Free Form",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "! file comment",
				line: 1,
			},
			{
				text: "! INPUT is a subroutine",
				line: 3,
			},
			{
				text: "! Random comment",
				line: 7,
			},
		},
	},
	{
		name: "comments_in_string.f90",
		src: `! file comment

			! INPUT is a subroutine
			SUBROUTINE INPUT(X, Y, Z)
			REAL X,Y,Z
			PRINT *,'ENTER THREE NUMBERS ! => '
			READ *,X,Y,Z  ! Random comment
			RETURN
			END`,
		config: "Fortran Free Form",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "! file comment",
				line: 1,
			},
			{
				text: "! INPUT is a subroutine",
				line: 3,
			},
			{
				text: "! Random comment",
				line: 7,
			},
		},
	},

	// Go
	{
		name: "line_comments.go",
		src: `package foo
			// package comment

			// TODO is a function.
			func TODO() {
				return // Random comment
			}`,
		config: "Go",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// package comment",
				line: 2,
			},
			{
				text: "// TODO is a function.",
				line: 4,
			},
			{
				text: "// Random comment",
				line: 6,
			},
		},
	},
	{
		name: "comments_in_string.go",
		src: `package foo
			// package comment

			// TODO is a function.
			func TODO() string {
				x := "// Random comment"
				y := '// Random comment'
				return x + y
			}`,
		config: "Go",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// package comment",
				line: 2,
			},
			{
				text: "// TODO is a function.",
				line: 4,
			},
		},
	},
	{
		name: "escaped_string.go",
		src: `package foo
			// package comment

			// TODO is a function.
			func TODO() string {
				x := "\"// Random comment"
				y := '\'// Random comment'
				return x + y
			}`,
		config: "Go",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// package comment",
				line: 2,
			},
			{
				text: "// TODO is a function.",
				line: 4,
			},
		},
	},
	{
		name: "raw_string.go",
		src: `package foo
			// package comment

			var z = ` + "`" + `
			// TODO is a function.
			` + "`" + `

			func TODO() string {
				// Random comment
				x := "\"// Random comment"
				y := '\'// Random comment'
				return x + y
			}`,
		config: "Go",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// package comment",
				line: 2,
			},
			{
				text: "// Random comment",
				line: 9,
			},
		},
	},
	{
		name: "multi_line.go",
		src: `package foo
			// package comment

			/*
			TODO is a function.
			*/
			func TODO() {
				return // Random comment
			}
			/* extra comment */`,
		config: "Go",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// package comment",
				line: 2,
			},
			{
				text: "/*\n\t\t\tTODO is a function.\n\t\t\t*/",
				line: 4,
			},
			{
				text: "// Random comment",
				line: 8,
			},
			{
				text: "/* extra comment */",
				line: 10,
			},
		},
	},

	// Groovy
	{
		name: "line_comments.groovy",
		src: `#!/usr/bin/env groovy
			// package comment

			// TODO is a function.
			def TODO() {
				return // Random comment
			}`,
		config: "Groovy",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// package comment",
				line: 2,
			},
			{
				text: "// TODO is a function.",
				line: 4,
			},
			{
				text: "// Random comment",
				line: 6,
			},
		},
	},
	{
		name: "comments_in_string.groovy",
		src: `#!/usr/bin/env groovy
			// package comment

			// TODO is a function.
			def TODO() String {
				x = "// Random comment"
				y = '// Random comment'
				return x + y
			}`,
		config: "Groovy",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// package comment",
				line: 2,
			},
			{
				text: "// TODO is a function.",
				line: 4,
			},
		},
	},
	{
		name: "escaped_string.groovy",
		src: `#!/usr/bin/env groovy
			// package comment

			// TODO is a function.
			def TODO() String {
				x = "\"// Random comment"
				y = '\'// Random comment'
				return x + y
			}`,
		config: "Groovy",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// package comment",
				line: 2,
			},
			{
				text: "// TODO is a function.",
				line: 4,
			},
		},
	},
	{
		name: "multiline_string.groovy",
		src: `#!/usr/bin/env groovy
			// package comment

			z = '''
			// TODO is a function.
			'''

			def TODO() String {
				// Random comment
				x = "\"// Random comment"
				y = '\'// Random comment'
				return x + y
			}`,
		config: "Groovy",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// package comment",
				line: 2,
			},
			{
				text: "// Random comment",
				line: 9,
			},
		},
	},
	{
		name: "multi_line.groovy",
		src: `#!/usr/bin/env groovy
			// package comment

			/*
			TODO is a function.
			*/
			def TODO() String {
				return // Random comment
			}
			/* extra comment */`,
		config: "Groovy",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// package comment",
				line: 2,
			},
			{
				text: "/*\n\t\t\tTODO is a function.\n\t\t\t*/",
				line: 4,
			},
			{
				text: "// Random comment",
				line: 8,
			},
			{
				text: "/* extra comment */",
				line: 10,
			},
		},
	},

	// Kotlin
	{
		name: "line_comments.kt",
		src: `// file comment

			// TODO is a function.
			fun TODO() {
				return // Random comment
			}`,
		config: "Kotlin",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// file comment",
				line: 1,
			},
			{
				text: "// TODO is a function.",
				line: 3,
			},
			{
				text: "// Random comment",
				line: 5,
			},
		},
	},
	{
		name: "multi_line.kt",
		src: `// file comment

			/*
			TODO is a function.
			*/
			fun TODO() {
				return // Random comment
			}
			/* extra comment */`,
		config: "Kotlin",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// file comment",
				line: 1,
			},
			{
				text: "/*\n\t\t\tTODO is a function.\n\t\t\t*/",
				line: 3,
			},
			{
				text: "// Random comment",
				line: 7,
			},
			{
				text: "/* extra comment */",
				line: 9,
			},
		},
	},
	{
		name: "comments_in_string.kt",
		src: `// file comment

			// TODO is a function.
			fun TODO():String {
				String x = "// Random comment";
				String y = "/* Random comment */";
				String z = '// Random comment';
				return x + y + z;
			}`,
		config: "Kotlin",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// file comment",
				line: 1,
			},
			{
				text: "// TODO is a function.",
				line: 3,
			},
		},
	},
	{
		name: "escaped_string.kt",
		src: `// file comment

			// TODO is a function.
			fun TODO():String {
				String x = "\"// Random comment";
				String y = '\'// Random comment';
				return x + y;
			}`,
		config: "Kotlin",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// file comment",
				line: 1,
			},
			{
				text: "// TODO is a function.",
				line: 3,
			},
		},
	},

	// PHP
	{
		name: "line_comments.php",
		src: `// file comment

			# TODO is a function.
			function TODO() {
				return // Random comment
			}`,
		config: "PHP",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// file comment",
				line: 1,
			},
			{
				text: "# TODO is a function.",
				line: 3,
			},
			{
				text: "// Random comment",
				line: 5,
			},
		},
	},

	// PowerShell
	{
		name: "line_comments.ps1",
		src: `# file comment

			# TODO is a function.
			function TODO {
				Write-Output "TODO" # Random comment
			}`,
		config: "PowerShell",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# TODO is a function.",
				line: 3,
			},
			{
				text: "# Random comment",
				line: 5,
			},
		},
	},
	{
		name: "comments_in_string.ps1",
		src: `# file comment

			# TODO is a function.
			function TODO {
				Write-Output "# TODO"
				Write-Output '# TODO'
			}`,
		config: "PowerShell",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# TODO is a function.",
				line: 3,
			},
		},
	},
	{
		name: "escaped_string.ps1",
		src: `# file comment

			# TODO is a function.
			function TODO {
				Write-Output "` + "`" + `"# Random comment"
				Write-Output '` + "`" + `'# Random comment'
			}`,
		config: "PowerShell",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# TODO is a function.",
				line: 3,
			},
		},
	},
	{
		name: "multi_line.ps1",
		src: `# file comment

			<#
			TODO is a function.
			#>
			function TODO {
				Write-Output "TODO" # Random comment
			}
			<# extra comment #>`,
		config: "PowerShell",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "<#\n\t\t\tTODO is a function.\n\t\t\t#>",
				line: 3,
			},
			{
				text: "# Random comment",
				line: 7,
			},
			{
				text: "<# extra comment #>",
				line: 9,
			},
		},
	},

	// Puppet
	{
		name: "line_comments.pp",
		src: `# file comment

			# todo is a service
			service { 'todo':
			  name      => $service_name, # Random comment
			  ensure    => running,
			  enable    => true,
			}
		`,

		config: "Puppet",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# todo is a service",
				line: 3,
			},
			{
				text: "# Random comment",
				line: 5,
			},
		},
	},
	{
		name: "comments_in_string.pp",
		src: `# file comment

			# hello.txt
			file { '/tmp/hello.txt':
			  ensure  => file,
			  content => "# some comment",
			}
		`,

		config: "Puppet",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# hello.txt",
				line: 3,
			},
		},
	},
	{
		name: "escaped_string.pp",
		src: `# file comment

			# hello.txt
			file { '/tmp/hello.txt':
			  ensure  => file,
			  content => "\"# some comment",
			}
		`,

		config: "Puppet",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# hello.txt",
				line: 3,
			},
		},
	},

	// Haskell
	{
		name: "line_comments.hs",
		src: `-- file comment

			-- TODO is a function.
			TODO = do
				putStrLn "fizzbuzz" -- Random comment
			`,
		config: "Haskell",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "-- file comment",
				line: 1,
			},
			{
				text: "-- TODO is a function.",
				line: 3,
			},
			{
				text: "-- Random comment",
				line: 5,
			},
		},
	},
	{
		name: "comments_in_string.hs",
		src: `-- file comment

			-- TODO is a function.
			TODO = do
				x = "-- Random comment";
				y = '-- Random comment';
			`,
		config: "Haskell",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "-- file comment",
				line: 1,
			},
			{
				text: "-- TODO is a function.",
				line: 3,
			},
		},
	},
	{
		name: "escaped_string.hs",
		src: `-- file comment

			-- TODO is a function.
			TODO = do
				x = "\"-- Random comment";
				y = '\'-- Random comment';
				x + y
			}`,
		config: "Haskell",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "-- file comment",
				line: 1,
			},
			{
				text: "-- TODO is a function.",
				line: 3,
			},
		},
	},
	{
		name: "multi_line_string.hs",
		src: `-- file comment

			z = = " \
			-- TODO is a function. \
			";

			TODO = do
				-- Random comment
			`,
		config: "Haskell",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "-- file comment",
				line: 1,
			},
			{
				text: "-- Random comment",
				line: 8,
			},
		},
	},
	{
		name: "multi_line.hs",
		src: `-- file comment

			{-
			TODO is a function.
			-}
			TODO = do
				putStrLn "fizzbuzz" -- Random comment
			}
			{- extra comment -}`,
		config: "Haskell",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "-- file comment",
				line: 1,
			},
			{
				text: "{-\n\t\t\tTODO is a function.\n\t\t\t-}",
				line: 3,
			},
			{
				text: "-- Random comment",
				line: 7,
			},
			{
				text: "{- extra comment -}",
				line: 9,
			},
		},
	},

	// Lua
	{
		name: "line_comments.lua",
		src: `-- package comment

			-- TODO is a function.
			function TODO() {
				return -- Random comment
			}`,
		config: "Lua",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "-- package comment",
				line: 1,
			},
			{
				text: "-- TODO is a function.",
				line: 3,
			},
			{
				text: "-- Random comment",
				line: 5,
			},
		},
	},
	{
		name: "comments_in_string.lua",
		src: `-- package comment

			-- TODO is a function.
			func TODO() {
				x = "-- Random comment"
				y = '-- Random comment'
				return x + y
			}`,
		config: "Lua",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "-- package comment",
				line: 1,
			},
			{
				text: "-- TODO is a function.",
				line: 3,
			},
		},
	},
	{
		name: "escaped_string.lua",
		src: `-- package comment

			-- TODO is a function.
			func TODO() {
				x = "\"-- Random comment"
				y = '\'-- Random comment'
				return x + y
			}`,
		config: "Lua",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "-- package comment",
				line: 1,
			},
			{
				text: "-- TODO is a function.",
				line: 3,
			},
		},
	},
	{
		name: "multi_line.lua",
		src: `-- package comment

			--[[
			TODO is a function.
			--]]
			func TODO() {
				return -- Random comment
			}
			--[[ extra comment --]]`,
		config: "Lua",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "-- package comment",
				line: 1,
			},
			{
				text: "--[[\n\t\t\tTODO is a function.\n\t\t\t--]]",
				line: 3,
			},
			{
				text: "-- Random comment",
				line: 7,
			},
			{
				text: "--[[ extra comment --]]",
				line: 9,
			},
		},
	},

	// MATLAB
	{
		name: "line_comments.m",
		src: `% file comment

			% TODO is an array.
			TODO = [1, 2, 3, 4]
			% end of file`,
		config: "MATLAB",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "% file comment",
				line: 1,
			},
			{
				text: "% TODO is an array.",
				line: 3,
			},
			{
				text: "% end of file",
				line: 5,
			},
		},
	},
	{
		name: "comments_in_string.m",
		src: `% file comment

			% TODO is a string.
			TODO = "-- Random comment"
			% end of file`,
		config: "MATLAB",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "% file comment",
				line: 1,
			},
			{
				text: "% TODO is a string.",
				line: 3,
			},
			{
				text: "% end of file",
				line: 5,
			},
		},
	},
	{
		name: "escaped_string.m",
		src: `% file comment

			% TODO is a string.
			TODO = "\"% Random comment"
			% end of file`,
		config: "MATLAB",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "% file comment",
				line: 1,
			},
			{
				text: "% TODO is a string.",
				line: 3,
			},
			{
				text: "% end of file",
				line: 5,
			},
		},
	},
	{
		name: "multi_line.m",
		src: `% file comment

			%{
			TODO is a string.
			}%
			TODO = "% Random comment"
			% end of file`,
		config: "MATLAB",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "% file comment",
				line: 1,
			},
			{
				text: "%{\n\t\t\tTODO is a string.\n\t\t\t}%",
				line: 3,
			},
			{
				text: "% end of file",
				line: 7,
			},
		},
	},

	// Python
	{
		name: "raw_string.py",
		src: `# file comment

			def foo():
				# Random comment
				x = "\"# Random comment"
				y = '\'# Random comment'
				return x + y
			`,
		config: "Python",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# Random comment",
				line: 4,
			},
		},
	},
	{
		name: "multi_line.py",
		src: `# file comment

			"""
			TODO is a function.
			"""

			def foo():
				# Random comment
				x = "\"# Random comment"
				y = '\'# Random comment'
				return x + y
			`,
		config: "Python",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "\"\"\"\n\t\t\tTODO is a function.\n\t\t\t\"\"\"",
				line: 3,
			},
			{
				text: "# Random comment",
				line: 8,
			},
		},
	},

	// R
	{
		name: "line_comments.r",
		src: `# file comment

			# TODO is a function
			TODO <- function() {
				print("Hello World") # Random comment
			`,
		config: "R",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# TODO is a function",
				line: 3,
			},
			{
				text: "# Random comment",
				line: 5,
			},
		},
	},
	{
		name: "comments_in_string.r",
		src: `# file comment

			# TODO is a function
			TODO <- function() {
				print("# Random comment")
				print('# Random comment')
			`,
		config: "R",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# TODO is a function",
				line: 3,
			},
		},
	},
	{
		name: "escaped_string.r",
		src: `# file comment

			# TODO is a function
			TODO <- function() {
				print("\"# Random comment")
				print('\'# Random comment')
			`,
		config: "R",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# TODO is a function",
				line: 3,
			},
		},
	},

	// Ruby
	{
		name: "raw_string.rb",
		src: `# file comment

			z = %{
			# TODO is a function.
			}

			def foo()
				# Random comment
				x = "\"# Random comment x"
				y = '\'# Random comment y'
				return x + y
			end`,
		config: "Ruby",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# Random comment",
				line: 8,
			},
		},
	},
	{
		name: "multi_line.rb",
		src: `# file comment

=begin
TODO is a function.
=end

			def foo()
				# Random comment
				x = "\"# Random comment x"
				y = '\'# Random comment y'
				return x + y
			end`,
		config: "Ruby",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "=begin\nTODO is a function.\n=end",
				line: 3,
			},
			{
				text: "# Random comment",
				line: 8,
			},
		},
	},
	{
		name: "multi_line_not_line_start.rb",
		src: `# file comment

	=begin
TODO is a function.
=end

			def foo()
				# Random comment
				x = "\"# Random comment x"
				y = '\'# Random comment y'
				return x + y
			end`,
		config: "Ruby",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# Random comment",
				line: 8,
			},
		},
	},
	{
		name: "multi_line_end.rb",
		src: `# file comment

=begin
TODO is a function.
	=end
=end

			def foo()
				# Random comment
				x = "\"# Random comment x"
				y = '\'# Random comment y'
				return x + y
			end`,
		config: "Ruby",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "=begin\nTODO is a function.\n\t=end\n=end",
				line: 3,
			},
			{
				text: "# Random comment",
				line: 9,
			},
		},
	},

	// Rust
	{
		name: "line_comments.rs",
		src: `// file comment

			// TODO is a function.
			fn TODO() {
				println!("fizzbuzz"); // Random comment
			}`,
		config: "Rust",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// file comment",
				line: 1,
			},
			{
				text: "// TODO is a function.",
				line: 3,
			},
			{
				text: "// Random comment",
				line: 5,
			},
		},
	},
	{
		name: "comments_in_string.rs",
		src: `// file comment

			// TODO is a function.
			fn TODO() {
				let x: String = "// Random comment";
				let y: String = '// Random comment';
				x + y
			}`,
		config: "Rust",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// file comment",
				line: 1,
			},
			{
				text: "// TODO is a function.",
				line: 3,
			},
		},
	},
	{
		name: "escaped_string.rs",
		src: `// file comment

			// TODO is a function.
			fn TODO() {
				let x: String "\"// Random comment";
				let y: String '\'// Random comment';
				x + y
			}`,
		config: "Rust",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// file comment",
				line: 1,
			},
			{
				text: "// TODO is a function.",
				line: 3,
			},
		},
	},
	{
		name: "multi_line_string.rs",
		src: `// file comment

			let z: String = "
			// TODO is a function.
			";

			fn TODO() -> String {
				// Random comment
				let x: String = "\"// Random comment";
				let y: String = '\'// Random comment';
				x + y
			}`,
		config: "Rust",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// file comment",
				line: 1,
			},
			{
				text: "// Random comment",
				line: 8,
			},
		},
	},
	{
		name: "multi_line.rs",
		src: `// file comment

			/*
			TODO is a function.
			*/
			fn TODO() {
				println!("fizzbuzz"); // Random comment
			}
			/* extra comment */`,
		config: "Rust",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "// file comment",
				line: 1,
			},
			{
				text: "/*\n\t\t\tTODO is a function.\n\t\t\t*/",
				line: 3,
			},
			{
				text: "// Random comment",
				line: 7,
			},
			{
				text: "/* extra comment */",
				line: 9,
			},
		},
	},

	// Shell
	{
		name: "line_comments.sh",
		src: `#!/bin/bash

			echo "foo" # Random comment`,
		config: "Shell",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "#!/bin/bash",
				line: 1,
			},
			{
				text: "# Random comment",
				line: 3,
			},
		},
	},
	{
		name: "comments_in_string.sh",
		src: `#!/bin/bash

			echo "foo" # Random comment
			echo "#My comment"`,
		config: "Shell",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "#!/bin/bash",
				line: 1,
			},
			{
				text: "# Random comment",
				line: 3,
			},
		},
	},
	{
		name: "raw_strings.sh",
		src: `#!/bin/bash

			echo "foo" # Random comment
			echo '#My comment'`,
		config: "Shell",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "#!/bin/bash",
				line: 1,
			},
			{
				text: "# Random comment",
				line: 3,
			},
		},
	},

	// SQL
	{
		name: "line_comments.sql",
		src: `-- file comment

			-- TODO is a table.
			SELECT * from TODO
			WHERE
				x = "TODO" -- Random comment
			LIMIT 1;`,
		config: "SQL",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "-- file comment",
				line: 1,
			},
			{
				text: "-- TODO is a table.",
				line: 3,
			},
			{
				text: "-- Random comment",
				line: 6,
			},
		},
	},
	{
		name: "comments_in_string.sql",
		src: `-- file comment

			-- TODO is a table.
			SELECT * from TODO
			WHERE
				x = "-- Random comment" AND
				y = "-- Random comment"
			LIMIT 1;`,
		config: "SQL",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "-- file comment",
				line: 1,
			},
			{
				text: "-- TODO is a table.",
				line: 3,
			},
		},
	},
	{
		name: "escaped_string.rs",
		src: `-- file comment

			-- TODO is a table.
			SELECT * from TODO
			WHERE
				x = 'foo '' -- Random comment' AND
				y = "foo \"-- Random comment"
			LIMIT 1;`,
		config: "SQL",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "-- file comment",
				line: 1,
			},
			{
				text: "-- TODO is a table.",
				line: 3,
			},
			{
				text: "-- Random comment\"",
				line: 7,
			},
		},
	},
	{
		name: "multi_line_string.rs",
		src: `-- file comment

			-- TODO is a table.
			SELECT * from TODO
			WHERE
				x = "
				-- Random comment
				" AND
				y = "-- Random comment"
			LIMIT 1;`,
		config: "SQL",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "-- file comment",
				line: 1,
			},
			{
				text: "-- TODO is a table.",
				line: 3,
			},
		},
	},
	{
		name: "multi_line.sql",
		src: `-- file comment

			/*
			TODO is a function.
			*/
			SELECT * from TODO
			LIMIT 1;`,
		config: "SQL",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "-- file comment",
				line: 1,
			},
			{
				text: "/*\n\t\t\tTODO is a function.\n\t\t\t*/",
				line: 3,
			},
		},
	},

	// TeX
	{
		name: "line_comments.tex",
		src: `% file comment

			% Random comment
			This is some text.`,
		config: "TeX",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "% file comment",
				line: 1,
			},
			{
				text: "% Random comment",
				line: 3,
			},
		},
	},

	// VBA
	{
		name: "line_comments.vba",
		src: `' A "Hello, World!" program in Visual Basic.
Module Hello
  Sub Main()
      MsgBox("Hello, World!") ' Display message on computer screen.
  End Sub
End Module`,
		config: "VBA",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "' A \"Hello, World!\" program in Visual Basic.",
				line: 1,
			},
			{
				text: "' Display message on computer screen.",
				line: 4,
			},
		},
	},

	// Visual Basic .NET
	{
		name: "line_comments.vba",
		src: `' A "Hello, World!" program in Visual Basic.
Imports System 'System is a Namespace
Module Hello_Program
    Sub Main()
        Console.WriteLine("Hello, Welcome to the world of VB.NET")
        Console.WriteLine("Press any key to continue...")
        Console.ReadKey()
    End Sub
End Module`,
		config: "Visual Basic .NET",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "' A \"Hello, World!\" program in Visual Basic.",
				line: 1,
			},
			{
				text: "'System is a Namespace",
				line: 2,
			},
		},
	},

	// Vim Script
	{
		name: "line_comments.vim",
		src: `" file comment

			" TODO is a function.
			function TODO()
				return "Hello" " Random comment
			endfunction
			" extra comment`,
		config: "Vim Script",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "\" file comment",
				line: 1,
			},
			{
				text: "\" TODO is a function.",
				line: 3,
			},
			{
				text: "\" Random comment",
				line: 5,
			},
			{
				text: "\" extra comment",
				line: 7,
			},
		},
	},
	{
		name: "escaped_string.vim",
		src: `" module comment

			" TODO is a function
			function TODO()
				return "\" Random comment"
			endfunction`,
		config: "Vim Script",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "\" module comment",
				line: 1,
			},
			{
				text: "\" TODO is a function",
				line: 3,
			},
		},
	},
	{
		name: "closed_string_comment.vim",
		src: `" file comment

			" TODO is a function."
			function TODO()
				return "Hello" " Random comment"
			endfunction
			" extra comment`,
		config: "Vim Script",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "\" file comment",
				line: 1,
			},
			// TODO(#1540): Read this closed string as a comment.
			// {
			//	text: "\" TODO is a function.\"",
			//	line: 3,
			// },
			// TODO(#1540): Read this closed string as a comment.
			// {
			//	text: "\" Random comment\"",
			//	line: 5,
			// },
			{
				text: "\" extra comment",
				line: 7,
			},
		},
	},
}

func TestCommentScanner(t *testing.T) {
	t.Parallel()

	for i := range scannerTestCases {
		tc := scannerTestCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := New(strings.NewReader(tc.src), LanguagesConfig[tc.config])

			var comments []*Comment
			for s.Scan() {
				comments = append(comments, s.Next())
			}
			if err := s.Err(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if want, got := len(tc.comments), len(comments); want != got {
				t.Fatalf("unexpected # of comments, want: %v, got: %v\n\ncomments:\n%q", want, got, comments)
			}

			for i := range tc.comments {
				if want, got := tc.comments[i].text, comments[i].String(); want != got {
					t.Errorf("unexpected text, want: %q, got: %q", want, got)
				}

				if want, got := tc.comments[i].line, comments[i].Line; want != got {
					t.Errorf("unexpected line, want: %d, got: %d", want, got)
				}
			}
		})
	}
}

var scannerRegressionTestCases = []*struct {
	name   string
	src    string
	config *Config

	expectedComments []*Comment
	expectedErr      error
}{
	{
		name: "last_line.go",
		src:  `// last line`,
		config: &Config{
			LineCommentStart: [][]rune{[]rune("//")},
		},
		expectedComments: []*Comment{
			{
				Text: "// last line",
				Line: 1,
			},
		},
	},
	{
		name: "double_escape_1538.foo",
		src:  `x = ''''''  % foo''`,
		config: &Config{
			LineCommentStart: [][]rune{[]rune("%")},
			Strings: []StringConfig{
				{
					Start:      []rune("''"),
					End:        []rune("''"),
					EscapeFunc: DoubleEscape,
				},
			},
		},

		expectedComments: nil,
	},
}

func TestCommentScanner_regression(t *testing.T) {
	t.Parallel()

	for i := range scannerRegressionTestCases {
		tc := scannerRegressionTestCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New(strings.NewReader(tc.src), tc.config)

			var comments []*Comment
			for s.Scan() {
				comments = append(comments, s.Next())
			}

			{
				got, want := s.Err(), tc.expectedErr
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("unexpected error (-want +got):\n%s", diff)
				}
			}

			{
				got, want := comments, tc.expectedComments
				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("unexpected comments (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func BenchmarkCommentScanner(b *testing.B) {
	for i := range scannerTestCases {
		tc := scannerTestCases[i]
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				s := New(strings.NewReader(tc.src), LanguagesConfig[tc.config])
				for s.Scan() {
				}
			}
		})
	}
}

var loaderTestCases = []struct {
	name    string
	charset string
	src     []byte

	scanCharset    string
	expectedConfig string
	err            error
}{
	// Emacs Lisp
	{
		name:        "line_comments.el",
		charset:     "ISO-8859-1",
		scanCharset: "ISO-8859-1",
		src: []byte(`; file comment

			;; TODO is a function.
			(defun TODO () "Hello") ; Random comment
			;;; extra comment ;;;`),
		expectedConfig: "Emacs Lisp",
	},

	// Go
	{
		name:        "ascii.go",
		charset:     "ISO-8859-1",
		scanCharset: "ISO-8859-1",
		src: []byte(`package foo
			// package comment

			// TODO is a function.
			func TODO() {
				return // Random comment
			}`),
		expectedConfig: "Go",
	},
	{
		name:        "utf8.go",
		charset:     "UTF-8",
		scanCharset: "UTF-8",
		src: []byte(`package foo
			// Hello, 世界

			// TODO is a function.
			func TODO() {
				return // Random comment
			}`),
		expectedConfig: "Go",
	},
	{
		name:        "shift_jis.go",
		charset:     "SHIFT_JIS",
		scanCharset: "SHIFT_JIS",
		src: []byte(`package foo
			// Hello, 世界

			// TODO is a function.
			func TODO() {
				return // Random comment
			}`),
		expectedConfig: "Go",
	},
	{
		name:           "gb18030.go",
		src:            []byte{255, 255, 255, 255, 255, 255, 250},
		scanCharset:    "detect",
		expectedConfig: "Go",
	},
	{
		name:        "zeros.go",
		src:         []byte{0, 0, 0, 0, 0, 0},
		scanCharset: "detect",
		// Detected as binary
		expectedConfig: "",
	},
	{
		name:           "detect_by_filename.go",
		src:            []byte{},
		scanCharset:    "UTF-8",
		expectedConfig: "Go",
	},
	// TODO: enry doesn't do a good job with it's classifier.
	// {
	//	name: "detect_by_contents.foo",
	//	src: []byte(`package foo
	//		// package comment

	//		// TODO is a function.
	//		func TODO() {
	//			return // Random comment
	//		}`),
	//	scanCharset:    "UTF-8",
	//	expectedConfig: "Go",
	// },
	{
		name: "binary.exe",
		// NOTE: Control codes rarely seen in text.
		scanCharset: "UTF-8",
		src:         []byte{1, 2, 3, 4, 5},
	},
	{
		name:           "unsupported_lang.coq",
		src:            []byte{},
		scanCharset:    "UTF-8",
		expectedConfig: "", // nil
	},
	{
		name: "typescript_is_not_xml.ts",
		src: []byte(`
			function wrapInArray(obj: string | string[]) {
				if (typeof obj === "string") {
					return [obj];
				}
				return obj;
			}
		`),
		scanCharset:    "UTF-8",
		expectedConfig: "TypeScript",
	},
	{
		name: "qt_translation_file.ts",
		src: []byte(`
			<!DOCTYPE TS><TS>
			<context>
				<name>QPushButton</name>
				<message>
					<source>Hello world!</source>
					<translation type="unfinished"></translation>
				</message>
			</context>
			</TS>
		`),
		scanCharset:    "UTF-8",
		expectedConfig: "XML",
	},
	{
		name: "texfile.tex",
		src: []byte(`
		% file comment
		\documentclass{jsarticle}
		\begin{document}
		This is some tex.
		\end{document}
		`),
		scanCharset:    "UTF-8",
		expectedConfig: "TeX",
	},
}

func TestFromFile(t *testing.T) {
	t.Parallel()

	for i := range loaderTestCases {
		tc := loaderTestCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Create a temporary file.
			// NOTE: File extensions are used as hints so the file name must be part of the suffix.
			f := testutils.Must(os.CreateTemp("", fmt.Sprintf("*.%s", tc.name)))
			defer os.Remove(f.Name())

			var w io.Writer
			w = f
			if tc.charset != "" {
				e := testutils.Must(ianaindex.IANA.Encoding(tc.charset))
				w = e.NewEncoder().Writer(f)
			}
			_ = testutils.Must(w.Write(tc.src))
			_ = testutils.Must(f.Seek(0, io.SeekStart))

			s, err := FromFile(f, tc.scanCharset)
			if got, want := err, tc.err; got != nil {
				if !errors.Is(got, want) {
					t.Fatalf("unexpected err, got: %v, want: %v", got, want)
				}
				return
			} else if want != nil {
				t.Fatalf("unexpected err, got: %v, want: %v", got, want)
			}

			var config *Config
			if s != nil {
				config = s.Config()
			}
			if got, want := config, LanguagesConfig[tc.expectedConfig]; got != want {
				t.Fatalf("unexpected config, got: %#v, want: %#v", got, want)
			}
		})
	}
}

func TestFromBytes(t *testing.T) {
	t.Parallel()

	for i := range loaderTestCases {
		tc := loaderTestCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			text := tc.src
			if tc.charset != "" {
				e := testutils.Must(ianaindex.IANA.Encoding(tc.charset))
				text = testutils.Must(e.NewDecoder().Bytes(tc.src))
			}

			s, err := FromBytes(tc.name, text, tc.scanCharset)
			if got, want := err, tc.err; got != nil {
				if !errors.Is(got, want) {
					t.Fatalf("unexpected err, got: %v, want: %v", got, want)
				}
				return
			} else if want != nil {
				t.Fatalf("unexpected err, got: %v, want: %v", got, want)
			}

			var config *Config
			if s != nil {
				config = s.Config()
			}
			if got, want := config, LanguagesConfig[tc.expectedConfig]; got != want {
				t.Fatalf("unexpected config, got: %#v, want: %#v", got, want)
			}
		})
	}
}

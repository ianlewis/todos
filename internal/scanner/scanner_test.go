// Copyright 2023 Google LLC
// Copyright 2025 Ian Lewis, Marcin Wiśniowski, Steffen Raabe
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
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"golang.org/x/text/encoding/ianaindex"

	"github.com/ianlewis/todos/internal/testutils"
)

//nolint:gochecknoglobals // allow global table-driven tests.
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

	// CODEOWNERS
	{
		name: "CODEOWNERS",
		src: `# file comment
		* @ianlewis # Not a comment
		# another comment`,
		config: "CODEOWNERS",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# another comment",
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

	// Dart
	{
		name: "line_comments.dart",
		src: `/// Module documentation

            /* block comment */

            /// Function documentation
            int fibonacci(int n) {
                if (n == 0 || n == 1) return n;
                // fibonacci is recursive.
                return fibonacci(n - 1) + fibonacci(n - 2);
            }

            var block_comment = "/* block comment in string */ is a block comment";
            var line_comment = "line comment: // line comment in string";
            var doc_comment = "doc comment: /// doc comment in string";
            var result = fibonacci(20);`,
		config: "Dart",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "/// Module documentation",
				line: 1,
			},
			{
				text: "/* block comment */",
				line: 3,
			},
			{
				text: "/// Function documentation",
				line: 5,
			},
			{
				text: "// fibonacci is recursive.",
				line: 8,
			},
		},
	},

	// Elixir
	{
		name: "line_comments.ex",
		//nolint:dupword // Allow duplicate words in test code.
		src: `# module comment
			defmodule Math do
				# TODO is a function.
				def TODO(a, b) do
					a + b # Random comment
				end
			end`,
		config: "Elixir",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# module comment",
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
		name: "comments_in_string.ex",
		//nolint:dupword // Allow duplicate words in test code.
		src: `# module comment
			defmodule Math do
				# TODO is a function.
				def TODO(a, b) do
					"# Random comment"
				end
			end`,
		config: "Elixir",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# module comment",
				line: 1,
			},
			{
				text: "# TODO is a function.",
				line: 3,
			},
			// NOTE: No '# Random comment'
		},
	},
	{
		name: "comments_in_multiline_string.ex",
		//nolint:dupword // Allow duplicate words in test code.
		src: `# module comment
			defmodule Math do
				# TODO is a function.
				def TODO(a, b) do
					"""
					# Random comment
					"""
				end
			end`,
		config: "Elixir",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# module comment",
				line: 1,
			},
			{
				text: "# TODO is a function.",
				line: 3,
			},
			// NOTE: No '# Random comment'
		},
	},
	{
		name: "comments_in_charlist.ex",
		//nolint:dupword // Allow duplicate words in test code.
		src: `# module comment
			defmodule Math do
				# TODO is a function.
				def TODO(a, b) do
					'# Random comment'
				end
			end`,
		config: "Elixir",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# module comment",
				line: 1,
			},
			{
				text: "# TODO is a function.",
				line: 3,
			},
			// NOTE: No '# Random comment'
		},
	},
	{
		name: "comments_in_multiline_charlist.ex",
		//nolint:dupword // Allow duplicate words in test code.
		src: `# module comment
			defmodule Math do
				# TODO is a function.
				def TODO(a, b) do
					'''
					# Random comment
					'''
				end
			end`,
		config: "Elixir",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# module comment",
				line: 1,
			},
			{
				text: "# TODO is a function.",
				line: 3,
			},
			// NOTE: No '# Random comment'
		},
	},
	{
		name: "moduledoc_comment.ex",
		//nolint:dupword // Allow duplicate words in test code.
		src: `@moduledoc """
			module comment
			"""
			defmodule Math do
				def TODO(a, b) do
					a + b
				end
			end`,
		config: "Elixir",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "@moduledoc \"\"\"\n\t\t\tmodule comment\n\t\t\t\"\"\"",
				line: 1,
			},
		},
	},
	{
		name: "doc_comment.ex",
		//nolint:dupword // Allow duplicate words in test code.
		src: `# module comment
			defmodule Math do
				@doc """
				TODO is a function.
				"""
				def TODO(a, b) do
					a + b
				end
			end`,
		config: "Elixir",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# module comment",
				line: 1,
			},
			{
				text: "@doc \"\"\"\n\t\t\t\tTODO is a function.\n\t\t\t\t\"\"\"",
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

	// GraphQL
	{
		name: "multi_line.graphql",
		src: `"""
Author of questions and answers in a website
"""
type Author {
  # ... username is the author name , this is an example of a dropped comment
  username: String! @id
  """
  The questions submitted by this author
  """
  questions: [Question] @hasInverse(field: author)
  """
  The answers submitted by this author
  """
  answers: [Answer] @hasInverse(field: author)
}`,
		config: "GraphQL",
		comments: []struct {
			text string
			line int
		}{
			{
				text: `"""
Author of questions and answers in a website
"""`,
				line: 1,
			},
			{
				text: "# ... username is the author name , this is an example of a dropped comment",
				line: 5,
			},
			{
				text: `"""
  The questions submitted by this author
  """`,
				line: 7,
			},
			{
				text: `"""
  The answers submitted by this author
  """`,
				line: 11,
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

	// HCL
	{
		name: "line_comments.tf",
		src: `# file comment

        # vm instance
        resource "aws_instance" "example" {
            ami = "abc123 # comment in string"

            network_interface {
              // network interface options
            }
        }`,
		config: "HCL",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# vm instance",
				line: 3,
			},
			// NOTE: Comment in string not included
			{
				text: "// network interface options",
				line: 8,
			},
		},
	},

	// HTML
	{
		name: "multi_line.html",
		src: `<!-- file comment -->
		<html attribute="<!-- not a comment -->">
			<body attribute='<!-- also not a comment -->'>
				<div>Hello World!</div>
			</body>
		</html>
		<!-- extra comment -->`,
		config: "HTML",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "<!-- file comment -->",
				line: 1,
			},
			{
				text: "<!-- extra comment -->",
				line: 7,
			},
		},
	},

	// HTML+ERB
	{
		name: "multi_line.erb",
		src: `<!-- file comment -->
		<html attribute="<!-- not a comment -->">
			<%# erb comment %>
			<body attribute='<!-- also not a comment -->'>
				<% message = "Hello World!" %>
				<div><%= message %></div>
			</body>
		</html>
		<!-- extra comment -->`,
		config: "HTML+ERB",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "<!-- file comment -->",
				line: 1,
			},
			{
				text: "<%# erb comment %>",
				line: 3,
			},
			{
				text: "<!-- extra comment -->",
				line: 9,
			},
		},
	},

	// Ignore List (.gitignore)
	{
		name: "Ignore List",
		src: `# file comment
		/some/file/path # Not a comment
		# another comment`,
		config: "CODEOWNERS",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: "# another comment",
				line: 3,
			},
		},
	},

	// Julia
	{
		name: "comments.jl",
		src: `# module comment

"""
TODOs in doc comments are not currently included
TODO: foo
"""
function mandelbrot(a)
    z = 0
    for i=1:50
        z = z^2 + a
    end
    return z # this is another comment
end

#=
This is a comment.
#=
This is a nested comment.
=#
for y=1.0:-0.05:-1.0
    for x=-2.0:0.0315:0.5
        abs(mandelbrot(complex(x, y))) < 2 ? print("*") : print(" ")
    end
    println()
end
=#`,
		config: "Julia",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# module comment",
				line: 1,
			},
			{
				text: "# this is another comment",
				line: 12,
			},
			{
				text: `#=
This is a comment.
#=
This is a nested comment.
=#
for y=1.0:-0.05:-1.0
    for x=-2.0:0.0315:0.5
        abs(mandelbrot(complex(x, y))) < 2 ? print("*") : print(" ")
    end
    println()
end
=#`,
				line: 15,
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

	// Pascal
	{
		name: "line_comments.pas",
		src: `// file comment

			// TODO is a function.
			function TODO(num1, num2: integer): integer;
			var
				result: integer;

			begin
				if (num1 > num2) then
					result := num1
				else
					result := num2;
				max := result;
			end; // Random comment`,
		config: "Pascal",
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
				line: 14,
			},
		},
	},
	{
		name: "multi_line.pas",
		src: `(* file comment *)

			{
			TODO is a function.
			}
			function TODO(num1, num2: integer): integer;
			var
				result: integer;

			begin
				if (num1 > num2) then
					result := num1
				else
					result := num2;
				max := result;
			end; (* Random comment *)`,
		config: "Pascal",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "(* file comment *)",
				line: 1,
			},
			{
				text: "{\n\t\t\tTODO is a function.\n\t\t\t}",
				line: 3,
			},
			{
				text: "(* Random comment *)",
				line: 16,
			},
		},
	},
	{
		name: "comments_in_string.pas",
		src: `(* file comment *)

			{
			TODO is a function.
			}
			function TODO(num1, num2: integer): integer;
			var
				result: integer;

			begin
				if (num1 > num2) then
					result := 'num1 (* Not a comment *)'
				else
					result := "num2 { Also not a comment }";
				max := result;
			end; (* Random comment *)`,
		config: "Pascal",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "(* file comment *)",
				line: 1,
			},
			{
				text: "{\n\t\t\tTODO is a function.\n\t\t\t}",
				line: 3,
			},
			{
				text: "(* Random comment *)",
				line: 16,
			},
		},
	},
	{
		name: "escaped_string.pas",
		src: `(* file comment *)

			{
			TODO is a function.
			}
			function TODO(num1, num2: integer): integer;
			var
				result: integer;

			begin
				unused := '''(* not a comment *)'''
				if (num1 > num2) then
					result := 'num1 ''(* Not a comment *)'' { not a comment here either}'
				else
					result := "num2 ""{ Also not a comment }"" (* this is also not a comment *)";
				max := result;
			end; (* Random comment *)`,
		config: "Pascal",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "(* file comment *)",
				line: 1,
			},
			{
				text: "{\n\t\t\tTODO is a function.\n\t\t\t}",
				line: 3,
			},
			{
				text: "(* Random comment *)",
				line: 17,
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
		//nolint:dupword // Allow duplicate words in test code.
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

	// Markdown
	{
		name: "block_comments.md",
		src: `# Title

## Header

This is some text

<!-- this is a comment -->

This is more text.<!-- this is another comment -->

` + "```" + `
<!-- comments in code blocks don't count -->
` + "```" + `

` + "`" + `<!-- comments in inline code doesn't count -->` + "`",
		config: "Markdown",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "<!-- this is a comment -->",
				line: 7,
			},
			{
				text: "<!-- this is another comment -->",
				line: 9,
			},
		},
	},

	// Nix
	{
		name: "comments.nix",
		src: `# file comment

/*
Block comments
can span multiple lines.
*/ "hello"
        `,

		config: "Nix",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "# file comment",
				line: 1,
			},
			{
				text: `/*
Block comments
can span multiple lines.
*/`,
				line: 3,
			},
		},
	},

	// OCaml
	{
		name: "comments.ml",
		src: `(* single line comment *)

(* multiple line comment, commenting out part of a program:
let f = function
  | 'A'..'Z' -> "Uppercase"
*)`,
		config: "OCaml",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "(* single line comment *)",
				line: 1,
			},
			{
				text: `(* multiple line comment, commenting out part of a program:
let f = function
  | 'A'..'Z' -> "Uppercase"
*)`,
				line: 3,
			},
		},
	},
	{
		name: "nested_comments.ml",
		src: `(* multiple line comment, commenting out part of a program, and containing a
nested comment:
let f = function
  | 'A'..'Z' -> "Uppercase"
(* Add other cases later... (* more nesting *) *)
*)`,
		config: "OCaml",
		comments: []struct {
			text string
			line int
		}{
			{
				text: `(* multiple line comment, commenting out part of a program, and containing a
nested comment:
let f = function
  | 'A'..'Z' -> "Uppercase"
(* Add other cases later... (* more nesting *) *)
*)`,
				line: 1,
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
				x
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
				x
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
				x
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
		name: "lifetime_specifier.rs",
		src: `// file comment

			const A: &'static str = "some string";
			// another file comment
			const B: &'static str = "some other string";

			// yet another file comment`,
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
				text: "// another file comment",
				line: 4,
			},
			{
				text: "// yet another file comment",
				line: 7,
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
	// Svelte
	{
		name: "Component.svelte",
		src: `<!-- file comment -->
		<script lang="ts">
			// line comment
			/*
			block comment
			*/
		</script>
		<style>
			/*
			block comment
			*/
		</style>
		<div attribute="<!-- not a comment -->">
			<Component attribute='<!-- also not a comment -->'>
				<div>Hello World!</div>
			</Component>
		</div>
		<!-- extra comment -->`,
		config: "Svelte",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "<!-- file comment -->",
				line: 1,
			},
			{
				text: "// line comment",
				line: 3,
			},
			{
				text: "/*\n\t\t\tblock comment\n\t\t\t*/",
				line: 4,
			},
			{
				text: "/*\n\t\t\tblock comment\n\t\t\t*/",
				line: 9,
			},
			{
				text: "<!-- extra comment -->",
				line: 18,
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
		//nolint:dupword // Allow duplicate words in test code.
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

	// XML
	{
		name: "block_comments.xml",
		src: `<?xml version="1.0" encoding="UTF-8"?>
			<message>
				<to>Joe</to><!-- To: header -->
				<from>Susan</from><!-- From: header -->
				<body attr="<!-- comment in string -->">Hello World</body>
			</message>`,
		config: "XML",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "<!-- To: header -->",
				line: 3,
			},
			{
				text: "<!-- From: header -->",
				line: 4,
			},
		},
	},

	// XSLT
	{
		name: "block_comments.xslt",
		src: `<?xml version="1.0" encoding="UTF-8"?>
			<xsl:stylesheet version="1.0"
			xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
			<xsl:template match="/">
			  <html>
			  <body attr="<!-- comment in string -->">
				<h2>CD Collection</h2>
				<table border="1">
				  <tr bgcolor="#9acd32">
				  <th>Title</th>
				  <th>Artist</th>
				  <th>Price</th>
				</tr>
				<!-- for each CD -->
				<xsl:for-each select="catalog/cd">
				<xsl:if test="price>10"><!-- if price greater than 10 -->
				  <tr>
					<td><xsl:value-of select="title"/></td>
					<td><xsl:value-of select="artist"/></td>
					<td><xsl:value-of select="price"/></td>
				  </tr>
				</xsl:if>
				</xsl:for-each>
				</table>
			  </body>
			  </html>
			</xsl:template>
			</xsl:stylesheet>`,
		config: "XSLT",
		comments: []struct {
			text string
			line int
		}{
			{
				text: "<!-- for each CD -->",
				line: 14,
			},
			{
				text: "<!-- if price greater than 10 -->",
				line: 16,
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

//nolint:gochecknoglobals // allow global table-driven tests.
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
			LineComments: cLineComments,
		},
		expectedComments: []*Comment{
			{
				Text:       "// last line",
				Line:       1,
				LineConfig: &cLineComments[0],
			},
		},
	},
	{
		name: "double_escape_1538.foo",
		src:  `x = ''''''  % foo''`,
		config: &Config{
			LineComments: []LineCommentConfig{
				{
					Start: []rune("%"),
				},
			},
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
			for b.Loop() {
				s := New(strings.NewReader(tc.src), LanguagesConfig[tc.config])
				for s.Scan() {
				}
			}
		})
	}
}

//nolint:gochecknoglobals // allow global table-driven tests.
var loaderTestCases = []struct {
	name    string
	charset string
	src     []byte

	scanCharset    string
	expectedConfig string
	err            error
}{
	// Dart
	{
		name: "dart_detection.dart",
		src: []byte(`/// Module documentation

            /* block comment */

            /// Function documentation
            int fibonacci(int n) {
                if (n == 0 || n == 1) return n;
                // fibonacci is recursive.
                return fibonacci(n - 1) + fibonacci(n - 2);
            }

            var block_comment = "/* block comment in string */ is a block comment";
            var line_comment = "line comment: // line comment in string";
            var doc_comment = "doc comment: /// doc comment in string";
            var result = fibonacci(20);`),
		scanCharset:    "UTF-8",
		expectedConfig: "Dart",
	},

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
		name:        "binary.exe",
		src:         []byte{0, 0, 0, 0, 0, 0},
		scanCharset: "detect",
		// Detected as binary
		expectedConfig: "",
		err:            ErrBinaryFile,
	},
	{
		name:           "detect_by_filename.go",
		src:            []byte{},
		scanCharset:    "UTF-8",
		expectedConfig: "Go",
	},

	// Julia
	{
		name: "mandelbrot.jl",
		src: []byte(`
function mandelbrot(a)
    z = 0
    for i=1:50
        z = z^2 + a
    end
    return z # this is another comment
end
`),
		scanCharset:    "UTF-8",
		expectedConfig: "Julia",
	},

	// Nix
	{
		name: "attribute_set.nix",
		src: []byte(`{
  x = 123;
  text = "Hello";
  y = f { bla = 456; };
}`),
		scanCharset:    "UTF-8",
		expectedConfig: "Nix",
	},

	// OCaml
	{
		name: "oca.ml",
		src: []byte(`(* this is a comment *)

print_endline "hello world"

let () = print_endline "hello world"
`),
		scanCharset:    "UTF-8",
		expectedConfig: "OCaml",
	},

	// XML
	{
		name: "block_comments.xml",
		src: []byte(`
			<?xml version="1.0" encoding="UTF-8"?>
			<message>
				<to>Joe</to><!-- To: header -->
				<from>Susan</from><!-- From: header -->
				<body attr="<!-- comment in string -->">Hello World</body>
			</message>`),
		scanCharset:    "UTF-8",
		expectedConfig: "XML",
	},

	{
		name: "block_comments.xslt",
		src: []byte(`
			<?xml version="1.0" encoding="UTF-8"?>
			<xsl:stylesheet version="1.0"
			xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
			<xsl:template match="/">
			  <html>
			  <body>
				<h2>CD Collection</h2>
				<table border="1">
				  <tr bgcolor="#9acd32">
				  <th>Title</th>
				  <th>Artist</th>
				  <th>Price</th>
				</tr>
				<xsl:for-each select="catalog/cd">
				<xsl:if test="price>10">
				  <tr>
					<td><xsl:value-of select="title"/></td>
					<td><xsl:value-of select="artist"/></td>
					<td><xsl:value-of select="price"/></td>
				  </tr>
				</xsl:if>
				</xsl:for-each>
				</table>
			  </body>
			  </html>
			</xsl:template>
			</xsl:stylesheet>`),
		scanCharset:    "UTF-8",
		expectedConfig: "XSLT",
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
		name:           "unsupported_lang.coq",
		src:            []byte{},
		scanCharset:    "UTF-8",
		expectedConfig: "", // nil
		err:            ErrUnsupportedLanguage,
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
		testCase := loaderTestCases[i]

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Create a temporary file.
			// NOTE: File extensions are used as hints so the file name must be part of the suffix.
			f := testutils.Must(os.CreateTemp(t.TempDir(), "*."+testCase.name))
			defer f.Close()

			var w io.Writer

			w = f

			if testCase.charset != "" {
				e := testutils.Must(ianaindex.IANA.Encoding(testCase.charset))
				w = e.NewEncoder().Writer(f)
			}

			_ = testutils.Must(w.Write(testCase.src))
			_ = testutils.Must(f.Seek(0, io.SeekStart))

			s, err := FromFile(f, testCase.scanCharset)
			if diff := cmp.Diff(testCase.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("FromFile: unexpected err (-want +got):\n%s", diff)
			}

			var config *Config
			if s != nil {
				config = s.Config()
			}

			if got, want := config, LanguagesConfig[testCase.expectedConfig]; got != want {
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
			if diff := cmp.Diff(tc.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("FromBytes: unexpected err (-want +got):\n%s", diff)
			}

			var config *Config
			if s != nil {
				config = s.Config()
			}

			if got, want := config, LanguagesConfig[tc.expectedConfig]; got != want {
				t.Fatalf("unexpected LanguagesConfig, got: %#v, want: %#v", got, want)
			}
		})
	}
}

func TestOverlappingConfig(t *testing.T) {
	t.Parallel()

	type expectedComment struct {
		Text string
		Line int
	}

	testCases := map[string]struct {
		src              string
		config           *Config
		expectedComments []expectedComment
	}{
		"longer multi-line comment and string": {
			src: `x = #this is a string#
			###this is a comment###`,
			config: &Config{
				MultilineComments: []MultilineCommentConfig{
					{
						Start: []rune("###"),
						End:   []rune("###"),
					},
				},
				Strings: []StringConfig{
					{
						Start:      []rune("#"),
						End:        []rune("#"),
						EscapeFunc: NoEscape,
					},
				},
			},
			expectedComments: []expectedComment{
				{
					Text: "###this is a comment###",
					Line: 2,
				},
			},
		},
		"multi-line comment and longer string": {
			src: `x = ###this is a string###
			#this is a comment#`,
			config: &Config{
				MultilineComments: []MultilineCommentConfig{
					{
						Start: []rune("#"),
						End:   []rune("#"),
					},
				},
				Strings: []StringConfig{
					{
						Start:      []rune("###"),
						End:        []rune("###"),
						EscapeFunc: NoEscape,
					},
				},
			},
			expectedComments: []expectedComment{
				{
					Text: "#this is a comment#",
					Line: 2,
				},
			},
		},
		"longer line comments and string": {
			src: `x = #this is a string#
			### this is a comment # foo`,
			config: &Config{
				LineComments: []LineCommentConfig{
					{
						Start: []rune("###"),
					},
				},
				Strings: []StringConfig{
					{
						Start:      []rune("#"),
						End:        []rune("#"),
						EscapeFunc: NoEscape,
					},
				},
			},
			expectedComments: []expectedComment{
				{
					Text: "### this is a comment # foo",
					Line: 2,
				},
			},
		},
		"line comments and longer string": {
			src: `x = ###this is a string###
			# this is a comment`,
			config: &Config{
				LineComments: []LineCommentConfig{
					{
						Start: []rune("#"),
					},
				},
				Strings: []StringConfig{
					{
						Start:      []rune("###"),
						End:        []rune("###"),
						EscapeFunc: NoEscape,
					},
				},
			},
			expectedComments: []expectedComment{
				{
					Text: "# this is a comment",
					Line: 2,
				},
			},
		},
		"line comments and string": {
			src: `x = #this is a string#; meh
			# this is a comment`,
			config: &Config{
				LineComments: []LineCommentConfig{
					{
						Start: []rune("#"),
					},
				},
				Strings: []StringConfig{
					{
						Start:      []rune("#"),
						End:        []rune("#"),
						EscapeFunc: NoEscape,
					},
				},
			},
			expectedComments: []expectedComment{
				{
					Text: "# this is a comment",
					Line: 2,
				},
			},
		},
		"longer line comment and multi-line comment": {
			src: `#
			multi-line comment
			#
			x = "some code";
			### this is a comment ###; foo`,
			config: &Config{
				LineComments: []LineCommentConfig{
					{
						Start: []rune("###"),
					},
				},
				MultilineComments: []MultilineCommentConfig{
					{
						Start: []rune("#"),
						End:   []rune("#"),
					},
				},
			},
			expectedComments: []expectedComment{
				{
					Text: "#\n\t\t\tmulti-line comment\n\t\t\t#",
					Line: 1,
				},
				{
					Text: "### this is a comment ###; foo",
					Line: 5,
				},
			},
		},
		"line comment and longer multi-line comment": {
			src: `###
			multi-line comment
			###
			x = "some code";
			# this is a comment ###; foo`,
			config: &Config{
				LineComments: []LineCommentConfig{
					{
						Start: []rune("#"),
					},
				},
				MultilineComments: []MultilineCommentConfig{
					{
						Start: []rune("###"),
						End:   []rune("###"),
					},
				},
			},
			expectedComments: []expectedComment{
				{
					Text: "###\n\t\t\tmulti-line comment\n\t\t\t###",
					Line: 1,
				},
				{
					Text: "# this is a comment ###; foo",
					Line: 5,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := New(strings.NewReader(tc.src), tc.config)

			var comments []expectedComment

			for s.Scan() {
				c := s.Next()
				comments = append(comments, expectedComment{
					Text: c.Text,
					Line: c.Line,
				})
			}

			if err := s.Err(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			want, got := tc.expectedComments, comments
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("unexpected comments (-want +got):\n%s", diff)
			}
		})
	}
}

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
	"strings"
	"testing"
)

var testCases = []*struct {
	name     string
	src      string
	config   Config
	comments []struct {
		text string
		line int
	}
}{
	{
		name: "line_comments.go",
		src: `package foo
			// package comment

			// TODO is a function.
			func TODO() {
				return // Random comment
			}`,
		config: GoConfig,
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
		name: "line_comments.php",
		src: `// file comment

			# TODO is a function.
			function TODO() {
				return // Random comment
			}`,
		config: PHPConfig,
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
	{
		name: "comments_in_string.go",
		src: `package foo
			// package comment

			// TODO is a function.
			func TODO() {
				x := "// Random comment"
				y := '// Random comment'
				return x + y
			}`,
		config: GoConfig,
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
			func TODO() {
				x := "\"// Random comment"
				y := '\'// Random comment'
				return x + y
			}`,
		config: GoConfig,
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

			func TODO() {
				// Random comment
				x := "\"// Random comment"
				y := '\'// Random comment'
				return x + y
			}`,
		config: GoConfig,
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
		config: GoConfig,
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
		config: RubyConfig,
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
		name: "raw_string.py",
		src: `# file comment

			def foo():
				# Random comment
				x = "\"# Random comment"
				y = '\'# Random comment'
				return x + y
			`,
		config: PythonConfig,
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
		config: PythonConfig,
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
	{
		name: "line_comments.sh",
		src: `#!/bin/bash

			echo "foo" # Random comment`,
		config: ShellConfig,
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
		config: ShellConfig,
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
		config: ShellConfig,
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
}

func TestCommentScanner(t *testing.T) {
	t.Parallel()

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := New(strings.NewReader(tc.src), &tc.config)

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

func BenchmarkCommentScanner(b *testing.B) {
	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			s := New(strings.NewReader(tc.src), &tc.config)
			for s.Scan() {
			}
		})
	}
}

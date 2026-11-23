// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitdiff

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-cmp/cmp"

	"github.com/ianlewis/todos/internal/testutils"
)

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		diff string
		want map[string][]LineRange
	}{
		{
			name: "empty diff",
			diff: "",
			want: map[string][]LineRange{},
		},
		{
			name: "new file adds every line",
			diff: "diff --git a/src/a.ts b/src/a.ts\n" +
				"new file mode 100644\n" +
				"index 0000000..1111111\n" +
				"--- /dev/null\n" +
				"+++ b/src/a.ts\n" +
				"@@ -0,0 +1,3 @@\n" +
				"+import x;\n" +
				"+// TODO(alice): impl\n" +
				"+export {};\n",
			want: map[string][]LineRange{
				"src/a.ts": {{Start: 1, End: 3}},
			},
		},
		{
			name: "single modified line",
			diff: "diff --git a/a.ts b/a.ts\n" +
				"--- a/a.ts\n" +
				"+++ b/a.ts\n" +
				"@@ -10,1 +10,1 @@\n" +
				"-// TODO(alice): old\n" +
				"+// TODO(alice): new\n",
			want: map[string][]LineRange{
				"a.ts": {{Start: 10, End: 10}},
			},
		},
		{
			name: "context lines are not new",
			diff: "diff --git a/a.ts b/a.ts\n" +
				"--- a/a.ts\n" +
				"+++ b/a.ts\n" +
				"@@ -1 +1,2 @@\n" +
				" // TODO(alice): old\n" +
				"+// TODO(bob): new\n",
			want: map[string][]LineRange{
				"a.ts": {{Start: 2, End: 2}},
			},
		},
		{
			name: "hunk with more additions than removals",
			diff: "diff --git a/a.ts b/a.ts\n" +
				"--- a/a.ts\n" +
				"+++ b/a.ts\n" +
				"@@ -1,2 +1,4 @@\n" +
				"-a\n" +
				"-b\n" +
				"+a\n" +
				"+b\n" +
				"+// TODO(alice): c\n" +
				"+// TODO(alice): d\n",
			want: map[string][]LineRange{
				"a.ts": {{Start: 1, End: 4}},
			},
		},
		{
			name: "pure deletion contributes no range",
			diff: "diff --git a/a.ts b/a.ts\n" +
				"--- a/a.ts\n" +
				"+++ b/a.ts\n" +
				"@@ -5,3 +4,0 @@\n" +
				"-// TODO(alice): gone\n" +
				"-b\n" +
				"-c\n",
			want: map[string][]LineRange{},
		},
		{
			name: "deleted file contributes no range",
			diff: "diff --git a/a.ts b/a.ts\n" +
				"deleted file mode 100644\n" +
				"--- a/a.ts\n" +
				"+++ /dev/null\n" +
				"@@ -1,2 +0,0 @@\n" +
				"-// TODO(alice): gone\n" +
				"-b\n",
			want: map[string][]LineRange{},
		},
		{
			name: "hunk header without explicit count means one line",
			diff: "diff --git a/a.ts b/a.ts\n" +
				"--- a/a.ts\n" +
				"+++ b/a.ts\n" +
				"@@ -5 +5 @@\n" +
				"-// TODO(alice): old\n" +
				"+// TODO(alice): new\n",
			want: map[string][]LineRange{
				"a.ts": {{Start: 5, End: 5}},
			},
		},
		{
			name: "multiple hunks in one file",
			diff: "diff --git a/a.ts b/a.ts\n" +
				"--- a/a.ts\n" +
				"+++ b/a.ts\n" +
				"@@ -1,0 +2,2 @@\n" +
				"+// TODO(alice): one\n" +
				"+x\n" +
				"@@ -10,0 +20,1 @@\n" +
				"+// TODO(alice): two\n",
			want: map[string][]LineRange{
				"a.ts": {
					{Start: 2, End: 3},
					{Start: 20, End: 20},
				},
			},
		},
		{
			name: "multiple files",
			diff: "diff --git a/a.ts b/a.ts\n" +
				"--- a/a.ts\n" +
				"+++ b/a.ts\n" +
				"@@ -1,0 +1,1 @@\n" +
				"+// TODO(alice): a\n" +
				"diff --git a/dir/b.py b/dir/b.py\n" +
				"--- a/dir/b.py\n" +
				"+++ b/dir/b.py\n" +
				"@@ -3,0 +3,1 @@\n" +
				"+# TODO(alice): b\n",
			want: map[string][]LineRange{
				"a.ts":     {{Start: 1, End: 1}},
				"dir/b.py": {{Start: 3, End: 3}},
			},
		},
		{
			name: "CRLF line endings",
			diff: "diff --git a/a.ts b/a.ts\r\n" +
				"--- a/a.ts\r\n" +
				"+++ b/a.ts\r\n" +
				"@@ -2,0 +2,1 @@\r\n" +
				"+// TODO(alice): crlf\r\n",
			want: map[string][]LineRange{
				"a.ts": {{Start: 2, End: 2}},
			},
		},
		{
			name: "binary file contributes no range",
			diff: "diff --git a/img/a.png b/img/a.png\n" +
				"index 1111111..2222222 100644\n" +
				"Binary files a/img/a.png and b/img/a.png differ\n",
			want: map[string][]LineRange{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(test.diff)
			if err != nil {
				t.Fatalf("Parse() returned unexpected error: %v", err)
			}

			if diff := cmp.Diff(test.want, got.files); diff != "" {
				t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestChangesContains(t *testing.T) {
	t.Parallel()

	changes := &Changes{
		files: map[string][]LineRange{
			"src/a.ts": {
				{Start: 5, End: 7},
				{Start: 20, End: 20},
			},
		},
	}

	tests := []struct {
		name     string
		fileName string
		root     string
		line     int
		want     bool
	}{
		{
			name:     "line inside first range",
			fileName: "src/a.ts",
			line:     5,
			want:     true,
		},
		{
			name:     "line inside first range upper bound",
			fileName: "src/a.ts",
			line:     7,
			want:     true,
		},
		{
			name:     "line just outside first range",
			fileName: "src/a.ts",
			line:     8,
			want:     false,
		},
		{
			name:     "line inside second range",
			fileName: "src/a.ts",
			line:     20,
			want:     true,
		},
		{
			name:     "line before any range",
			fileName: "src/a.ts",
			line:     4,
			want:     false,
		},
		{
			name:     "unknown file",
			fileName: "src/b.ts",
			line:     5,
			want:     false,
		},
		{
			name:     "dot slash prefix is normalised",
			fileName: "./src/a.ts",
			line:     6,
			want:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := changes.Contains(test.fileName, test.root, test.line); got != test.want {
				t.Errorf("Contains(%q, %q, %d) = %v, want %v",
					test.fileName, test.root, test.line, got, test.want)
			}
		})
	}
}

func TestChangesContainsWindowsSeparators(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "windows" {
		t.Skip("backslash path separators only occur on Windows")
	}

	changes := &Changes{
		files: map[string][]LineRange{"src/a.ts": {{Start: 5, End: 7}}},
	}

	if !changes.Contains(`src\a.ts`, "", 6) {
		t.Error("Contains() with backslash separators = false, want true")
	}
}

func TestChangesContainsAbsolutePath(t *testing.T) {
	t.Parallel()

	// filepath.IsAbs reports false for rooted-but-driveless paths on Windows,
	// so each platform needs a path that is genuinely absolute there.
	root := "/repo"
	fileName := "/repo/src/a.ts"

	if runtime.GOOS == "windows" {
		root = `C:\repo`
		fileName = `C:\repo\src\a.ts`
	}

	changes := &Changes{
		files: map[string][]LineRange{"src/a.ts": {{Start: 5, End: 7}}},
	}

	if !changes.Contains(fileName, root, 6) {
		t.Errorf("Contains(%q, %q, 6) = false, want true", fileName, root)
	}
}

func TestChangesEmpty(t *testing.T) {
	t.Parallel()

	if !NewChanges().Empty() {
		t.Error("NewChanges().Empty() = false, want true")
	}

	changes, err := Parse("diff --git a/a.ts b/a.ts\n+++ b/a.ts\n@@ -1,0 +1,1 @@\n+x\n")
	if err != nil {
		t.Fatalf("Parse() returned unexpected error: %v", err)
	}

	if changes.Empty() {
		t.Error("changes with one added line reported Empty() = true, want false")
	}
}

func TestAddedLines(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	repo := testutils.NewTestRepo(t, dir, "Test Author", "test@example.com", []*testutils.File{
		{
			Path:     "src/a.ts",
			Contents: []byte("// TODO(alice): old debt\n"),
			Mode:     0o600,
		},
	}, nil)

	base, err := repo.Repository().Head()
	if err != nil {
		t.Fatalf("resolve HEAD: %v", err)
	}

	worktree, err := repo.Repository().Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}

	// Append a line to an existing file and add a brand new one.
	updated := filepath.Join(dir, "src", "a.ts")
	if err := os.WriteFile(updated, []byte("// TODO(alice): old debt\n// TODO(bob): new one\n"), 0o600); err != nil {
		t.Fatalf("write %s: %v", updated, err)
	}
	if _, err := worktree.Add("src/a.ts"); err != nil {
		t.Fatalf("add src/a.ts: %v", err)
	}

	created := filepath.Join(dir, "b.py")
	if err := os.WriteFile(created, []byte("# TODO(carol): from scratch\n"), 0o600); err != nil {
		t.Fatalf("write %s: %v", created, err)
	}
	if _, err := worktree.Add("b.py"); err != nil {
		t.Fatalf("add b.py: %v", err)
	}

	sig := &object.Signature{Name: "Test Author", Email: "test@example.com", When: time.Now()}
	if _, err := worktree.Commit("add a todo", &git.CommitOptions{Author: sig}); err != nil {
		t.Fatalf("commit: %v", err)
	}

	changes, err := AddedLines(dir, base.Hash().String())
	if err != nil {
		t.Fatalf("AddedLines() returned unexpected error: %v", err)
	}

	tests := []struct {
		name string
		file string
		line int
		want bool
	}{
		{name: "existing line untouched", file: "src/a.ts", line: 1, want: false},
		{name: "appended line is new", file: "src/a.ts", line: 2, want: true},
		{name: "new file line is new", file: "b.py", line: 1, want: true},
		{name: "unknown file", file: "c.go", line: 1, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := changes.Contains(test.file, dir, test.line); got != test.want {
				t.Errorf("Contains(%q, %q, %d) = %v, want %v",
					test.file, dir, test.line, got, test.want)
			}
		})
	}
}

func TestAddedLinesRequiresBase(t *testing.T) {
	t.Parallel()

	if _, err := AddedLines(".", ""); err == nil {
		t.Error("AddedLines() with empty base returned nil error, want error")
	}
}

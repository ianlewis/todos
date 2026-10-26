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

// Package gitdiff reports which lines a change added or modified.
//
// It exists to support the --fail-on-new flag. Without it, `todos` fails
// whenever a TODO exists anywhere in the tree, which makes the tool unusable
// on repositories that already contain TODOs -- which is to say, on most
// repositories. With it, the exit status reflects only TODOs introduced by the
// change under review, so a project can adopt the check without first having
// to clear its backlog.
package gitdiff

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// LineRange is an inclusive range of 1-based line numbers.
type LineRange struct {
	Start int
	End   int
}

// Contains reports whether line falls within the range.
func (r LineRange) Contains(line int) bool {
	return line >= r.Start && line <= r.End
}

// Changes records the line ranges added or modified by a change, keyed by
// repository-relative, slash-separated file path.
type Changes struct {
	files map[string][]LineRange
}

// NewChanges returns an empty Changes value.
func NewChanges() *Changes {
	return &Changes{files: make(map[string][]LineRange)}
}

// Empty reports whether the change added no lines at all.
func (c *Changes) Empty() bool {
	return len(c.files) == 0
}

// Contains reports whether the given line of the given file was added or
// modified by the change.
//
// fileName is in whatever form the caller has it -- typically the path
// reported by the walker, which mirrors how the user invoked the tool, so it
// may be relative to the working directory or absolute. root is the working
// tree root used to relativise absolute paths; it may be empty, in which case
// no relativisation is attempted.
func (c *Changes) Contains(fileName, root string, line int) bool {
	ranges, ok := c.files[normalizePath(fileName, root)]
	if !ok {
		return false
	}
	for _, r := range ranges {
		if r.Contains(line) {
			return true
		}
	}
	return false
}

// normalizePath converts a walker-reported file name into the
// repository-relative, slash-separated form used as the map key.
func normalizePath(fileName, root string) string {
	path := fileName
	if root != "" && filepath.IsAbs(path) {
		if rel, err := filepath.Rel(root, path); err == nil {
			path = rel
		}
	}
	return filepath.ToSlash(filepath.Clean(path))
}

var (
	// hunkHeaderRE matches a unified diff hunk header, capturing the start
	// line of the post-image and its optional line count.
	hunkHeaderRE = regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@`)

	// fileHeaderRE matches the "+++ " line that introduces a file's hunks.
	fileHeaderRE = regexp.MustCompile(`^\+\+\+ (?:b/)?(.+)$`)
)

// Parse parses a unified diff as produced by `git diff` and returns the ranges
// of lines added or modified in each file.
//
// Only lines introduced by a "+" line count. The counts in a hunk header
// describe the whole post-image of the hunk, context lines included, so using
// them directly would report untouched lines as new.
//
// Deleted files, binary files and hunks consisting only of deletions
// contribute no ranges.
func Parse(diff string) (*Changes, error) {
	changes := NewChanges()

	var (
		current string
		newLine int
		inHunk  bool
	)

	for _, raw := range strings.Split(diff, "\n") {
		line := strings.TrimSuffix(raw, "\r")

		if strings.HasPrefix(line, "diff --git ") {
			current = ""
			inHunk = false
			continue
		}

		if m := fileHeaderRE.FindStringSubmatch(line); m != nil {
			current = strings.TrimSpace(m[1])
			if current == "/dev/null" {
				// The file was deleted; nothing in it can be new.
				current = ""
			}
			inHunk = false
			continue
		}

		if strings.HasPrefix(line, "--- ") {
			inHunk = false
			continue
		}

		if m := hunkHeaderRE.FindStringSubmatch(line); m != nil {
			if current == "" {
				inHunk = false
				continue
			}

			start, err := strconv.Atoi(m[1])
			if err != nil {
				return nil, fmt.Errorf("parse hunk start line %q: %w", m[1], err)
			}

			newLine = start
			inHunk = true
			continue
		}

		if !inHunk || current == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "+"):
			changes.files[current] = append(changes.files[current], LineRange{Start: newLine, End: newLine})
			newLine++
		case strings.HasPrefix(line, " "):
			newLine++
		case strings.HasPrefix(line, "-"):
			// Removed line: it does not advance the post-image counter.
		case strings.HasPrefix(line, `\`):
			// "\ No newline at end of file" marker: not a content line.
		default:
			// Anything else ends the hunk.
			inHunk = false
		}
	}

	for name, ranges := range changes.files {
		changes.files[name] = coalesce(ranges)
	}

	return changes, nil
}

// coalesce sorts ranges and merges those that touch or overlap.
func coalesce(ranges []LineRange) []LineRange {
	if len(ranges) <= 1 {
		return ranges
	}

	sort.Slice(ranges, func(i, j int) bool { return ranges[i].Start < ranges[j].Start })

	merged := ranges[:1]
	for _, r := range ranges[1:] {
		last := &merged[len(merged)-1]
		if r.Start <= last.End+1 {
			if r.End > last.End {
				last.End = r.End
			}
			continue
		}
		merged = append(merged, r)
	}

	return merged
}

// AddedLines opens the git repository containing dir and returns the lines
// added or modified between base and the current HEAD.
//
// base may be any revision git accepts, including a full commit SHA.
//
// NOTE: this performs a two-dot (base..HEAD) comparison. When the base branch
// has advanced since the change branched off, that would attribute unrelated
// upstream commits to the change under review. Callers running in CI should
// therefore pass the merge base, e.g. the output of
//
//	git merge-base origin/main HEAD
//
// rather than the base branch name itself.
func AddedLines(dir, base string) (*Changes, error) {
	if base == "" {
		return nil, fmt.Errorf("base revision is required")
	}

	repo, err := git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, fmt.Errorf("open git repository: %w", err)
	}

	baseHash, err := repo.ResolveRevision(plumbing.Revision(base))
	if err != nil {
		return nil, fmt.Errorf("resolve base revision %q: %w", base, err)
	}

	baseCommit, err := repo.CommitObject(*baseHash)
	if err != nil {
		return nil, fmt.Errorf("load base commit %q: %w", base, err)
	}

	headRef, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("resolve HEAD: %w", err)
	}

	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("load HEAD commit: %w", err)
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("load base tree: %w", err)
	}

	headTree, err := headCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("load HEAD tree: %w", err)
	}

	changes, err := object.DiffTree(baseTree, headTree)
	if err != nil {
		return nil, fmt.Errorf("diff trees: %w", err)
	}

	patch, err := changes.Patch()
	if err != nil {
		return nil, fmt.Errorf("build patch: %w", err)
	}

	return Parse(patch.String())
}

// Root returns the absolute path of the working tree root of the git
// repository containing dir, or the empty string if it cannot be determined.
func Root(dir string) string {
	repo, err := git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return ""
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return ""
	}

	return worktree.Filesystem.Root()
}

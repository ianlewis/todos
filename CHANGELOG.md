# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## Changed in X.Y.Z

- Leading whitespaces is now trimmed from TODOS in multi-line comments.
- The `--include-hidden` option was replaced with the `--exclude-hidden`
  option and including hidden files was made the default.
- An `--include-vcs` option was added and the VCS directories `.git`, `.hg`,
  and `.svn` are skipped by default.

## Fixed in X.Y.Z

- Hidden, Vendored, and Docs files are now properly excluded by default.

## Removed in X.Y.Z

- The `--include-docs` option was removed.

## [0.1.0] - 2023-05-24

### Added in 0.1.0

- Added a `--todo-types` flag which allows users to specify the TODO tags to
  search for.

### Changed in 0.1.0

- TODOs matched in multi-line comments no longer print the entire comment. Only
  the line containing the TODO is printed. Line numbers printed also correspond
  to the line where the TODO occurs rather than the starting line of the
  comment.
- Filenames and line numbers are now colored in the terminal if it supports it.
- Hidden files are now supported properly on Windows.

### Fixed in 0.1.0

- TODOs are no longer matched when starting in the middle of a comment line.

## [0.0.1] - 2023-05-15

### Added in 0.0.1

- Initial release of `todos` CLI application.
- Simple support for scanning directories for TODO/FIXME/BUG/HACK/XXX comments.

[unreleased]: https://github.com/ianlewis/todos/compare/v0.1.0...HEAD
[0.0.1]: https://github.com/ianlewis/todos/releases/tag/v0.0.1
[0.1.0]: https://github.com/ianlewis/todos/releases/tag/v0.1.0

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Added

- Support for Erlang, Haskell, R, and SQL programming languages has been added.
- A new `exclude-dir` flag was added to `todos` that allows for excluding
  matching directories from the search.
- TODO comments are matched more loosely with more delimeters such as '/' or '-'
  in addition to ':' being recognized.
- A new `--charset` flag was added which defaults to `UTF-8`. This is the
  character set used to read the code files. A special name of `detect`
  specifies that the character set of each file should be detected.
- Support was added for TODO comments with no space between the comment start
  and "TODO" marker and no delimiter.

  ```go
  //TODO Add some code here.
  ```

### Changed

- `todos` no longer detects character encodings by default and now defaults to
  reading files as UTF-8. Character detection can be enabled by using the
  `--charset=detect` flag.

## [0.4.0] - 2023-08-23

### Added in 0.4.0

- The `todos` CLI now supports a new JSON output format.
- A new [`issue-reopener`](actions/issue-reopener/README.md) GitHub action was
  added that uses the `todos` CLI to scan a repository checkout for TODOs
  referencing a GitHub issue and reopen issues that have been prematurely
  closed.

### Removed in 0.4.0

- The `github-issue-reopener` binary was removed in favor of the
  `issue-reopener` action.

## [0.3.0] - 2023-08-19

### Added in 0.3.0

- Support for [Rust](https://www.rust-lang.org/) was added.
- Support for Unix assembly language was added.
- Support for [Lua](https://www.lua.org/) was added.
- A new `github-issue-reopener` binary was added. This tool will scan a
  directory for TODOs referencing a GitHub issue and reopen issues that have
  been prematurely closed.

### Changed in 0.3.0

- Lowercase "Todo", "todo", "Fixme", "fixme", "Hack", "hack" were added as
  default TODO types.

## [0.2.0] - 2023-06-30

### Changed in 0.2.0

- Leading whitespaces is now trimmed from TODOS in multi-line comments.
- The `--include-hidden` option was replaced with the `--exclude-hidden`
  option and including hidden files was made the default.
- An `--include-vcs` option was added and the VCS directories `.git`, `.hg`,
  and `.svn` are skipped by default.

### Fixed in 0.2.0

- Hidden, Vendored, and Docs files are now properly excluded by default.

### Removed in 0.2.0

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
[0.2.0]: https://github.com/ianlewis/todos/releases/tag/v0.2.0
[0.3.0]: https://github.com/ianlewis/todos/releases/tag/v0.3.0
[0.4.0]: https://github.com/ianlewis/todos/releases/tag/v0.4.0

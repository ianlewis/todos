# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Added

- Support was added for
  [MATLAB](https://www.mathworks.com/products/matlab.html) and
  [Vim Script](https://vimdoc.sourceforge.net/htmldoc/usr_41.html).

## [0.9.0] - 2024-08-08

### Added in 0.9.0

- Support for [Fortran](https://fortran-lang.org/) was added.
- Generated files are now ignored by default. The option `--include-generated`
  was added to allow generated files to be scanned for TODOs.
- A new `--blame` option (BETA) was added which tells todos to try and get the
  VCS committer of each TODO.

### Fixed in 0.9.0

- Fixed a bug where TODOs were not being reported if they were located after a
  multi-line comment with no TODOs in the same file
  ([#1520](https://github.com/ianlewis/todos/issues/1520))

## [0.8.0] - 2024-02-21

### Added in 0.8.0

- Support for [Kotlin](https://kotlinlang.org/) was added.

### Fixed in 0.8.0

- The `--exclude-dir` option now ignores trailing path separators ([#1463](https://github.com/ianlewis/todos/issues/1463))
- The `--output` flag defaults to `github` when `todos` is run on GitHub
  Actions ([#1459](https://github.com/ianlewis/todos/issues/1459)).

### Removed in 0.8.0

- The `action/issue-reopener` GitHub Action was removed in favor of the
  [`ianlewis/todo-issue-reopener`](https://github.com/ianlewis/todo-issue-reopener).

## [0.7.0] - 2023-12-01

### Added in 0.7.0

- Support for [Emacs Lisp](https://www.gnu.org/software/emacs/), [Puppet
  manifests](https://www.puppet.com/docs/puppet/8/puppet_language), and [Visual
  Basic](https://learn.microsoft.com/en-us/dotnet/visual-basic/) was added.
- Support for recognizing multi-line comments only at the beginning of a line
  (Ruby, Perl) was added.
- Added support for [vanity urls](actions/issue-reopener/README.md#vanityurls)
  to the issue-reopener action.

## [0.6.0] - 2023-09-23

### Added in 0.6.0

- Support for [Clojure](https://clojure.org/),
  [CoffeeScript](https://coffeescript.org/),
  [Groovy](https://groovy-lang.org/), and [TeX](https://tug.org/) were added.
- Support for an `@` prefix on TODOs was added.

### Fixed in 0.6.0

- All TODOs in a multi-line comment are now reported
  ([#721](https://github.com/ianlewis/todos/pull/721)).

### Changed in 0.6.0

- The language of each file is now determined by it's file name in most
  circumstances allowing for much faster language detection.

## [0.5.0] - 2023-09-04

### Added in 0.5.0

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

- Support was added for TODO comments in JavaDoc/JSDoc/TSDoc that had a `*`
  prefix.

  ```java
  /**
   * Some comment.
   * TODO: Add some code here.
   */
  ```

### Changed in 0.5.0

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

[Unreleased]: https://github.com/ianlewis/todos/compare/v0.9.0-rc.3...HEAD
[0.0.1]: https://github.com/ianlewis/todos/releases/tag/v0.0.1
[0.1.0]: https://github.com/ianlewis/todos/releases/tag/v0.1.0
[0.2.0]: https://github.com/ianlewis/todos/releases/tag/v0.2.0
[0.3.0]: https://github.com/ianlewis/todos/releases/tag/v0.3.0
[0.4.0]: https://github.com/ianlewis/todos/releases/tag/v0.4.0
[0.5.0]: https://github.com/ianlewis/todos/releases/tag/v0.5.0
[0.6.0]: https://github.com/ianlewis/todos/releases/tag/v0.6.0
[0.7.0]: https://github.com/ianlewis/todos/releases/tag/v0.7.0
[0.8.0]: https://github.com/ianlewis/todos/releases/tag/v0.8.0
[0.9.0-rc.1]: https://github.com/ianlewis/todos/releases/tag/v0.9.0-rc.1
[0.9.0-rc.2]: https://github.com/ianlewis/todos/releases/tag/v0.9.0-rc.2
[0.9.0-rc.3]: https://github.com/ianlewis/todos/releases/tag/v0.9.0-rc.3
[0.9.0]: https://github.com/ianlewis/todos/releases/tag/v0.9.0

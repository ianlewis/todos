# Changelog

All notable changes will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### Fixed

- A new option `--follow` was added to specify whether `todos` should follow
  symlinks when traversing directories.
  ([#1752](https://github.com/ianlewis/todos/issues/1752)).

## [`0.13.0`] - 2025-05-18

### Added in `0.13.0`

- Support was added for the [Nix language](https://nix.dev/)
  ([#1653](https://github.com/ianlewis/todos/issues/1653)).
- Support was added for the
  [`CODEOWNERS`](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners)
  file and [Ignore List](https://git-scm.com/docs/gitignore) files (e.g.
  `.gitignore`) ([#1667](https://github.com/ianlewis/todos/issues/1667),
  [#1700](https://github.com/ianlewis/todos/issues/1700))
- Support was added for
  [`requirements.txt`](https://pip.pypa.io/en/stable/reference/requirements-file-format/)
  files ([#1729](https://github.com/ianlewis/todos/issues/1729)).
- `todos` has been added to the [Aqua](https://aquaproj.github.io/) registry.
  This allows `todos` to be installed using the Aqua package manager.

    ```shell
    aqua generate -g ianlewis/todos
    ```

- `todos` can now be installed via `npm`

    ```shell
    npm install -g @ianlewis/todos
    ```

- `todos` can now be run from a Docker container

    ```shell
    docker run --rm -t -v $(pwd):/src ghcr.io/ianlewis/todos /src
    ```

### Changed in `0.13.0`

- **BREAKING CHANGE:** `todos` now returns an exit status of 1 when TODOs are
  found ([#1657](https://github.com/ianlewis/todos/issues/1657)).
- **BREAKING CHANGE:** `todos` now returns an exit status of 3 when an file of
  an unsupported languages is passed via the command line
  ([#1668](https://github.com/ianlewis/todos/issues/1668)).

## [`0.12.0`] - 2025-02-28

### Changed in `0.12.0`

- `todos` now ignores files matching patterns found in `.gitignore` and
  `.todosignore` files. ([#125](https://github.com/ianlewis/todos/issues/125)).

## [`0.11.0`] - 2025-02-12

### Fixed in `0.11.0`

- Simple TODO block comments (e.g. `/* TODO */`) are now shown
  ([#1631](https://github.com/ianlewis/todos/issues/1631)).

### Added in `0.11.0`

- Support was added for [HCL](https://github.com/hashicorp/hcl)
  ([#1586](https://github.com/ianlewis/todos/issues/1586)).
- Support was added for [XSLT](https://www.w3.org/TR/xslt-30/)
  ([#1613](https://github.com/ianlewis/todos/issues/1613)).
- Support was added for [Markdown](https://en.wikipedia.org/wiki/Markdown)
  ([#1608](https://github.com/ianlewis/todos/issues/1608)).
- Support was added for [GraphQL](https://graphql.org/)
  ([#1609](https://github.com/ianlewis/todos/issues/1609)).
- Support was added for [Dart](https://dart.dev/)
  ([#1612](https://github.com/ianlewis/todos/issues/1612)).
- Support was added for [OCaml](https://ocaml.org/)
  ([#1610](https://github.com/ianlewis/todos/issues/1610)).
- Support was added for [Julia](https://julialang.org/)
  ([#1611](https://github.com/ianlewis/todos/issues/1611)).

## [`0.10.0`] - 2024-10-31

### Added in `0.10.0`

- A new `--label` flag was added to support filtering TODOs by label
  ([#1562](https://github.com/ianlewis/todos/issues/1562)).
- Support was added for
  [MATLAB](https://www.mathworks.com/products/matlab.html)
- Support was added for
  [Vim Script](https://vimdoc.sourceforge.net/htmldoc/usr_41.html).
- Support was added for
  [PowerShell](https://learn.microsoft.com/en-us/powershell/).
- Support was added for [Elixir](https://elixir-lang.org/).
- Support was added for [ERB templates](https://github.com/ruby/erb).
- Support was added for [Pascal](<https://en.wikipedia.org/wiki/Pascal_(programming_language)>).

## [`0.9.0`] - 2024-08-08

### Added in `0.9.0`

- Support for [Fortran](https://fortran-lang.org/) was added.
- Generated files are now ignored by default. The option `--include-generated`
  was added to allow generated files to be scanned for TODOs.
- A new `--blame` option (BETA) was added which tells `todos` to try and get the
  VCS committer of each TODO.

### Fixed in `0.9.0`

- Fixed a bug where TODOs were not being reported if they were located after a
  multi-line comment with no TODOs in the same file
  ([#1520](https://github.com/ianlewis/todos/issues/1520)).

## [`0.8.0`] - 2024-02-21

### Added in `0.8.0`

- Support for [Kotlin](https://kotlinlang.org/) was added.

### Fixed in `0.8.0`

- The `--exclude-dir` option now ignores trailing path separators ([#1463](https://github.com/ianlewis/todos/issues/1463))
- The `--output` flag defaults to `github` when `todos` is run on GitHub
  Actions ([#1459](https://github.com/ianlewis/todos/issues/1459)).

### Removed in `0.8.0`

- The `action/issue-reopener` GitHub Action was removed in favor of the
  [`ianlewis/todo-issue-reopener`](https://github.com/ianlewis/todo-issue-reopener).

## [`0.7.0`] - 2023-12-01

### Added in `0.7.0`

- Support for [Emacs Lisp](https://www.gnu.org/software/emacs/), [Puppet
  manifests](https://www.puppet.com/docs/puppet/8/puppet_language), and [Visual
  Basic](https://learn.microsoft.com/en-us/dotnet/visual-basic/) was added.
- Support for recognizing multi-line comments only at the beginning of a line
  (Ruby, Perl) was added.
- Added support for [vanity URLs](actions/issue-reopener/README.md#vanityurls)
  to the `todo-issue-reopener` action.

## [`0.6.0`] - 2023-09-23

### Added in `0.6.0`

- Support for [Clojure](https://clojure.org/),
  [CoffeeScript](https://coffeescript.org/),
  [Groovy](https://groovy-lang.org/), and [TeX](https://tug.org/) were added.
- Support for an `@` prefix on TODOs was added.

### Fixed in `0.6.0`

- All TODOs in a multi-line comment are now reported
  ([#721](https://github.com/ianlewis/todos/pull/721)).

### Changed in `0.6.0`

- The language of each file is now determined by it's file name in most
  circumstances allowing for much faster language detection.

## [`0.5.0`] - 2023-09-04

### Added in `0.5.0`

- Support for Erlang, Haskell, R, and SQL programming languages has been added.
- A new `exclude-dir` flag was added to `todos` that allows for excluding
  matching directories from the search.
- TODO comments are matched more loosely with more delimiters such as '/' or '-'
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

### Changed in `0.5.0`

- `todos` no longer detects character encoding by default and now defaults to
  reading files as UTF-8. Character detection can be enabled by using the
  `--charset=detect` flag.

## [`0.4.0`] - 2023-08-23

### Added in `0.4.0`

- The `todos` CLI now supports a new JSON output format.
- A new [`issue-reopener`](actions/issue-reopener/README.md) GitHub action was
  added that uses the `todos` CLI to scan a repository checkout for TODOs
  referencing a GitHub issue and reopen issues that have been prematurely
  closed.

### Removed in `0.4.0`

- The `github-issue-reopener` binary was removed in favor of the
  `issue-reopener` action.

## [`0.3.0`] - 2023-08-19

### Added in `0.3.0`

- Support for [Rust](https://www.rust-lang.org/) was added.
- Support for Unix assembly language was added.
- Support for [Lua](https://www.lua.org/) was added.
- A new `github-issue-reopener` binary was added. This tool will scan a
  directory for TODOs referencing a GitHub issue and reopen issues that have
  been prematurely closed.

### Changed in `0.3.0`

- Lowercase `Todo`, `todo`, `Fixme`, `fixme`, `Hack`, `hack` were added as
  default TODO types.

## [`0.2.0`] - 2023-06-30

### Changed in `0.2.0`

- Leading whitespace is now trimmed from TODOs in multi-line comments.
- The `--include-hidden` option was replaced with the `--exclude-hidden`
  option and including hidden files was made the default.
- An `--include-vcs` option was added and the VCS directories `.git`, `.hg`,
  and `.svn` are skipped by default.

### Fixed in `0.2.0`

- Hidden, Vendored, and Docs files are now properly excluded by default.

### Removed in `0.2.0`

- The `--include-docs` option was removed.

## [`0.1.0`] - 2023-05-24

### Added in `0.1.0`

- Added a `--todo-types` flag which allows users to specify the TODO tags to
  search for.

### Changed in `0.1.0`

- TODOs matched in multi-line comments no longer print the entire comment. Only
  the line containing the TODO is printed. Line numbers printed also correspond
  to the line where the TODO occurs rather than the starting line of the
  comment.
- Filenames and line numbers are now colored in the terminal if it supports it.
- Hidden files are now supported properly on Windows.

### Fixed in `0.1.0`

- TODOs are no longer matched when starting in the middle of a comment line.

## [`0.0.1`] - 2023-05-15

### Added in `0.0.1`

- Initial release of `todos` CLI application.
- Simple support for scanning directories for `TODO`/`FIXME`/`BUG`/`HACK`/`XXX`
  comments.

[`0.0.1`]: https://github.com/ianlewis/todos/releases/tag/v0.0.1
[`0.1.0`]: https://github.com/ianlewis/todos/releases/tag/v0.1.0
[`0.2.0`]: https://github.com/ianlewis/todos/releases/tag/v0.2.0
[`0.3.0`]: https://github.com/ianlewis/todos/releases/tag/v0.3.0
[`0.4.0`]: https://github.com/ianlewis/todos/releases/tag/v0.4.0
[`0.5.0`]: https://github.com/ianlewis/todos/releases/tag/v0.5.0
[`0.6.0`]: https://github.com/ianlewis/todos/releases/tag/v0.6.0
[`0.7.0`]: https://github.com/ianlewis/todos/releases/tag/v0.7.0
[`0.8.0`]: https://github.com/ianlewis/todos/releases/tag/v0.8.0
[`0.9.0`]: https://github.com/ianlewis/todos/releases/tag/v0.9.0
[`0.10.0`]: https://github.com/ianlewis/todos/releases/tag/v0.10.0
[`0.11.0`]: https://github.com/ianlewis/todos/releases/tag/v0.11.0
[`0.12.0`]: https://github.com/ianlewis/todos/releases/tag/v0.12.0
[`0.13.0`]: https://github.com/ianlewis/todos/releases/tag/v0.13.0

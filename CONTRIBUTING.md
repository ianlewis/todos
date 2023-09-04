# How to contribute

We'd love to accept your patches and contributions to this project.

## How can I help?

There are many areas in the project that need help. These are managed in GitHub
issues. Please let us know if you are willing to work on the issue and how you
can contribute.

- For new developers and contributors, please see issues labeled
  [good first issue]. These issues should require minimal background knowledge
  to contribute.
- For slightly more involved changes that may require some background knowledge,
  please see issues labeled [help wanted]
- For experienced developers, any of the [open issues] is open to contribution.

If you don't find an existing issue for your contribution feel free to
[create an issue].

## Before you begin

### Sign our Contributor License Agreement

Contributions to this project must be accompanied by a [Contributor License
Agreement] (CLA). You (or your employer) retain the copyright to your
contribution; this simply gives us permission to use and redistribute your
contributions as part of the project.

If you or your current employer have already signed the Google CLA (even if it
was for a different project), you probably don't need to do it again.

Visit <https://cla.developers.google.com/> to see your current agreements or to
sign a new one.

### Review our community guidelines and Code of Conduct

This project follows [Google's Open Source Community Guidelines] and has a [Code
of Conduct] which all contributors are expected to follow. Please be familiar
with them. Please see the Code of Conduct about how to report instances of
abusive, harassing, or otherwise unacceptable behavior.

## Providing feedback

Feedback can include bug reports, feature requests, documentation change
proposals, or just general feedback. The best way to provide feedback to the
project is to [create an issue]. Please provide as much info as you can about
your project feedback.

For reporting a security vulnerability see the [Security Policy].

## Code contribution process

This section describes how to make a contribution to this project.

### Create a fork

You should start by [creating a fork](https://github.com/ianlewis/todos/fork) of
the repository under your own account so that you can push your commits.

### Clone the Repository

Make sure you have configured `git` to connect to GitHub using `SSH`. See
[Connecting to GitHub with SSH] for more info.

One you have done that you can clone the repository to get a checkout of the
code. Substitute your username here.

```shell
git clone git@github.com:myuser/todos.git
```

### Create a local branch

Create a local branch to do development in. This will make it easier to create a
pull request later. You can give this branch an appropriate name.

```shell
git checkout -b my-new-feature
```

### Development

Next you can develop your new feature or bug-fix. Please see the following
sections on how to use the various tools used by this project during
development.

#### The Makefile

The local repository makes heavy use of `make`. Type `make` to see a full list
of `Makefile` targets.

```shell
$ make
todos Makefile
Usage: make [COMMAND]

  help                 Shows all targets and help from the Makefile (this message).
Testing
  unit-test            Runs all unit tests.
  go-test              Runs Go unit tests.
  ts-test              Run TypeScript unit tests.
Benchmarking
  go-benchmark         Runs Go benchmarks.
Tools
  autogen              Runs autogen on code files.
Linters
  lint                 Run all linters.
  actionlint           Runs the actionlint linter.
  eslint               Runs the eslint linter.
  markdownlint         Runs the markdownlint linter.
  golangci-lint        Runs the golangci-lint linter.
  yamllint             Runs the yamllint linter.
Maintenance
  clean                Delete temporary files.
```

#### Linters

The project uses the [`actionlint`](https://github.com/rhysd/actionlint),
[`eslint`](https://eslint.org/),
[`golangci-lint`](https://github.com/golangci/golangci-lint),
[`markdownlint`](https://github.com/DavidAnson/markdownlint), and
[`yamllint`](https://www.yamllint.com/) for linting various files. You do not
necessarily need to have all installed but you will need to install those that
you want to run them locally.

You can run all linters with the `lint` make target:

```shell
make lint
```

or individually by name:

```shell
make golangci-lint
```

#### Running tests

You can run all unit tests using the `unit-test` make target:

```shell
make unit-test
```

#### Running tests for Go code

You can lint Go code and run tests with the `golangci-lint` and `go-test`
targets.

```shell
make golangci-lint go-test
```

#### Running tests for TypeScript code

You can lint TypeScript code and run tests with the `eslint` and `ts-test`
targets.

```shell
make eslint ts-test
```

#### Commit and push your code

Make sure to stage any change or new files or new files.

```shell
git add .
```

Commit your code to your branch. Commit messages should follow the [Conventional
Commits] format but this isn't required.

```shell
git commit -sm "feat: My new feature"
```

You can now push your changes to your fork.

```shell
git push origin my-new-feature
```

### Pull requests

Once you have your code pushed to your fork you can now created a new
[pull request] (PR). This allows the project maintainers to review your submission.

#### Create a PR

You can
[create a new pull request via the GitHub UI](https://github.com/ianlewis/todos/compare)
or via the [`gh` CLI] tool. Create the PR as a
[draft](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests#draft-pull-requests)
to start.

```shell
gh pr create --title "feat: My new feature" --draft
```

#### Review the PR checklist

The newly created PR will have a checklist in the description pulled from the PR
template. Please review this checklist and mark each item as complete.

Once you have finished you can mark the PR as "Ready for review".

#### Code reviews

All PR submissions require review. We use GitHub pull requests for this purpose.
Consult [About pull request reviews] for more information on pull request code
reviews. Once you have created a PR you should get a response within a few days.

#### Merge

After your code is approved it will be merged into the `main` branch! Congrats!

## Conventions

This section contains info on general conventions in the project.

### Code style and formatting

Most code, scripts, and documentation should be auto-formatted using a
formatting tool.

1. Go code should be is formatted using [`gofumpt`].
2. TypeScript code should be [`prettier`].
3. YAML should be formatted using [`prettier`].
4. Markdown should be formatted using [`prettier`].

### Semantic Versioning

This project uses [Semantic Versioning] for release versions.

This means that when creating a new release version, in general, given a version
number MAJOR.MINOR.PATCH, increment the:

1. MAJOR version when you make incompatible API changes
2. MINOR version when you add functionality in a backward compatible manner
3. PATCH version when you make backward compatible bug fixes

### Conventional Commits

PR titles are required to be in [Conventional Commits] format. Supported
prefixes are defined in the file
[`.github/pr-title-checker-config.json`](./.github/pr-title-checker-config.json).

The following prefixes are supported:

1. `fix`: patches a bug
2. `feat`: introduces a new feature
3. `docs`: a change in the documentation.
4. `chore`: a change that performs a task but doesn't change functionality, such as updating dependencies.
5. `refactor`: a code change that improves code quality
6. `style`: coding style or format changes
7. `build`: changes that affect the build system
8. `ci`: changes to CI/CD configuration files or scripts
9. `perf`: change to improve performance
10. `revert`: reverts a previous change
11. `test`: adds missing tests or corrects existing tests

[create an issue]: https://github.com/ianlewis/todos/issues/new/choose
[good first issue]: https://github.com/ianlewis/todos/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22
[help wanted]: https://github.com/ianlewis/todos/issues?q=is%3Aissue+is%3Aopen+label%3A%22status%3Ahelp+wanted%22
[open issues]: https://github.com/ianlewis/todos/issues
[Security Policy]: SECURITY.md
[Code of Conduct]: CODE_OF_CONDUCT.md
[Contributor License Agreement]: https://cla.developers.google.com/about
[Google's Open Source Community Guidelines]: https://opensource.google/conduct/
[Connecting to GitHub with SSH]: https://docs.github.com/en/authentication/connecting-to-github-with-ssh
[pull request]: https://docs.github.com/pull-requests
[`gh` CLI]: https://cli.github.com/
[About pull request reviews]: https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/about-pull-request-reviews
[Semantic Versioning]: https://semver.org/
[Conventional Commits]: https://www.conventionalcommits.org/en/v1.0.0/
[`gofumpt`]: https://github.com/mvdan/gofumpt
[`prettier`]: https://prettier.io/

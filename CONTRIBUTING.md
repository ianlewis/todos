# Contributor Guide

This doc describes how to contribute to this repository.

First, We'd love to accept your patches and contributions to my projects!

## How can I help?

There are many areas repositories that need help. These are managed in GitHub
issues. Please let us know if you are willing to work on the issue and how you
can contribute.

- For new developers and contributors, please see issues labeled
  `good first issue`. These issues should require minimal background knowledge
  to contribute.
- For slightly more involved changes that may require some background knowledge,
  please see issues labeled `help wanted`
- For experienced developers, any of the open issues are open to contribution.

If you don't find an existing issue for your contribution feel free to
create an issue.

## Before you begin

### Review the community guidelines and Code of Conduct

All of my repositories follow [Google's Open Source Community Guidelines] and
contributors are also expected to follow my [Code of Conduct]. Please be
familiar with them. Please see the Code of Conduct about how to report instances
of abusive, harassing, or otherwise unacceptable behavior.

## Providing feedback

Feedback can include bug reports, feature requests, documentation change
proposals, or just general feedback. The best way to provide feedback to the
project is to create an issue. Please provide as much info as you can about
your project feedback.

For reporting a security vulnerability see the [Security Policy].

## Code contribution process

This section describes how to make a contribution to my repositories.

### Create a fork

You should start by
[creating a fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo)
of the repository under your own account so that you can push your commits.

### Clone the Repository

Make sure you have configured `git` to connect to GitHub using `SSH`. See
[Connecting to GitHub with SSH] for more info.

One you have done that you can clone the repository to get a checkout of the
code. Substitute your username and repository name here.

```shell
git clone git@github.com:myuser/myrepo.git
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

Most of my repositories make heavy use of `make` during development. This helps
with automation of tasks locally on your machine. Type `make` to see a full list
of `Makefile` targets.

Here is an example from the
[`repo-template`](https://github.com/ianlewis/repo-template) repository.

```shell
$ make
repo-template Makefile
Usage: make [COMMAND]

  help                 Shows all targets and help from the Makefile (this message).
Tools
  license-headers      Update license headers.
Formatting
  format               Format all files
  md-format            Format Markdown files.
  yaml-format          Format YAML files.
Linting
  lint                 Run all linters.
  actionlint           Runs the actionlint linter.
  zizmor               Runs the zizmor linter.
  markdownlint         Runs the markdownlint linter.
  yamllint             Runs the yamllint linter.
Maintenance
  clean                Delete temporary files.
```

#### Linters

Most projects use the following linters depending on the programming languages
used.

- [`actionlint`](https://github.com/rhysd/actionlint): For GitHub actions
  workflows.
- [`eslint`](https://eslint.org/): For JavaScript and TypeScript.
- [`golangci-lint`](https://github.com/golangci/golangci-lint): For Go.
- [`markdownlint`](https://github.com/DavidAnson/markdownlint): For markdown.
- [`yamllint`](https://www.yamllint.com/): For YAML (GitHub Actions workflows,
  configuration files etc.)

You do not necessarily need to have all installed but you will need to install
those that you want to run them locally.

You can run all linters with the `lint` make target:

```shell
make lint
```

or individually by name:

```shell
make markdownlint
```

#### Running tests

Where unit tests exist, you can run all unit tests using the `unit-test` make
target:

```shell
make unit-test
```

You can run unit tests for individual languages with the appropriate `make`
target. These may vary a bit depending on the repository and code layout.
Typing `make` or `make help` will show the full list of targets.

For example, this runs Go unit tests.

```shell
make go-test
```

#### Commit and push your code

Make sure to stage any change or new files or new files.

```shell
git add .
```

Commit your code to your branch. For most repositories, messages should follow
the [Conventional Commits] format but this isn't always required.

Commits should include a [Developer Certificate of Origin] (DCO). This can be
included automatically in commits using the `-s` flag.

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
[create a new pull request via the GitHub UI](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request?tool=webui)
or [via the `gh` CLI tool](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request?tool=cli). Create the PR as a
[draft](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-pull-requests#draft-pull-requests)
to start.

```shell
gh pr create --title "feat: My new feature" --draft
```

#### Review the PR template/checklist

Some repositories will have a PR template with instructions on how to document
your PR. Some will include a checklist. Please review the template doc and mark
checklist items as complete before finalizing your PR.

Once you have finished you can mark the PR as "Ready for review".

#### Pre-submits

PRs perform number of [GitHub status checks] which run linters and tests. These
tests must all pass before a PR will be accepted. These tests are located in
the [`.github/workflows`](.github/workflows) directory and begin with the
prefix `pre-sumbit`.

Unit tests are run as pre-submit tests in the
[`pre-submit.units.yml`](.github/workflows/pre-submit.units.yml) file. To run
these tests locally see the instructions in the [`Running
Tests`](#running-tests) section.

#### Code reviews

All PR submissions require review. I use GitHub pull requests for this purpose.
Consult [About pull request reviews] for more information on pull request code
reviews. Once you have created a PR you should get a response within a few days.

#### Merge

After your code is approved it will be merged into the `main` branch! Congrats!

## Conventions

This section contains info on general conventions I use in my repositories.

### Code style and formatting

Most code, scripts, and documentation should be auto-formatted using a
formatting tool.

1. Go code should be is formatted using [`gofumpt`].
2. TypeScript code should be [`prettier`].
3. Python code should be formatted with [`black`].
4. Rust code should be formatted with [`rustfmt`].
5. YAML should be formatted using [`prettier`].
6. Markdown should be formatted using [`prettier`].

### Semantic Versioning

My repositories use [Semantic Versioning] for release versions.

This means that when creating a new release version, in general, given a version
number MAJOR.MINOR.PATCH, increment the:

1. MAJOR version when you make incompatible API changes
2. MINOR version when you add functionality in a backward compatible manner
3. PATCH version when you make backward compatible bug fixes

### Conventional Commits

PR titles should be in [Conventional Commits] format. Usually this is required
by not always.

In general, the following prefixes are supported:

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

[Security Policy]: SECURITY.md
[Code of Conduct]: CODE_OF_CONDUCT.md
[Developer Certificate of Origin]: https://en.wikipedia.org/wiki/Developer_Certificate_of_Origin
[Google's Open Source Community Guidelines]: https://opensource.google/conduct/
[Connecting to GitHub with SSH]: https://docs.github.com/en/authentication/connecting-to-github-with-ssh
[pull request]: https://docs.github.com/pull-requests
[About pull request reviews]: https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/about-pull-request-reviews
[Semantic Versioning]: https://semver.org/
[Conventional Commits]: https://www.conventionalcommits.org/en/v1.0.0/
[`gofumpt`]: https://github.com/mvdan/gofumpt
[`prettier`]: https://prettier.io/
[`black`]: https://github.com/psf/black
[`rustfmt`]: https://github.com/rust-lang/rustfmt
[GitHub status checks]: https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/collaborating-on-repositories-with-code-quality-features/about-status-checks

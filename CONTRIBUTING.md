# How to contribute

We'd love to accept your patches and contributions to this project.

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

### Review our community guidelines and Code of Conduct.

This project follows [Google's Open Source Community Guidelines] and has a [Code
of Conduct] which all contributors are expected to follow. Please be familiar
with them. Please see the Code of Conduct about how to report instances of
abusive, harassing, or otherwise unacceptable behavior.

## Contribution process

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

You can [create a new pull request via the GitHub UI] or via the
[`gh` CLI] tool. Create the PR as a draft to start.

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

#### Merge!

After your code is approved it will be merged into the `main` branch! Congrats!

[Code of Conduct]: CODE_OF_CONDUCT.md
[Contributor License Agreement]: https://cla.developers.google.com/about
[Google's Open Source Community Guidelines]: https://opensource.google/conduct/
[Connecting to GitHub with SSH]: https://docs.github.com/en/authentication/connecting-to-github-with-ssh
[pull request]: https://docs.github.com/pull-requests
[`gh` CLI]: https://cli.github.com/
[About pull request reviews]: https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests/about-pull-request-reviews

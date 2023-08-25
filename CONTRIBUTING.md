# How to contribute

We'd love to accept your patches and contributions to this project.

## Before you begin

### Sign our Contributor License Agreement

Contributions to this project must be accompanied by a
[Contributor License Agreement](https://cla.developers.google.com/about) (CLA).
You (or your employer) retain the copyright to your contribution; this simply
gives us permission to use and redistribute your contributions as part of the
project.

If you or your current employer have already signed the Google CLA (even if it
was for a different project), you probably don't need to do it again.

Visit <https://cla.developers.google.com/> to see your current agreements or to
sign a new one.

### Review our community guidelines

This project follows
[Google's Open Source Community Guidelines](https://opensource.google/conduct/).

## Contribution process

### Code reviews

All submissions, including submissions by project members, require review. We
use GitHub pull requests for this purpose. Consult
[GitHub Help](https://help.github.com/articles/about-pull-requests/) for more
information on using pull requests.

## Development

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

### Linters

The project uses the [`actionlint`](https://github.com/rhysd/actionlint),
[`eslint`](https://eslint.org/),
[`golangci-lint`](https://github.com/golangci/golangci-lint),
[`markdownlint`](https://github.com/DavidAnson/markdownlint), and
[`yamllint`](https://www.yamllint.com/) for linting various files. You do not
necessarily need to have all installed but you will need to install each to run
them locally.

You can run all linters with the `lint` make target:

```shell
make lint
```

### Running tests

You can run all unit tests using the `unit-test` make target:

```shell
make unit-test
```

### Running tests for Go code

You can lint Go code and run tests with the `golangci-lint` and `go-test`
targets.

```shell
make golangci-lint go-test
```

### Running tests for TypeScript code

You can lint TypeScript code and run tests with the `eslint` and `ts-test`
targets.

```shell
make eslint ts-test
```

# todos

[![unit tests](https://github.com/ianlewis/todos/actions/workflows/pre-submit.units.yml/badge.svg)](https://github.com/ianlewis/todos/actions/workflows/pre-submit.units.yml)
[![codecov](https://codecov.io/gh/ianlewis/todos/branch/main/graph/badge.svg?token=0EBN8DQYFR)](https://codecov.io/gh/ianlewis/todos)
[![Go Report Card](https://goreportcard.com/badge/github.com/ianlewis/todos)](https://goreportcard.com/report/github.com/ianlewis/todos)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fianlewis%2Ftodos.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fianlewis%2Ftodos?ref=badge_shield)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/ianlewis/todos/badge)](https://api.securityscorecards.dev/projects/github.com/ianlewis/todos)
[![SLSA 3](https://slsa.dev/images/gh-badge-level3.svg)](https://slsa.dev)

Tools for dealing with TODOs in code.

- [`todos` CLI](#todos-cli): searches for TODO comments in code and prints them in various
  formats.
- [`actions/issue-reopener`](actions/issue-reopener/README.md): A GitHub Action to reopen issues that still have
  TODO comments referencing them.

## TODO comments

"TODO" comments are comments in code that mark a task that is intended to be
done in the future.

For example:

```go
// TODO: Update this code.
```

### TODO comment variants

There a few veriants of this type of comment thare are in wide use.

- **TODO**: A general TODO comment indicating something that is to be done in
  the future.
- **FIXME**: Something that is broken that needs to be fixed in the code.
- **BUG**: A bug in the code that needs to be fixed.
- **HACK**: This code is a "hack"; a hard to understand or brittle piece of
  code. It could use a cleanup.
- **XXX**: Danger! Similar to "HACK". Modifying this code is dangerous. It
- **COMBAK**: Something you should "come back" to.

### TODO comment formats

There are a few ways to format a TODO comment with metadata.

- A naked TODO comment.

  ```go
  // TODO
  ```

- A TODO comment with an explanation message

  ```go
  // TODO: Do something.
  ```

- A TODO comment with a linked bug or issue and optional message

  ```go
  // TODO(github.com/ianlewis/todos/issues/8): Do something.
  ```

- A TODO comment with a username and optional message. This type is discouraged
  as it links the issue to a specific developer but can be helpful temporarily
  when making changes to a PR. Linking to issues is recommended for permanent
  comments.

  ```go
  // TODO(ianlewis): Do something.
  ```

### Finding TODOs in your code

You can use the [`todos` CLI] to find TODO comments in your code and print them
out. Running it will search the directory tree starting at the current
directory by default.

```shell
$ todos
main.go:27:// TODO(#123): Return a proper exit code.
main.go:28:// TODO(ianlewis): Implement the main method.
```

## todos CLI

The `todos` CLI scans files in a directory and prints any "TODO" comments it
finds in various formats.

```shell
$ todos
internal/scanner/config.go:134:// TODO: Perl supports strings with any delimiter.
internal/walker/walker.go:213:// TODO(github.com/ianlewis/linguist/issues/1): Update when linguist supports Windows.
```

### Install the todos CLI

### Install from a release

Download the `slsa-verifier` and verify it's checksum:

```shell
curl -sSLo slsa-verifier https://github.com/slsa-framework/slsa-verifier/releases/download/v2.3.0/slsa-verifier-linux-amd64 && \
echo "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d  slsa-verifier" | sha256sum -c - && \
chmod +x slsa-verifier
```

Download and verify the `todos` CLI binary and verify it's provenance:

```shell
curl -sSLo todos https://github.com/ianlewis/todos/releases/download/v0.4.0/todos-linux-amd64 && \
curl -sSLo todos.intoto.jsonl https://github.com/ianlewis/todos/releases/download/v0.4.0/todos-linux-amd64.intoto.jsonl && \
./slsa-verifier verify-artifact todos --provenance-path todos.intoto.jsonl --source-uri github.com/ianlewis/todos --source-tag v0.4.0 && \
chmod +x todos && \
cp todos /usr/local/bin
```

#### Install from source

If you already have Go 1.20+ you can install the latest version using `go install`:

```shell
go install github.com/ianlewis/todos/internal/cmd/todos
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contributor documentation.

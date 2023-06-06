# todos

Tools for dealing with TODOs in code.

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
  generaly works but it's not fully understood why or is hard to follow.

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
  as it links the issue to a specific developer. Linking to issues is
  recommended.

  ```go
  // TODO(ianlewis): Do something.
  ```

## todos CLI

The `todos` CLI scans files in a directory and prints any "TODO" comments it
finds.

```shell
$ todos
internal/scanner/config.go:134:// TODO: Perl supports strings with any delimiter.
internal/walker/walker.go:213:// TODO(github.com/ianlewis/linguist/issues/1): Update when linguist supports Windows.
```

### Install the todos CLI

#### Install from source

Install the latest version using `go install`:

```shell
go install github.com/ianlewis/todos/internal/cmd/todos
```

## Development

### Running tests

You can run unit tests using the `unit-test` make target:

```shell
make unit-test
```

# issue-reopener

The `ianlewis/todos/actions/issue-reopener` GitHub Action searches your checked
out code for TODO comments that reference issues and reopens issues that have
been closed prematurely. `issue-reopener` will also add a comment on the issue
with a link to the source code where the TODO can be found.

TODO comments can take the following forms:

```golang
// TODO(#123): Referencing the issue number with a pound sign.
// TODO(123): Referencing the issue number only.
// TODO(github.com/owner/repo/issues/123): Referencing the issue url without scheme.
// TODO(https://github.com/owner/repo/issues/123): Referencing the issue url with scheme.
```

## Getting Started

First use the `actions/checkout` action to check out your repository. After that
you can call `ianlewis/todos/actions/issue-reopener` to scan your codebase for
TODO comments.

Note that you must set the `issues: write` permission on the job if using the
default `GITHUB_TOKEN`.

```yaml
on:
  workflow_dispatch:
  schedule:
    - cron: "0 0 * * *"

permissions: {}

jobs:
  issue-reopener:
    runs-on: ubuntu-latest
    permissions:
      issues: write
    steps:
      - uses: actions/checkout@v3
      - name: Issue Reopener
        uses: ianlewis/todos/actions/issue-reopener@v0.4.0
```

## Inputs

| Name        | Required | Default            | Description                                                                |
| ----------- | -------- | ------------------ | -------------------------------------------------------------------------- |
| path        | No       | `github.workspace` | The root path of the source code to search.                                |
| token       | No       | `github.token`     | The GitHub token to use. This token must have `issues: write` permissions. |
| dry-run     | No       | `false`            | If true, issues are only output to logs and not actually reopened.         |
| config-path | No       | `.todos.yml`       | Path to the todos config file.                                             |

## Outputs

There are currently no outputs.

## Configuration

A configuration file in YAML format can added to your repository.

### vanityURLs

Some projects use a custom vanity URL for issues. `vanityURLs` is a list of
[`RegExp`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/RegExp)
values used to match URLs and extract a GitHub issue number. The issue number
must be extracted with a
[named capturing group](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Regular_expressions/Named_capturing_group)
named "id".

Example:

```yaml
vanityURLs:
  - "^\\s*(https?://)?golang.org/issues/(?<id>[0-9]+)\\s*$",
```

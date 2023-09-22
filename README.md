# todos

[![unit tests](https://github.com/ianlewis/todos/actions/workflows/pre-submit.units.yml/badge.svg)](https://github.com/ianlewis/todos/actions/workflows/pre-submit.units.yml)
[![codecov](https://codecov.io/gh/ianlewis/todos/branch/main/graph/badge.svg?token=0EBN8DQYFR)](https://codecov.io/gh/ianlewis/todos)
[![Go Report Card](https://goreportcard.com/badge/github.com/ianlewis/todos)](https://goreportcard.com/report/github.com/ianlewis/todos)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fianlewis%2Ftodos.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fianlewis%2Ftodos?ref=badge_shield)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/ianlewis/todos/badge)](https://api.securityscorecards.dev/projects/github.com/ianlewis/todos)
[![SLSA 3](https://slsa.dev/images/gh-badge-level3.svg)](https://slsa.dev)

Parse out TODOs, FIXMEs, BUGs and HACKs from your code.

<p align="center"><img src="img/todos.png?raw=true"/></p>

This repo contains the following tools for dealing with TODOs in code.

- [`todos` CLI]: searches for TODO comments in code and prints them in various
  formats.
- [`actions/issue-reopener`]: A GitHub Action to reopen issues that still have
  TODO comments referencing them.

See the [FAQ] for more info on the philosophy behind the project.

## TODO comment format

"TODO" comments are comments in code that mark a task that is intended to be
done in the future.

For example:

```go
// TODO(some label or reference): Update this code.
```

### TODO comment type variants

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

TODO comments can include some optional metadata. Here are some examples:

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

In order for the comments to be more easily parsed keep in mind the following:

- Spaces between the comment start and 'TODO' is optional (e.g. `//TODO: some comment`)
- TODOs should have a colon if a message is present so it can be distingished
  from normal comments.
- Comments can be on the same line with other code (e.g. `x = f() // TODO: call f`
- Only the single line where the TODO occurs is printed for multi-line comments.
- `TODO`,`FIXME`,`BUG`,`HACK`,`XXX`,`COMBAK` are supported by default. You can
  change this with the `--todo-types` flag.

## `todos` CLI tool

The `todos` CLI scans files in a directory and prints any "TODO" comments it
finds in various formats.

### Install `todos`

There are two methods for installing `todos`.

#### Install `todos` from a release

Download the [`slsa-verifier`](https://github.com/slsa-framework/slsa-verifier)
and verify it's checksum:

```shell
curl -sSLo slsa-verifier https://github.com/slsa-framework/slsa-verifier/releases/download/v2.3.0/slsa-verifier-linux-amd64 && \
echo "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d  slsa-verifier" | sha256sum -c - && \
chmod +x slsa-verifier
```

Download and verify the `todos` CLI binary and verify it's provenance:

```shell
curl -sSLo todos https://github.com/ianlewis/todos/releases/download/v0.5.0/todos-linux-amd64 && \
curl -sSLo todos.intoto.jsonl https://github.com/ianlewis/todos/releases/download/v0.5.0/todos-linux-amd64.intoto.jsonl && \
./slsa-verifier verify-artifact todos --provenance-path todos.intoto.jsonl --source-uri github.com/ianlewis/todos --source-tag v0.5.0 && \
chmod +x todos && \
sudo cp todos /usr/local/bin
```

#### Install `todos` from source

If you already have Go 1.20+ you can install the latest version using `go install`:

```shell
go install github.com/ianlewis/todos/internal/cmd/todos
```

### Usage

Simply running `todos` will search TODO comments starting in the current
directory. By default it ignores files that are in "VCS" directories (such as`.git`
or `.hg`) and vendored code (such as `node_modules`, `vendor`, and `third_party`).

Here is an example running in a checkout of the
[`Kubernetes`](https://github.com/kubernetes/kubernetes) codebase.

```shell
kubernetes$ todos
build/common.sh:346:# TODO: remove when 17.06.0 is not relevant anymore
build/lib/release.sh:148:# TODO: Docker images here
cluster/addons/addon-manager/kube-addons.sh:233:# TODO: Remove the first command in future release.
cluster/addons/calico-policy-controller/ipamblock-crd.yaml:41:# TODO: This nullable is manually added in. We should update controller-gen
cluster/addons/dns/kube-dns/kube-dns.yaml.base:119:# TODO: Set memory limits when we've profiled the container for large
cluster/addons/dns/kube-dns/kube-dns.yaml.in:119:# TODO: Set memory limits when we've profiled the container for large
cluster/addons/fluentd-gcp/fluentd-gcp-configmap-old.yaml:120:# TODO(random-liu): Remove this after cri container runtime rolls out.
cluster/addons/fluentd-gcp/fluentd-gcp-configmap-old.yaml:244:# TODO(random-liu): Keep this for compatibility, remove this after
cluster/addons/fluentd-gcp/fluentd-gcp-configmap-old.yaml:345:# TODO(instrumentation): Reconsider this workaround later.
cluster/addons/fluentd-gcp/fluentd-gcp-configmap.yaml:135:# TODO(random-liu): Remove this after cri container runtime rolls out.
cluster/addons/fluentd-gcp/fluentd-gcp-configmap.yaml:259:# TODO(random-liu): Keep this for compatibility, remove this after
cluster/addons/fluentd-gcp/fluentd-gcp-configmap.yaml:362:# TODO(instrumentation): Reconsider this workaround later.
...
```

#### Running on sub-directories or files

You can run `todos` on sub-directories or individual files by passing them on
the command line.

```shell
kubernetes$ todos hack/ Makefile
hack/e2e-internal/e2e-cluster-size.sh:32:#TODO(colhom): spec and implement federated version of this
hack/ginkgo-e2e.sh:118:# TODO(kubernetes/test-infra#3330): Allow NODE_INSTANCE_GROUP to be
hack/lib/golang.sh:456:# TODO: This symlink should be relative.
hack/lib/protoc.sh:119:# TODO: switch to universal binary when updating to 3.20+
hack/lib/util.sh:337:# TODO(lavalamp): Simplify this by moving pkg/api/v1 and splitting pkg/api,
hack/lib/version.sh:74:# TODO: We continue calling this "git version" because so many
hack/make-rules/test.sh:61:# TODO: This timeout should really be lower, this is a *long* time to test one
hack/module-graph.sh:19:# TODO: Containerize the script to remove dependency issues with go mod and dot.
hack/update-codegen.sh:420:# TODO: it might be better in the long term to make peer-types explicit in the
hack/verify-api-groups.sh:96:# TODO: Remove this package completely and from this list
hack/verify-e2e-test-ownership.sh:20:# TODO: these two can be dropped if KubeDescribe is gone from codebase
hack/verify-external-dependencies-version.sh:39:# TODO: revert sed hack when zetigeist respects CLICOLOR/ttys
hack/verify-licenses.sh:101:# TODO: Remove this workaround check once PR https://github.com/google/go-licenses/pull/110 is merged
Makefile:313:# TODO(thockin): Remove this in v1.29.
Makefile:504:#TODO: make EXCLUDE_TARGET auto-generated when there are other files in cmd/
```

#### Running in GitHub Actions

If run as part of a GitHub action `todos` will function much like a linter and
output GitHub workflow commands which will add check comments to PRs.

```shell
kubernetes$ todos -o github Makefile
::warning file=Makefile,line=313::# TODO(thockin): Remove this in v1.29.
::warning file=Makefile,line=504::#TODO: make EXCLUDE_TARGET auto-generated when there are other files in cmd/
```

An example workflow might look like the following. `todos` will output GitHub
Actions workflow commands by default when running on GitHub Actions:

```yaml
on:
  pull_request:
    branches: [main]
  workflow_dispatch:

permissions: {}

jobs:
  todos:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@v3
      - name: install todos
        run: |
          curl -sSLo slsa-verifier https://github.com/slsa-framework/slsa-verifier/releases/download/v2.3.0/slsa-verifier-linux-amd64 && \
          echo "ea687149d658efecda64d69da999efb84bb695a3212f29548d4897994027172d  slsa-verifier" | sha256sum -c - && \
          chmod +x slsa-verifier

          curl -sSLo todos https://github.com/ianlewis/todos/releases/download/v0.5.0/todos-linux-amd64 && \
          curl -sSLo todos.intoto.jsonl https://github.com/ianlewis/todos/releases/download/v0.5.0/todos-linux-amd64.intoto.jsonl && \
          ./slsa-verifier verify-artifact todos --provenance-path todos.intoto.jsonl --source-uri github.com/ianlewis/todos --source-tag v0.5.0 && \
          rm -f slsa-verifier && \
          chmod +x todos

      - name: run todos
        run: |
          ./todos .
```

#### Outputting JSON

`todos` can produce output in JSON format for more complicated processing.

```shell
kubernetes$ todos -o json
{"path":"build/common.sh","type":"TODO","text":"# TODO: remove when 17.06.0 is not relevant anymore","label":"","message":"remove when 17.06.0 is not relevant anymore","line":346,"comment_line":346}
{"path":"build/lib/release.sh","type":"TODO","text":"# TODO: Docker images here","label":"","message":"Docker images here","line":148,"comment_line":148}
{"path":"cluster/addons/addon-manager/kube-addons.sh","type":"TODO","text":"# TODO: Remove the first command in future release.","label":"","message":"Remove the first command in future release.","line":233,"comment_line":233}
{"path":"cluster/addons/calico-policy-controller/ipamblock-crd.yaml","type":"TODO","text":"# TODO: This nullable is manually added in. We should update controller-gen","label":"","message":"This nullable is manually added in. We should update controller-gen","line":41,"comment_line":41}
{"path":"cluster/addons/dns/kube-dns/kube-dns.yaml.base","type":"TODO","text":"# TODO: Set memory limits when we've profiled the container for large","label":"","message":"Set memory limits when we've profiled the container for large","line":119,"comment_line":119}
{"path":"cluster/addons/dns/kube-dns/kube-dns.yaml.in","type":"TODO","text":"# TODO: Set memory limits when we've profiled the container for large","label":"","message":"Set memory limits when we've profiled the container for large","line":119,"comment_line":119}
{"path":"cluster/addons/fluentd-gcp/fluentd-gcp-configmap-old.yaml","type":"TODO","text":"# TODO(random-liu): Remove this after cri container runtime rolls out.","label":"random-liu","message":"Remove this after cri container runtime rolls out.","line":120,"comment_line":120}
{"path":"cluster/addons/fluentd-gcp/fluentd-gcp-configmap-old.yaml","type":"TODO","text":"# TODO(random-liu): Keep this for compatibility, remove this after","label":"random-liu","message":"Keep this for compatibility, remove this after","line":244,"comment_line":244}
{"path":"cluster/addons/fluentd-gcp/fluentd-gcp-configmap-old.yaml","type":"TODO","text":"# TODO(instrumentation): Reconsider this workaround later.","label":"instrumentation","message":"Reconsider this workaround later.","line":345,"comment_line":345}
...
```

```shell
kubernetes$ # Get all the unique files with TODOs that Tim Hockin owns.
kubernetes$ todos -o json | jq -r '. | select(.label = "thockin") | .path' | uniq
```

## Related projects

- [pgilad/leasot](https://github.com/pgilad/leasot): A fairly robust tool with good integration with the Node.js ecosystem.
- [judepereira/checktodo](https://github.com/judepereira/checktodo): A GitHub PR checker that checks if PRs contain TODOs.
- [kynikos/report-todo](https://github.com/kynikos/report-todo): A generic reporting tool for TODOs.

## FAQ

### Why use this?

Tracking TODOs in code can help you have a cleaner and heathier code base.

1. It can help you realize when issues you thought were complete actually
   require some additional work (See [`actions/issue-reopener`]).
2. It makes it easier for contributors to find areas of the code that need work.
3. It makes it easier for contributors to find the relevant code for an issue.

### Why not just use `grep`?

`grep` is an amazing and blazingly fast tool. However, there are a few reasons
why you might use `todos`.

1. `grep` doesn't have much knowledge of code and languages so it's difficult to
   differentiate between comments and code. `todos` will ignore matches in code
   and only prints TODOs found it comments. It also ignores matches that occur
   in strings.
2. `grep` doesn't know about repository structure. It doesn't have inherant
   knowledge of VCS directories (e.g. `.git`) or vendored dependencies. It can't
   make use of `.gitignore` or other hints.
3. `todos` will parse TODO comments and can output in different formats such as
   JSON. This gives users an easy way to search for TODOs with their username,
   or with a specific issue number.

## Contributing

See [CONTRIBUTING.md] for contributor documentation.

[`todos` CLI]: #todos-cli-tool
[`actions/issue-reopener`]: actions/issue-reopener/README.md
[FAQ]: #faq
[CONTRIBUTING.md]: CONTRIBUTING.md

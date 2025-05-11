# Security Policy

This document describes the security policy that applies to this repository.

## Supported Versions

Security updates for this repository will be applied the most recent major
version and its minor versions.

For example if 2.2.0 is the latest version:

| Version | Supported          |
| ------- | ------------------ |
| v2.2.x  | :white_check_mark: |
| v2.1.x  | :white_check_mark: |
| v2.0.x  | :white_check_mark: |
| < 2.0.0 | :x:                |

However, if the repository has not made a stable release (e.g. the latest
release is < v1.0.0) then only the most latest minor version will be patched.

## Security Release & Disclosure Process

Security vulnerabilities should be handled quickly and sometimes privately. The
primary goal of this process is to reduce the total time users are vulnerable
to publicly known exploits.

### Private Disclosure

We ask that all suspected vulnerabilities be privately and responsibly
disclosed via the [private disclosure process](#reporting-a-vulnerability)
outlined above.

Fixes may be developed and tested by in a [temporary private
fork](https://docs.github.com/en/code-security/security-advisories/repository-security-advisories/collaborating-in-a-temporary-private-fork-to-resolve-a-repository-security-vulnerability)
that is private from the general public if deemed necessary.

### Public Disclosure

Vulnerabilities are disclosed publicly as GitHub [Security
Advisories](../security/advisories).

A public disclosure date is negotiated with the report submitter. We prefer to
fully disclose the bug as soon as possible once a user mitigation is available.
It is reasonable to delay disclosure when the bug or the fix is not yet fully
understood, the solution is not well-tested, or for vendor coordination. The
time frame for disclosure is from immediate (especially if it's already publicly
known) to several weeks. For a vulnerability with a straightforward mitigation,
we expect report date to disclosure date to be on the order of 14 days.

If you know of a publicly disclosed security vulnerability please IMMEDIATELY
[report the vulnerability](#reporting-a-vulnerability) so that the patch,
release, and communication process can be started as early as possible.

If the reporter does not go through the private disclosure process, the fix and
release process will proceed as swiftly as possible. In extreme cases you can
ask GitHub to delete the issue but this generally isn't necessary and is
unlikely to make a public disclosure less damaging.

### Security Releases

Once a fix is available it will be released, the GitHub Security Advisory made
public and announced via project communication channels. Security releases
will clearly marked as a security release and include information on which
vulnerabilities were fixed. As much as possible this announcement should be
actionable, and include any mitigating steps users can take prior to upgrading
to a fixed version.

Fixes will be applied in patch releases to all [supported
versions](#supported-versions) and all fixed vulnerabilities will be noted in
the [CHANGELOG](./CHANGELOG.md).

### Severity

Vulnerability severity is evaluated on a case-by-case basis, guided by [CVSS
3.1](https://www.first.org/cvss/v3.1/specification-document).

## Security Posture

We aim to reduce the number of security issues through several general
security-conscious development practices including the use of unit-tests,
end-to-end (e2e) tests, static and dynamic analysis tools, and use of
memory-safe languages.

We aim to fix issues discovered by analysis tools as quickly as possible. We
prefer to add these tools to "pre-submit" checks on PRs so that issues are
never added to the code in the first place.

In general, we observe the following security-conscious practices during
development (This is not an exhaustive list).

- Where possible, all PRs are reviewed by at least one [CODEOWNER](./CODEOWNERS).
- All unit and linter pre-submit tests must pass before a PRs is merged. See
  the [pre-submits](./CONTRIBUTING.md#pre-submits) section of the Contributor
  Guide for more information.
- All releases include no known test or linter failures.
- We refrain from using memory-unsafe languages (e.g. C, C++) or memory-unsafe
  use of languages that are memory-safe by default (e.g. the Go
  [unsafe](https://pkg.go.dev/unsafe) package). Use of these modules is
  restricted by static analysis tools.

## Reporting a Vulnerability

Vulnerabilities can be reported privately via GitHub's [Security
Advisories](https://docs.github.com/en/code-security/security-advisories)
feature.

Please see [Privately reporting a security
vulnerability](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing/privately-reporting-a-security-vulnerability#privately-reporting-a-security-vulnerability)
for more information on how to submit a vulnerability using GitHub's interface.

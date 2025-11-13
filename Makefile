# Copyright 2024 Ian Lewis
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Set the initial shell so we can determine extra options.
SHELL := /usr/bin/env bash -ueo pipefail
DEBUG_LOGGING ?= $(shell if [[ "${GITHUB_ACTIONS}" == "true" ]] && [[ -n "${RUNNER_DEBUG}" || "${ACTIONS_RUNNER_DEBUG}" == "true" || "${ACTIONS_STEP_DEBUG}" == "true" ]]; then echo "true"; else echo ""; fi)
BASH_OPTIONS := $(shell if [ "$(DEBUG_LOGGING)" == "true" ]; then echo "-x"; else echo ""; fi)

# Add extra options for debugging.
SHELL := /usr/bin/env bash -ueo pipefail $(BASH_OPTIONS)

uname_s := $(shell uname -s)
uname_m := $(shell uname -m)
arch.x86_64 := amd64
arch.arm64 := arm64
arch := $(arch.$(uname_m))
kernel.Linux := linux
kernel.Darwin := darwin
kernel := $(kernel.$(uname_s))

OUTPUT_FORMAT ?= $(shell if [ "${GITHUB_ACTIONS}" == "true" ]; then echo "github"; else echo ""; fi)
REPO_ROOT := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
REPO_NAME := $(shell basename "$(REPO_ROOT)")

# renovate: datasource=github-releases depName=aquaproj/aqua versioning=loose
AQUA_VERSION ?= v2.55.1
AQUA_REPO := github.com/aquaproj/aqua
AQUA_CHECKSUM.Linux.x86_64 = 7371b9785e07c429608a21e4d5b17dafe6780dabe306ec9f4be842ea754de48a
AQUA_CHECKSUM.Darwin.arm64 = cdaa13dd96187622ef5bee52867c46d4cf10765963423dc8e867c7c4decccf4d
AQUA_CHECKSUM ?= $(AQUA_CHECKSUM.$(uname_s).$(uname_m))
AQUA_URL := https://$(AQUA_REPO)/releases/download/$(AQUA_VERSION)/aqua_$(kernel)_$(arch).tar.gz
export AQUA_ROOT_DIR = $(REPO_ROOT)/.aqua

# Ensure that aqua and aqua installed tools are in the PATH.
export PATH := $(REPO_ROOT)/.bin/aqua-$(AQUA_VERSION):$(AQUA_ROOT_DIR)/bin:$(PATH)

# We want GNU versions of tools so prefer them if present.
GREP := $(shell command -v ggrep 2>/dev/null || command -v grep 2>/dev/null)
AWK := $(shell command -v gawk 2>/dev/null || command -v awk 2>/dev/null)
MKTEMP := $(shell command -v gmktemp 2>/dev/null || command -v mktemp 2>/dev/null)

BENCHTIME ?= 1s
TESTCOUNT ?= 1

# The help command prints targets in groups. Help documentation in the Makefile
# uses comments with double hash marks (##). Documentation is printed by the
# help target in the order in appears in the Makefile.
#
# Make targets can be documented with double hash marks as follows:
#
#	target-name: ## target documentation.
#
# Groups can be added with the following style:
#
#	## Group name

.PHONY: help
help: ## Print all Makefile targets (this message).
	@# bash \
	echo "$(REPO_NAME) Makefile"; \
	echo "Usage: $(MAKE) [COMMAND]"; \
	echo ""; \
	normal=""; \
	cyan=""; \
	if command -v tput >/dev/null 2>&1; then \
		if [ -t 1 ]; then \
			normal=$$(tput sgr0); \
			cyan=$$(tput setaf 6); \
		fi; \
	fi; \
	$(GREP) --no-filename -E '^([/a-z.A-Z0-9_%-]+:.*?|)##' $(MAKEFILE_LIST) | \
		$(AWK) \
			--assign=normal="$${normal}" \
			--assign=cyan="$${cyan}" \
			'BEGIN {FS = "(:.*?|)## ?"}; { \
				if (length($$1) > 0) { \
					printf("  " cyan "%-25s" normal " %s\n", $$1, $$2); \
				} else { \
					if (length($$2) > 0) { \
						printf("%s\n", $$2); \
					} \
				} \
			}'

package-lock.json: package.json $(AQUA_ROOT_DIR)/.installed
	@# bash \
	loglevel="notice"; \
	if [ -n "$(DEBUG_LOGGING)" ]; then \
		loglevel="verbose"; \
	fi; \
	# NOTE: npm install will happily ignore the fact that integrity hashes are \
	# missing in the package-lock.json. We need to check for missing integrity \
	# fields ourselves. If any are missing, then we need to regenerate the \
	# package-lock.json from scratch. \
	nointegrity=""; \
	noresolved=""; \
	if [ -f "$@" ]; then \
		nointegrity=$$(jq '.packages | del(."") | .[] | select(has("integrity") | not)' < $@); \
		noresolved=$$(jq '.packages | del(."") | .[] | select(has("resolved") | not)' < $@); \
	fi; \
	if [ ! -f "$@" ] || [ -n "$${nointegrity}" ] || [ -n "$${noresolved}" ]; then \
		# NOTE: package-lock.json is removed to ensure that npm includes the \
		# integrity field. npm install will not restore this field if \
		# missing in an existing package-lock.json file. \
		rm -f $@; \
		npm --loglevel="$${loglevel}" install \
			--no-audit \
			--no-fund; \
	else \
		npm --loglevel="$${loglevel}" install \
			--package-lock-only \
			--no-audit \
			--no-fund; \
	fi; \

node_modules/.installed: package-lock.json
	@# bash \
	loglevel="silent"; \
	if [ -n "$(DEBUG_LOGGING)" ]; then \
		loglevel="verbose"; \
	fi; \
	npm --loglevel="$${loglevel}" clean-install; \
	npm --loglevel="$${loglevel}" audit signatures; \
	touch $@

.venv/bin/activate:
	@# bash \
	python -m venv .venv

.venv/.installed: requirements-dev.txt .venv/bin/activate
	@# bash \
	$(REPO_ROOT)/.venv/bin/pip install -r $< --require-hashes; \
	touch $@

.bin/aqua-$(AQUA_VERSION)/aqua:
	@# bash \
	mkdir -p .bin/aqua-$(AQUA_VERSION); \
	tempfile=$$($(MKTEMP) --suffix=".aqua-$(AQUA_VERSION).tar.gz"); \
	curl -sSLo "$${tempfile}" "$(AQUA_URL)"; \
	echo "$(AQUA_CHECKSUM)  $${tempfile}" | shasum -a 256 -c; \
	tar -x -C .bin/aqua-$(AQUA_VERSION) -f "$${tempfile}"

$(AQUA_ROOT_DIR)/.installed: .aqua.yaml .bin/aqua-$(AQUA_VERSION)/aqua
	@# bash \
	loglevel="info"; \
	if [ -n "$(DEBUG_LOGGING)" ]; then \
		loglevel="debug"; \
	fi; \
	$(REPO_ROOT)/.bin/aqua-$(AQUA_VERSION)/aqua \
		--log-level "$${loglevel}" \
		--config .aqua.yaml \
		install; \
	touch $@

## Build
#####################################################################

.PHONY: all
all: test build-all build-npm docker-image ## Run all tests and build everything.

GO_SOURCE_FILES := $(shell git ls-files --deduplicate '*.go')

.PHONY: build
build: todos-$(kernel)-$(arch) ## Build the main binary for the current platform.

.PHONY: build-all
build-all: todos-linux-amd64 todos-linux-arm64 todos-darwin-amd64 todos-darwin-arm64 todos-windows-amd64 todos-windows-arm64 ## Build todos for all platforms.

.PHONY: build-with-pprof
build-with-pprof: todos-with-pprof-$(kernel)-$(arch) ## Build todos with profiling for current platform.

.PHONY: build-with-pprof-all
build-with-pprof-all: todos-with-pprof-linux-amd64 todos-with-pprof-linux-arm64 todos-with-pprof-darwin-amd64 todos-with-pprof-darwin-arm64 todos-with-pprof-windows-amd64 todos-with-pprof-windows-arm64 ## Build todos with profiling for all platforms.

.PHONY: build-npm
build-npm: node_modules/.installed build-all ## Build npm package tarball.
	@# bash \
	# NOTE: npm tarball is for local use only and is not used in releases. \
	npm pack; \
	cp todos-linux-amd64 packages/todos-linux-amd64/todos; \
	(cd packages/todos-linux-amd64 && npm pack); \
	cp todos-linux-arm64 packages/todos-linux-arm64/todos; \
	(cd packages/todos-linux-arm64 && npm pack); \
	cp todos-darwin-amd64 packages/todos-darwin-amd64/todos; \
	(cd packages/todos-darwin-amd64 && npm pack); \
	cp todos-darwin-arm64 packages/todos-darwin-arm64/todos; \
	(cd packages/todos-darwin-arm64 && npm pack); \
	cp todos-windows-amd64 packages/todos-windows-amd64/todos.exe; \
	(cd packages/todos-windows-amd64 && npm pack); \
	cp todos-windows-arm64 packages/todos-windows-arm64/todos.exe; \
	(cd packages/todos-windows-arm64 && npm pack)

todos-with-pprof-%: $(GO_SOURCE_FILES)
	# NOTE: $@ is for local use only and is not used in releases.
	@# bash \
	go mod vendor; \
	CGO_ENABLED=0 \
	GOOS=$(word 1,$(subst -, ,$*)) \
	GOARCH=$(word 2,$(subst -, ,$*)) \
		go build \
			-o todos-with-pprof-$* \
			-trimpath \
			-mod=vendor \
			-tags=netgo,profile \
			-ldflags="-s -w" \
			github.com/ianlewis/todos/cmd/todos

todos-%: $(GO_SOURCE_FILES)
	# NOTE: $@ is for local use only and is not used in releases.
	@# bash \
	go mod vendor; \
	CGO_ENABLED=0 \
	GOOS=$(word 1,$(subst -, ,$*)) \
	GOARCH=$(word 2,$(subst -, ,$*)) \
		go build \
			-o todos-$* \
			-trimpath \
			-mod=vendor \
			-tags=netgo \
			-ldflags="-s -w" \
			github.com/ianlewis/todos/cmd/todos

.PHONY: docker-image
docker-image: build-all ## Build Docker image.
	# NOTE: The Docker image is for local use only and is not used in releases.
	docker build \
		-t ghcr.io/ianlewis/todos \
		.

## Testing
#####################################################################

.PHONY: test
test: lint unit-test ## Run all linters and tests.

.PHONY: unit-test
unit-test: ## Runs all unit tests.
	@# bash \
	go mod vendor; \
	extraargs=""; \
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
		extraargs="-v"; \
	fi; \
	# TODO(github.com/golang/go/issues/75031): remove when 'covdata' issue is fixed. \
	GOTOOLCHAIN=$$(go version | $(AWK) '{ print $$3 }')+auto \
		go test \
			$$extraargs \
			-mod=vendor \
			-race \
			-coverprofile=coverage.out \
			-coverpkg=./... \
			-covermode=atomic \
			./...

## Benchmarking
#####################################################################

.PHONY: go-benchmark
go-benchmark: ## Runs Go benchmarks.
	@# bash \
	go mod vendor; \
	extraargs=""; \
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
		extraargs="-v"; \
	fi; \
	go test \
		$$extraargs \
		-mod=vendor \
		-bench=. \
		-count=$(TESTCOUNT) \
		-benchtime=$(BENCHTIME) \
		-run='^#' \
		./...

## Formatting
#####################################################################

.PHONY: format
format: go-format json-format license-headers md-format yaml-format ## Format all files

.PHONY: go-format
go-format: $(AQUA_ROOT_DIR)/.installed ## Format Go files (gofumpt).
	@# bash \
	files=$$( \
		git ls-files --deduplicate \
			'*.go' \
	); \
	if [ "$${files}" == "" ]; then \
		exit 0; \
	fi; \
	gofumpt -l -w $${files}; \
	goimports -l -w $${files}; \
	gci write \
		--skip-generated \
		--skip-vendor \
		-s standard \
		-s default \
		-s localmodule \
		$${files}

.PHONY: json-format
json-format: node_modules/.installed ## Format JSON files.
	@# bash \
	loglevel="log"; \
	if [ -n "$(DEBUG_LOGGING)" ]; then \
		loglevel="debug"; \
	fi; \
	files=$$( \
		git ls-files --deduplicate \
			'*.json' \
			'*.json5' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done \
	); \
	if [ "$${files}" == "" ]; then \
		exit 0; \
	fi; \
	$(REPO_ROOT)/node_modules/.bin/prettier \
		--log-level "$${loglevel}" \
		--no-error-on-unmatched-pattern \
		--write \
		$${files}

.PHONY: license-headers
license-headers: ## Update license headers.
	@# bash \
	files=$$( \
		git ls-files --deduplicate \
			'*.c' \
			'*.cpp' \
			'*.go' \
			'*.h' \
			'*.hpp' \
			'*.js' \
			'*.lua' \
			'*.py' \
			'*.rb' \
			'*.rs' \
			'*.yaml' \
			'*.yml' \
			'Makefile' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done \
	); \
	name=$$(git config user.name); \
	if [ "$${name}" == "" ]; then \
		>&2 echo "git user.name is required."; \
		>&2 echo "Set it up using:"; \
		>&2 echo "git config user.name \"John Doe\""; \
		exit 1; \
	fi; \
	for filename in $${files}; do \
		if ! ( head "$${filename}" | $(GREP) -iL "Copyright" > /dev/null ); then \
			$(REPO_ROOT)/third_party/mbrukman/autogen/autogen.sh \
				--in-place \
				--no-code \
				--no-tlc \
				--copyright "$${name}" \
				--license apache \
				"$${filename}"; \
		fi; \
	done

.PHONY: md-format
md-format: node_modules/.installed ## Format Markdown files.
	@# bash \
	loglevel="log"; \
	if [ -n "$(DEBUG_LOGGING)" ]; then \
		loglevel="debug"; \
	fi; \
	files=$$( \
		git ls-files --deduplicate \
			'*.md' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done \
	); \
	if [ "$${files}" == "" ]; then \
		exit 0; \
	fi; \
	# NOTE: prettier uses .editorconfig for tab-width. \
	$(REPO_ROOT)/node_modules/.bin/prettier \
		--log-level "$${loglevel}" \
		--no-error-on-unmatched-pattern \
		--write \
		$${files}

.PHONY: yaml-format
yaml-format: node_modules/.installed ## Format YAML files.
	@# bash \
	loglevel="log"; \
	if [ -n "$(DEBUG_LOGGING)" ]; then \
		loglevel="debug"; \
	fi; \
	files=$$( \
		git ls-files --deduplicate \
			'*.yml' \
			'*.yaml' \
	); \
	if [ "$${files}" == "" ]; then \
		exit 0; \
	fi; \
	$(REPO_ROOT)/node_modules/.bin/prettier \
		--log-level "$${loglevel}" \
		--no-error-on-unmatched-pattern \
		--write \
		$${files}

## Linting
#####################################################################

.PHONY: lint
lint: actionlint checkmake commitlint docs-check eslint fixme format-check golangci-lint hadolint markdownlint renovate-config-validator textlint yamllint zizmor ## Run all linters.

.PHONY: actionlint
actionlint: $(AQUA_ROOT_DIR)/.installed ## Runs the actionlint linter.
	@# bash \
	# NOTE: We need to ignore config files used in tests. \
	files=$$( \
		git ls-files --deduplicate \
			'.github/workflows/*.yml' \
			'.github/workflows/*.yaml' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done \
	); \
	if [ "$${files}" == "" ]; then \
		exit 0; \
	fi; \
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
		actionlint \
			-format '{{range $$err := .}}::error file={{$$err.Filepath}},line={{$$err.Line}},col={{$$err.Column}}::{{$$err.Message}}%0A```%0A{{replace $$err.Snippet "\\n" "%0A"}}%0A```\n{{end}}' \
			-ignore 'SC2016:' \
			$${files}; \
	else \
		actionlint \
			-ignore 'SC2016:' \
			$${files}; \
	fi

.PHONY: checkmake
checkmake: $(AQUA_ROOT_DIR)/.installed ## Runs the checkmake linter.
	@# bash \
	# NOTE: We need to ignore config files used in tests. \
	files=$$( \
		git ls-files --deduplicate \
			'Makefile' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done \
	); \
	if [ "$${files}" == "" ]; then \
		exit 0; \
	fi; \
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
		# TODO: Remove newline from the format string after updating checkmake. \
		checkmake \
			--config .checkmake.ini \
			--format '::error file={{.FileName}},line={{.LineNumber}}::{{.Rule}}: {{.Violation}}'$$'\n' \
			$${files}; \
	else \
		checkmake \
			--config .checkmake.ini \
			$${files}; \
	fi

.PHONY: commitlint
commitlint: node_modules/.installed ## Run commitlint linter.
	@# bash \
	commitlint_from=$(COMMITLINT_FROM_REF); \
	commitlint_to=$(COMMITLINT_TO_REF); \
	if [ "$${commitlint_from}" == "" ]; then \
		# Try to get the default branch without hitting the remote server \
		if git symbolic-ref --short refs/remotes/origin/HEAD >/dev/null 2>&1; then \
			commitlint_from=$$(git symbolic-ref --short refs/remotes/origin/HEAD); \
		elif git show-ref refs/remotes/origin/master >/dev/null 2>&1; then \
			commitlint_from="origin/master"; \
		else \
			commitlint_from="origin/main"; \
		fi; \
	fi; \
	if [ "$${commitlint_to}" == "" ]; then \
		# if head is on the commitlint_from branch, then we will lint the \
		# last commit by default. \
		current_branch=$$(git rev-parse --abbrev-ref HEAD); \
		if [ "$${commitlint_from}" == "$${current_branch}" ]; then \
			commitlint_from="HEAD~1"; \
		fi; \
		commitlint_to="HEAD"; \
	fi; \
	$(REPO_ROOT)/node_modules/.bin/commitlint \
		--config commitlint.config.mjs \
		--from "$${commitlint_from}" \
		--to "$${commitlint_to}" \
		--verbose \
		--strict

.PHONY: docs-check
docs-check: ## Check that generated documentation is up to date.
	@# bash \
	if [ -n "$$(git diff)" ]; then \
		>&2 echo "The working directory is dirty. Please commit, stage, or stash changes and try again."; \
		exit 1; \
	fi; \
	$(MAKE) docs; \
	exit_code=0; \
	if [ -n "$$(git diff)" ]; then \
		>&2 echo "Some files need to be generated. Please run '$(MAKE) docs' and try again."; \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			echo "::group::git diff"; \
		fi; \
		git --no-pager diff; \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			echo "::endgroup::"; \
		fi; \
		exit_code=1; \
	fi; \
	git restore .; \
	exit "$${exit_code}"

.PHONY: eslint
eslint: node_modules/.installed ## Runs eslint.
	@# bash \
	extraargs=""; \
	if [ -n "$(DEBUG_LOGGING)" ]; then \
		extraargs="--debug"; \
	fi; \
	files=$$( \
		git ls-files --deduplicate \
			'*.js' \
			'*.cjs' \
			'*.mjs' \
			'*.jsx' \
			'*.mjsx' \
			'*.ts' \
			'*.cts' \
			'*.mts' \
			'*.tsx' \
			'*.mtsx' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done \
	); \
	if [ "$${files}" == "" ]; then \
		exit 0; \
	fi; \
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
		exit_code=0; \
		while IFS="" read -r p && [ -n "$${p}" ]; do \
			file=$$(echo "$${p}" | jq -c '.filePath // empty' | tr -d '"'); \
			while IFS="" read -r m && [ -n "$${m}" ]; do \
				severity=$$(echo "$${m}" | jq -c '.severity // empty' | tr -d '"'); \
				line=$$(echo "$${m}" | jq -c '.line // empty' | tr -d '"'); \
				endline=$$(echo "$${m}" | jq -c '.endLine // empty' | tr -d '"'); \
				col=$$(echo "$${m}" | jq -c '.column // empty' | tr -d '"'); \
				endcol=$$(echo "$${m}" | jq -c '.endColumn // empty' | tr -d '"'); \
				message=$$(echo "$${m}" | jq -c '.message // empty' | tr -d '"'); \
				exit_code=1; \
				case $${severity} in \
				"1") \
					echo "::warning file=$${file},line=$${line},endLine=$${endline},col=$${col},endColumn=$${endcol}::$${message}"; \
					;; \
				"2") \
					echo "::error file=$${file},line=$${line},endLine=$${endline},col=$${col},endColumn=$${endcol}::$${message}"; \
					;; \
				esac; \
			done <<<$$(echo "$${p}" | jq -c '.messages[]'); \
		done <<<$$($(REPO_ROOT)/node_modules/.bin/eslint \
			--max-warnings 0 \
			--format json \
			$${extraargs} \
			$${files} | jq -c '.[]'); \
		exit "$${exit_code}"; \
	else \
		$(REPO_ROOT)/node_modules/.bin/eslint \
			--max-warnings 0 \
			$${extraargs} \
			$${files}; \
	fi

.PHONY: fixme
fixme: $(AQUA_ROOT_DIR)/.installed ## Check for outstanding FIXMEs.
	@# bash \
	output="default"; \
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
		output="github"; \
	fi; \
	# NOTE: todos does not use `git ls-files` because many files might be \
	# 		unsupported and generate an error if passed directly on the \
	# 		command line. \
	todos \
		--output "$${output}" \
		--todo-types="FIXME,Fixme,fixme,BUG,Bug,bug,XXX,COMBAK"

.PHONY: format-check
format-check: ## Check that files are properly formatted.
	@# bash \
	if [ -n "$$(git diff)" ]; then \
		>&2 echo "The working directory is dirty. Please commit, stage, or stash changes and try again."; \
		exit 1; \
	fi; \
	$(MAKE) format; \
	exit_code=0; \
	if [ -n "$$(git diff)" ]; then \
		>&2 echo "Some files need to be formatted. Please run 'make format' and try again."; \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			echo "::group::git diff"; \
		fi; \
		git --no-pager diff; \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			echo "::endgroup::"; \
		fi; \
		exit_code=1; \
	fi; \
	git restore .; \
	exit "$${exit_code}"

.PHONY: golangci-lint
golangci-lint: $(AQUA_ROOT_DIR)/.installed ## Runs the golangci-lint linter.
	@# bash \
	golangci-lint run -c .golangci.yml ./...

.PHONY: hadolint
hadolint: $(AQUA_ROOT_DIR)/.installed ## Runs the hadolint linter.
	@set -euo pipefail;\
		files=$$( \
			git ls-files --deduplicate \
				'[Dd]ockerfile' \
				'[Cc]ontainerfile' \
				| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done \
		); \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			hadolint -f checkstyle $${files}; \
		else \
			hadolint $${files}; \
		fi

.PHONY: markdownlint
markdownlint: node_modules/.installed $(AQUA_ROOT_DIR)/.installed ## Runs the markdownlint linter.
	@# bash \
	# NOTE: Issue and PR templates are handled specially so we can disable \
	# MD041/first-line-heading/first-line-h1 without adding an ugly html comment \
	# at the top of the file. \
	files=$$( \
		git ls-files --deduplicate \
			'*.md' \
			':!:.github/pull_request_template.md' \
			':!:.github/ISSUE_TEMPLATE/*.md' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done \
	); \
	if [ "$${files}" == "" ]; then \
		exit 0; \
	fi; \
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
		exit_code=0; \
		while IFS="" read -r p && [ -n "$$p" ]; do \
			file=$$(echo "$$p" | jq -cr '.fileName // empty'); \
			line=$$(echo "$$p" | jq -cr '.lineNumber // empty'); \
			endline=$${line}; \
			message=$$(echo "$$p" | jq -cr '.ruleNames[0] + "/" + .ruleNames[1] + " " + .ruleDescription + " [Detail: \"" + .errorDetail + "\", Context: \"" + .errorContext + "\"]"'); \
			exit_code=1; \
			echo "::error file=$${file},line=$${line},endLine=$${endline}::$${message}"; \
		done <<< "$$($(REPO_ROOT)/node_modules/.bin/markdownlint --config .markdownlint.yaml --dot --json $${files} 2>&1 | jq -c '.[]')"; \
		if [ "$${exit_code}" != "0" ]; then \
			exit "$${exit_code}"; \
		fi; \
	else \
		$(REPO_ROOT)/node_modules/.bin/markdownlint \
			--config .markdownlint.yaml \
			--dot \
			$${files}; \
	fi; \
	files=$$( \
		git ls-files --deduplicate \
			'.github/pull_request_template.md' \
			'.github/ISSUE_TEMPLATE/*.md' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done \
	); \
	if [ "$${files}" == "" ]; then \
		exit 0; \
	fi; \
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
		exit_code=0; \
		while IFS="" read -r p && [ -n "$$p" ]; do \
			file=$$(echo "$$p" | jq -cr '.fileName // empty'); \
			line=$$(echo "$$p" | jq -cr '.lineNumber // empty'); \
			endline=$${line}; \
			message=$$(echo "$$p" | jq -cr '.ruleNames[0] + "/" + .ruleNames[1] + " " + .ruleDescription + " [Detail: \"" + .errorDetail + "\", Context: \"" + .errorContext + "\"]"'); \
			exit_code=1; \
			echo "::error file=$${file},line=$${line},endLine=$${endline}::$${message}"; \
		done <<< "$$($(REPO_ROOT)/node_modules/.bin/markdownlint --config .github/template.markdownlint.yaml --dot --json $${files} 2>&1 | jq -c '.[]')"; \
		if [ "$${exit_code}" != "0" ]; then \
			exit "$${exit_code}"; \
		fi; \
	else \
		$(REPO_ROOT)/node_modules/.bin/markdownlint \
			--config .github/template.markdownlint.yaml \
			--dot \
			$${files}; \
	fi

.PHONY: renovate-config-validator
renovate-config-validator: node_modules/.installed ## Validate Renovate configuration.
	@# bash \
	$(REPO_ROOT)/node_modules/.bin/renovate-config-validator \
		--strict

.PHONY: textlint
textlint: node_modules/.installed $(AQUA_ROOT_DIR)/.installed ## Runs the textlint linter.
	@# bash \
	files=$$( \
		git ls-files --deduplicate \
			'*.md' \
			'*.txt' \
			':!:requirements*.txt' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done \
	); \
	if [ "$${files}" == "" ]; then \
		exit 0; \
	fi; \
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
		exit_code=0; \
		while IFS="" read -r p && [ -n "$$p" ]; do \
			filePath=$$(echo "$$p" | jq -cr '.filePath // empty'); \
			file=$$(realpath --relative-to="." "$${filePath}"); \
			while IFS="" read -r m && [ -n "$$m" ]; do \
				line=$$(echo "$$m" | jq -cr '.loc.start.line // empty'); \
				endline=$$(echo "$$m" | jq -cr '.loc.end.line // empty'); \
				col=$$(echo "$${m}" | jq -cr '.loc.start.column // empty'); \
				endcol=$$(echo "$${m}" | jq -cr '.loc.end.column // empty'); \
				message=$$(echo "$$m" | jq -cr '.message // empty'); \
				exit_code=1; \
				echo "::error file=$${file},line=$${line},endLine=$${endline},col=$${col},endColumn=$${endcol}::$${message}"; \
			done <<<"$$(echo "$$p" | jq -cr '.messages[] // empty')"; \
		done <<< "$$($(REPO_ROOT)/node_modules/.bin/textlint -c .textlintrc.yaml --format json $${files} 2>&1 | jq -c '.[]')"; \
		exit "$${exit_code}"; \
	else \
		$(REPO_ROOT)/node_modules/.bin/textlint \
			--config .textlintrc.yaml \
			$${files}; \
	fi

.PHONY: yamllint
yamllint: .venv/.installed ## Runs the yamllint linter.
	@# bash \
	files=$$( \
		git ls-files --deduplicate \
			'*.yml' \
			'*.yaml' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done \
	); \
	if [ "$${files}" == "" ]; then \
		exit 0; \
	fi; \
	format="standard"; \
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
		format="github"; \
	fi; \
	$(REPO_ROOT)/.venv/bin/yamllint \
		--strict \
		--config-file .yamllint.yaml \
		--format "$${format}" \
		$${files}

.PHONY: zizmor
zizmor: .venv/.installed ## Runs the zizmor linter.
	@# bash \
	# NOTE: On GitHub actions this outputs SARIF format to zizmor.sarif.json \
	#       in addition to outputting errors to the terminal. \
	files=$$( \
		git ls-files --deduplicate \
			'.github/workflows/*.yml' \
			'.github/workflows/*.yaml' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done \
	); \
	if [ "$${files}" == "" ]; then \
		exit 0; \
	fi; \
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
		$(REPO_ROOT)/.venv/bin/zizmor \
			--config .zizmor.yml \
			--quiet \
			--pedantic \
			--format sarif \
			$${files} > zizmor.sarif.json; \
	fi; \
	$(REPO_ROOT)/.venv/bin/zizmor \
		--config .zizmor.yml \
		--quiet \
		--pedantic \
		--format plain \
		$${files}

## Documentation
#####################################################################

.PHONY: docs
docs: SUPPORTED_LANGUAGES.md ## Generate all documentation.

SUPPORTED_LANGUAGES.md: node_modules/.installed internal/scanner/languages.go ## Supported languages documentation.
	@# bash \
	go mod vendor; \
	go run \
		-mod=vendor \
		./internal/cmd/genlangdocs | \
			./node_modules/.bin/prettier \
				--parser markdown > $@

## Maintenance
#####################################################################

.PHONY: todos
todos: $(AQUA_ROOT_DIR)/.installed ## Print outstanding TODOs.
	@# bash \
	output="default"; \
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
		output="github"; \
	fi; \
	# NOTE: todos does not use `git ls-files` because many files might be \
	# 		unsupported and generate an error if passed directly on the \
	# 		command line. \
	todos \
		--output "$${output}" \
		--todo-types="TODO,Todo,todo,FIXME,Fixme,fixme,BUG,Bug,bug,XXX,COMBAK"

.PHONY: clean
clean: ## Delete temporary files.
	@$(RM) -r .bin
	@$(RM) -r $(AQUA_ROOT_DIR)
	@$(RM) -r .venv
	@$(RM) -r node_modules
	@$(RM) *.sarif.json
	@$(RM) -r vendor
	@$(RM) coverage.out
	@$(RM) todos \
	@$(RM) todos-* \
	@$(RM) ianlewis-todos-*.tgz \
	@$(RM) packages/todos-linux-amd64/todos \
	@$(RM) packages/todos-linux-amd64/*.tgz \
	@$(RM) packages/todos-linux-arm64/todos \
	@$(RM) packages/todos-linux-arm64/*.tgz \
	@$(RM) packages/todos-darwin-amd64/todos \
	@$(RM) packages/todos-darwin-amd64/*.tgz \
	@$(RM) packages/todos-darwin-arm64/todos \
	@$(RM) packages/todos-darwin-arm64/*.tgz \
	@$(RM) packages/todos-windows-amd64/todos \
	@$(RM) packages/todos-windows-amd64/*.tgz \
	@$(RM) packages/todos-windows-arm64/todos \
	@$(RM) packages/todos-windows-arm64/*.tgz

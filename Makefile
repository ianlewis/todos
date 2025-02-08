# Copyright 2023 Google LLC
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

SHELL := /bin/bash
OUTPUT_FORMAT ?= $(shell if [ "${GITHUB_ACTIONS}" == "true" ]; then echo "github"; else echo ""; fi)
BENCHTIME ?= 1s
TESTCOUNT ?= 1
REPO_NAME = $(shell basename "$$(pwd)")

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
help: ## Shows all targets and help from the Makefile (this message).
	@echo "$(REPO_NAME) Makefile"
	@echo "Usage: make [COMMAND]"
	@echo ""
	@grep --no-filename -E '^([/a-z.A-Z0-9_%-]+:.*?|)##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = "(:.*?|)## ?"}; { \
			if (length($$1) > 0) { \
				printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2; \
			} else { \
				if (length($$2) > 0) { \
					printf "%s\n", $$2; \
				} \
			} \
		}'

package-lock.json:
	npm install

node_modules/.installed: package.json package-lock.json
	npm ci
	touch node_modules/.installed

.venv/bin/activate:
	python -m venv .venv

.venv/.installed: .venv/bin/activate requirements.txt
	./.venv/bin/pip install -r requirements.txt --require-hashes
	touch .venv/.installed

## Testing
#####################################################################

.PHONY: unit-test
unit-test: go-test ## Runs all unit tests.

.PHONY: go-test
go-test: ## Runs Go unit tests.
	@set -e;\
		go mod vendor; \
		extraargs=""; \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			extraargs="-v"; \
		fi; \
		go test $$extraargs -mod=vendor -race -coverprofile=coverage.out -covermode=atomic ./...

## Benchmarking
#####################################################################

.PHONY: go-benchmark
go-benchmark: ## Runs Go benchmarks.
	@set -e;\
		go mod vendor; \
		extraargs=""; \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			extraargs="-v"; \
		fi; \
		go test $$extraargs -mod=vendor -bench=. -count=$(TESTCOUNT) -benchtime=$(BENCHTIME) -run='^#' ./...

## Tools
#####################################################################

.PHONY: license-headers
license-headers: ## Update license headers.
	@set -euo pipefail; \
		files=$$( \
			git ls-files \
				'*.go' '**/*.go' \
				'*.ts' '**/*.ts' \
				'*.js' '**/*.js' \
				'*.py' '**/*.py' \
				'*.yaml' '**/*.yaml' \
				'*.yml' '**/*.yml' \
		); \
		name=$$(git config user.name); \
		if [ "$${name}" == "" ]; then \
			>&2 echo "git user.name is required."; \
			>&2 echo "Set it up using:"; \
			>&2 echo "git config user.name \"John Doe\""; \
		fi; \
		for filename in $${files}; do \
			if ! ( head "$${filename}" | grep -iL "Copyright" > /dev/null ); then \
				autogen -i --no-code --no-tlc -c "$${name}" -l apache "$${filename}"; \
			fi; \
		done; \
		if ! ( head Makefile | grep -iL "Copyright" > /dev/null ); then \
			autogen -i --no-code --no-tlc -c "$${name}" -l apache Makefile; \
		fi;

.PHONY: format
format: go-format md-format yaml-format ## Format all files

.PHONY: md-format
md-format: node_modules/.installed ## Format Markdown files.
	@set -euo pipefail; \
		files=$$( \
			git ls-files \
				'*.md' '**/*.md' \
				'*.markdown' '**/*.markdown' \
		); \
		npx prettier --write --no-error-on-unmatched-pattern $${files}

.PHONY: yaml-format
yaml-format: node_modules/.installed ## Format YAML files.
	@set -euo pipefail; \
		files=$$( \
			git ls-files \
				'*.yml' '**/*.yml' \
				'*.yaml' '**/*.yaml' \
		); \
		npx prettier --write --no-error-on-unmatched-pattern $${files}


.PHONY: go-format
go-format: ## Format Go files (gofumpt).
	@set -euo pipefail;\
		files=$$(git ls-files '*.go'); \
		if [ "$${files}" != "" ]; then \
			gofumpt -w $${files}; \
			gci write  --skip-generated -s standard -s default -s "prefix(github.com/ianlewis/todos)" $${files}; \
		fi

## Linters
#####################################################################

.PHONY: lint
lint: golangci-lint yamllint actionlint markdownlint ## Run all linters.

.PHONY: actionlint
actionlint: ## Runs the actionlint linter.
	@# NOTE: We need to ignore config files used in tests.
	@set -euo pipefail;\
		files=$$( \
			git ls-files \
				'.github/workflows/*.yml' \
				'.github/workflows/*.yaml' \
		); \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			actionlint -format '{{range $$err := .}}::error file={{$$err.Filepath}},line={{$$err.Line}},col={{$$err.Column}}::{{$$err.Message}}%0A```%0A{{replace $$err.Snippet "\\n" "%0A"}}%0A```\n{{end}}' -ignore 'SC2016:' $${files}; \
		else \
			actionlint $${files}; \
		fi

.PHONY: markdownlint
markdownlint: node_modules/.installed ## Runs the markdownlint linter.
	@set -euo pipefail;\
		files=$$( \
			git ls-files \
				'*.md' '**/*.md' \
				'*.markdown' '**/*.markdown' \
		); \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			exit_code=0; \
			while IFS="" read -r p && [ -n "$$p" ]; do \
				file=$$(echo "$$p" | jq -c -r '.fileName // empty'); \
				line=$$(echo "$$p" | jq -c -r '.lineNumber // empty'); \
				endline=$${line}; \
				message=$$(echo "$$p" | jq -c -r '.ruleNames[0] + "/" + .ruleNames[1] + " " + .ruleDescription + " [Detail: \"" + .errorDetail + "\", Context: \"" + .errorContext + "\"]"'); \
				exit_code=1; \
				echo "::error file=$${file},line=$${line},endLine=$${endline}::$${message}"; \
			done <<< "$$(npx markdownlint --dot --json $${files} 2>&1 | jq -c '.[]')"; \
			exit "$${exit_code}"; \
		else \
			npx markdownlint --dot $${files}; \
		fi

.PHONY: yamllint
yamllint: .venv/.installed ## Runs the yamllint linter.
	@set -euo pipefail;\
		extraargs=""; \
		files=$$( \
			git ls-files \
				'*.yml' '**/*.yml' \
				'*.yaml' '**/*.yaml' \
		); \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			extraargs="-f github"; \
		fi; \
		.venv/bin/yamllint --strict -c .yamllint.yaml $${extraargs} $${files}

.PHONY: golangci-lint
golangci-lint: ## Runs the golangci-lint linter.
	@golangci-lint run -c .golangci.yml ./...

## Documentation
#####################################################################

SUPPORTED_LANGUAGES.md: node_modules/.installed internal/scanner/languages.go ## Supported languages documentation.
	@set -e;\
		go mod vendor; \
		go run -mod=vendor ./internal/cmd/genlangdocs | ./node_modules/.bin/prettier --parser markdown > $@

## Maintenance
#####################################################################

.PHONY: clean
clean: ## Delete temporary files.
	rm -rf \
		.venv \
		node_modules \
		vendor \
		coverage.out

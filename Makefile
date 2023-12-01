# Copyright 2023 Google LLC
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

.PHONY: help
help: ## Shows all targets and help from the Makefile (this message).
	@echo "todos Makefile"
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

node_modules/.installed: package.json package-lock.json
	npm ci
	touch node_modules/.installed

## Testing
#####################################################################

.PHONY: unit-test
unit-test: go-test ts-test ## Runs all unit tests.

.PHONY: go-test
go-test: ## Runs Go unit tests.
	@set -e;\
		go mod vendor; \
		extraargs=""; \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			extraargs="-v"; \
		fi; \
		go test $$extraargs -mod=vendor -race -coverprofile=coverage.out -covermode=atomic ./...

.PHONY: ts-test
ts-test: ## Run TypeScript unit tests.
	@# Run unit tests for all TS actions where tests are found.
	@set -e;\
		PATHS=$$(find actions/ -not -path '*/node_modules/*' -name __tests__ -type d | xargs dirname); \
		for path in $$PATHS; do \
			make -C $$path unit-test; \
		done

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
		go test $$extraargs -bench=. -count=$(TESTCOUNT) -benchtime=$(BENCHTIME) -run='^#' ./...

## Tools
#####################################################################

.PHONY: autogen
autogen: ## Runs autogen on code files.
	@set -euo pipefail; \
		md_files=$$( \
			find . -type f \
				\( \
					-name '*.go' -o \
					-name '*.yaml' -o \
					-name '*.yml' \
				\) \
				-not -iwholename '*/.git/*' \
				-not -iwholename '*/vendor/*' \
				-not -iwholename '*/node_modules/*' \
		); \
		for filename in $${md_files}; do \
			if ! ( head "$${filename}" | grep -iL "Copyright" > /dev/null ); then \
				autogen -i --no-code --no-tlc -c "Google LLC" -l apache "$${filename}"; \
			fi; \
		done; \
		if ! ( head Makefile | grep -iL "Copyright" > /dev/null ); then \
			autogen -i --no-code --no-tlc -c "Google LLC" -l apache Makefile; \
		fi;


## Linters
#####################################################################

.PHONY: lint
lint: golangci-lint eslint yamllint actionlint markdownlint ## Run all linters.

.PHONY: actionlint
actionlint: ## Runs the actionlint linter.
	@# NOTE: We need to ignore config files used in tests.
	@set -e;\
		files=$$( \
			find .github/workflows/ -type f \
				\( \
					-name '*.yaml' -o \
					-name '*.yml' \
				\) \
		); \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			actionlint -format '{{range $$err := .}}::error file={{$$err.Filepath}},line={{$$err.Line}},col={{$$err.Column}}::{{$$err.Message}}%0A```%0A{{replace $$err.Snippet "\\n" "%0A"}}%0A```\n{{end}}' -ignore 'SC2016:' $${files}; \
		else \
			actionlint $${files}; \
		fi

.PHONY: eslint
eslint: ## Runs the eslint linter.
	@set -e;\
		PATHS=$$(find actions/ -not -path '*/node_modules/*' -name package.json -type f | xargs dirname); \
		for path in $$PATHS; do \
			make -C $$path lint; \
		done

.PHONY: markdownlint
markdownlint: node_modules/.installed ## Runs the markdownlint linter.
	@set -e;\
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			exit_code=0; \
			while IFS="" read -r p && [ -n "$$p" ]; do \
				file=$$(echo "$$p" | jq -c -r '.fileName // empty'); \
				line=$$(echo "$$p" | jq -c -r '.lineNumber // empty'); \
				endline=$${line}; \
				message=$$(echo "$$p" | jq -c -r '.ruleNames[0] + "/" + .ruleNames[1] + " " + .ruleDescription + " [Detail: \"" + .errorDetail + "\", Context: \"" + .errorContext + "\"]"'); \
				exit_code=1; \
				echo "::error file=$${file},line=$${line},endLine=$${endline}::$${message}"; \
			done <<< "$$(./node_modules/.bin/markdownlint --dot --json . 2>&1 | jq -c '.[]')"; \
			exit "$${exit_code}"; \
		else \
			npm run lint; \
		fi

.PHONY: golangci-lint
golangci-lint: ## Runs the golangci-lint linter.
	@set -e;\
		extraargs=""; \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			extraargs="--out-format github-actions"; \
		fi; \
		golangci-lint run -c .golangci.yml ./... $$extraargs

.PHONY: yamllint
yamllint: ## Runs the yamllint linter.
	@set -e;\
		extraargs=""; \
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			extraargs="-f github"; \
		fi; \
		yamllint --strict -c .yamllint.yaml . $$extraargs

## Documentation
#####################################################################

SUPPORTED_LANGUAGES.md: node_modules/.installed internal/scanner/languages.yml ## Supported languages documentation.
	@set -e;\
		go mod vendor; \
		go run ./internal/cmd/genlangdocs | ./node_modules/.bin/prettier --parser markdown > $@

## Maintenance
#####################################################################

.PHONY: clean
clean: ## Delete temporary files.
	rm -rf vendor node_modules coverage.out
	@set -e;\
		PATHS=$$(find actions/ -not -path '*/node_modules/*' -name package.json -type f | xargs dirname); \
		for path in $$PATHS; do \
			make -C $$path clean; \
		done

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
ACTION_NAME = $(shell basename "$$(pwd)")
OUTPUT_FORMAT ?= $(shell if [ "${GITHUB_ACTIONS}" == "true" ]; then echo "github"; else echo ""; fi)

REPO_ROOT := $(shell git rev-parse --show-toplevel)

.PHONY: help
help: ## Shows all targets and help from the Makefile (this message).
	@echo "$(ACTION_NAME) Makefile"
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

.PHONY: action
action: node_modules/.installed ## Builds the action.
	npm run build

.PHONY: package
package: action ## Builds the distribution package.
	npm run package

.PHONY: clean
clean:
	rm -rf dist lib node_modules

## Tools
#####################################################################

.PHONY: format
format: node_modules/.installed ## Formats code.
	npm run format

## Testing
#####################################################################

.PHONY: unit-test
unit-test: node_modules/.installed ## Runs all unit tests.
	# NOTE: Make sure the package builds.
	#npm run build
	npm run test

.PHONY: lint
lint: node_modules/.installed ## Runs eslint.
	@set -e;\
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then \
			exit_code=0; \
			while IFS="" read -r p && [ -n "$$p" ]; do \
				file=$$(echo "$$p" | jq -c '.filePath // empty' | tr -d '"'); \
				file=$$(realpath --relative-to "$(REPO_ROOT)" "$$file"); \
				while IFS="" read -r m && [ -n "$$m" ]; do \
					severity=$$(echo "$$m" | jq -c '.severity // empty' | tr -d '"'); \
					line=$$(echo "$$m" | jq -c '.line // empty' | tr -d '"'); \
					endline=$$(echo "$$m" | jq -c '.endLine // empty' | tr -d '"'); \
					col=$$(echo "$$m" | jq -c '.column // empty' | tr -d '"'); \
					endcol=$$(echo "$$m" | jq -c '.endColumn // empty' | tr -d '"'); \
					message=$$(echo "$$m" | jq -c '.message // empty' | tr -d '"'); \
					exit_code=1; \
					case $$severity in \
					"1") \
						echo "::warning file=$${file},line=$${line},endLine=$${endline},col=$${col},endColumn=$${endcol}::$${message}"; \
						;; \
					"2") \
						echo "::error file=$${file},line=$${line},endLine=$${endline},col=$${col},endColumn=$${endcol}::$${message}"; \
						;; \
					esac; \
				done <<<$$(echo "$$p" | jq -c '.messages[]'); \
			done <<<$$(./node_modules/.bin/eslint --max-warnings 0 -f json src/**/*.ts | jq -c '.[]'); \
			exit "$${exit_code}"; \
		else \
			./node_modules/.bin/eslint --max-warnings 0 src/**/*.ts; \
		fi

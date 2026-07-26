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

include include.mk

BENCHTIME ?= 1s
TESTCOUNT ?= 1

IMAGE_NAME ?= ghcr.io/ianlewis/todos

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
build-npm: $(REPO_ROOT)/node_modules/.installed build-all ## Build npm package tarball.
	@echo "Building npm package tarballs..."
	echo "NOTE: npm tarballs are for local use only and is not used in releases."
	npm pack
	cp todos-linux-amd64 packages/todos-linux-amd64/todos
	(cd packages/todos-linux-amd64 && npm pack)
	cp todos-linux-arm64 packages/todos-linux-arm64/todos
	(cd packages/todos-linux-arm64 && npm pack)
	cp todos-darwin-amd64 packages/todos-darwin-amd64/todos
	(cd packages/todos-darwin-amd64 && npm pack)
	cp todos-darwin-arm64 packages/todos-darwin-arm64/todos
	(cd packages/todos-darwin-arm64 && npm pack)
	cp todos-windows-amd64 packages/todos-windows-amd64/todos.exe
	(cd packages/todos-windows-amd64 && npm pack)
	cp todos-windows-arm64 packages/todos-windows-arm64/todos.exe
	(cd packages/todos-windows-arm64 && npm pack)

todos-with-pprof-%: $(GO_SOURCE_FILES)
	@echo "Building todos with profiling for $*..."
	echo "NOTE: $@ is for local use only and is not used in releases."
	go mod vendor
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
	@echo "Building todos for $*..."
	echo "NOTE: $@ is for local use only and is not used in releases."
	go mod vendor
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
	@echo "Building Docker image $(IMAGE_NAME)..."
	echo "NOTE: The Docker image is for local use only and is not used in releases."
	docker build \
		-t "$(IMAGE_NAME)" \
		.

## Testing
#####################################################################

.PHONY: test
test: lint unit-test ## Run all linters and tests.

.PHONY: unit-test
unit-test: ## Runs all unit tests.
	@echo "Running unit tests..."
	go mod vendor
	extraargs=""
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then
		extraargs="-v"
	fi
	go test \
		$${extraargs} \
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
	@echo "Running Go benchmarks..."
	go mod vendor
	extraargs=""
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then
		extraargs="-v"
	fi
	go test \
		$${extraargs} \
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
	@echo "Formatting Go files..."
	files=$$(
		git ls-files --deduplicate \
			'*.go'
	)
	if [ "$${files}" == "" ]; then
		exit 0
	fi
	gofumpt -l -w $${files}
	goimports -l -w $${files}
	gci write \
		--skip-generated \
		--skip-vendor \
		-s standard \
		-s default \
		-s localmodule \
		$${files}

.PHONY: json-format
json-format: $(REPO_ROOT)/node_modules/.installed ## Format JSON files.
	@echo "Formatting JSON files..."
	loglevel="log"
	if [ -n "$(DEBUG_LOGGING)" ]; then
		loglevel="debug"
	fi
	files=$$(
		git ls-files --deduplicate \
			'*.json' \
			'*.json5' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done
	)
	if [ "$${files}" == "" ]; then
		exit 0
	fi
	./node_modules/.bin/prettier \
		--log-level "$${loglevel}" \
		--no-error-on-unmatched-pattern \
		--write \
		$${files}

.PHONY: license-headers
license-headers: ## Update license headers.
	@echo "Updating license headers..."
	files=$$(
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
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done
	)
	name=$$(git config user.name)
	if [ "$${name}" == "" ]; then
		>&2 echo "git user.name is required."
		>&2 echo "Set it up using:"
		>&2 echo "git config user.name \"John Doe\""
		exit 1
	fi
	for filename in $${files}; do
		if ! ( head "$${filename}" | $(GREP) -iL "Copyright" > /dev/null ); then
			./third_party/mbrukman/autogen/autogen.sh \
				--in-place \
				--no-code \
				--no-tlc \
				--copyright "$${name}" \
				--license apache \
				"$${filename}"
		fi
	done

.PHONY: md-format
md-format: $(REPO_ROOT)/node_modules/.installed ## Format Markdown files.
	@echo "Formatting Markdown files..."
	loglevel="log"
	if [ -n "$(DEBUG_LOGGING)" ]; then
		loglevel="debug"
	fi
	files=$$(
		git ls-files --deduplicate \
			'*.md' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done
	)
	if [ "$${files}" == "" ]; then
		exit 0
	fi
	# NOTE: prettier uses .editorconfig for tab-width. \
	./node_modules/.bin/prettier \
		--log-level "$${loglevel}" \
		--no-error-on-unmatched-pattern \
		--write \
		$${files}

.PHONY: yaml-format
yaml-format: $(REPO_ROOT)/node_modules/.installed ## Format YAML files.
	@echo "Formatting YAML files..."
	loglevel="log"
	if [ -n "$(DEBUG_LOGGING)" ]; then
		loglevel="debug"
	fi
	files=$$(
		git ls-files --deduplicate \
			'*.yml' \
			'*.yaml' \
	)
	if [ "$${files}" == "" ]; then
		exit 0
	fi
	./node_modules/.bin/prettier \
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
	@echo "Running actionlint..."
	files=$$(
		git ls-files --deduplicate \
			'.github/workflows/*.yml' \
			'.github/workflows/*.yaml' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done
	)
	if [ "$${files}" == "" ]; then
		exit 0
	fi
	actionlint \
		-ignore 'SC2016:' \
		$${files}

.PHONY: checkmake
checkmake: $(AQUA_ROOT_DIR)/.installed ## Runs the checkmake linter.
	@echo "Running checkmake..."
	files=$$(
		git ls-files --deduplicate \
			'Makefile' \
			'GNUmakefile' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done \
	)
	if [ "$${files}" == "" ]; then
		exit 0
	fi
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then
		checkmake \
			--format '::error file={{.FileName}},line={{.LineNumber}}::{{.Rule}}: {{.Violation}}' \
			$${files}
	else
		checkmake $${files}
	fi

COMMITLINT_FROM_REF ?=
COMMITLINT_TO_REF ?=

.PHONY: commitlint
commitlint: $(REPO_ROOT)/node_modules/.installed ## Run commitlint linter.
	@echo "Running commitlint..."
	commitlint_from=$(COMMITLINT_FROM_REF)
	commitlint_to=$(COMMITLINT_TO_REF)
	if [ "$${commitlint_from}" == "" ]; then
		# Try to get the default branch without hitting the remote server
		if git symbolic-ref --short refs/remotes/origin/HEAD >/dev/null 2>&1; then
			commitlint_from=$$(git symbolic-ref --short refs/remotes/origin/HEAD)
		elif git show-ref refs/remotes/origin/master >/dev/null 2>&1; then
			commitlint_from="origin/master"
		else
			commitlint_from="origin/main"
		fi
	fi
	if [ "$${commitlint_to}" == "" ]; then
		# If upstream of HEAD is on the commitlint_from branch, then we will
		# lint the last commit by default.
		current_branch=$$(git rev-parse --abbrev-ref @{u})
		if [ "$${commitlint_from}" == "$${current_branch}" ]; then
			commitlint_from="HEAD~1"
		fi
		commitlint_to="HEAD"
	fi
	./node_modules/.bin/commitlint \
		--from "$${commitlint_from}" \
		--to "$${commitlint_to}" \
		--verbose \
		--strict

.PHONY: docs-check
docs-check: ## Check that generated documentation is up to date.
	@echo "Checking generated documentation..."
	if [ -n "$$(git diff)" ]; then
		>&2 echo "The working directory is dirty. Please commit, stage, or stash changes and try again."
		exit 1
	fi
	$(MAKE) docs
	exit_code=0
	if [ -n "$$(git diff)" ]; then
		>&2 echo "Some files need to be generated. Please run '$(MAKE) docs' and try again."
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then
			echo "::group::git diff"
		fi
		git --no-pager diff
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then
			echo "::endgroup::"
		fi
		exit_code=1
	fi
	git restore .
	exit "$${exit_code}"

.PHONY: eslint
eslint: $(REPO_ROOT)/node_modules/.installed ## Runs eslint.
	@echo "Running eslint..."
	extraargs=""
	if [ -n "$(DEBUG_LOGGING)" ]; then
		extraargs="--debug"
	fi
	files=$$(
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
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done
	)
	if [ "$${files}" == "" ]; then
		exit 0
	fi
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then
		exit_code=0
		while IFS="" read -r p && [ -n "$${p}" ]; do
			file=$$(echo "$${p}" | jq -c '.filePath // empty' | tr -d '"')
			while IFS="" read -r m && [ -n "$${m}" ]; do
				severity=$$(echo "$${m}" | jq -c '.severity // empty' | tr -d '"')
				line=$$(echo "$${m}" | jq -c '.line // empty' | tr -d '"')
				endline=$$(echo "$${m}" | jq -c '.endLine // empty' | tr -d '"')
				col=$$(echo "$${m}" | jq -c '.column // empty' | tr -d '"'); \
				endcol=$$(echo "$${m}" | jq -c '.endColumn // empty' | tr -d '"')
				message=$$(echo "$${m}" | jq -c '.message // empty' | tr -d '"')
				exit_code=1
				case $${severity} in
				"1")
					echo "::warning file=$${file},line=$${line},endLine=$${endline},col=$${col},endColumn=$${endcol}::$${message}"
					;;
				"2")
					echo "::error file=$${file},line=$${line},endLine=$${endline},col=$${col},endColumn=$${endcol}::$${message}"
					;;
				esac
			done <<<$$(echo "$${p}" | jq -c '.messages[]')
		done <<<$$($(REPO_ROOT)/node_modules/.bin/eslint \
			--max-warnings 0 \
			--format json \
			$${extraargs} \
			$${files} | jq -c '.[]')
		exit "$${exit_code}"
	else
		$(REPO_ROOT)/node_modules/.bin/eslint \
			--max-warnings 0 \
			$${extraargs} \
			$${files}
	fi

.PHONY: fixme
fixme: $(AQUA_ROOT_DIR)/.installed ## Check for outstanding FIXMEs.
	@echo "Checking for outstanding FIXMEs..."
	output="default"
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then
		output="github"
	fi
	# NOTE: todos does not use `git ls-files` because many files might be
	# 		unsupported and generate an error if passed directly on the
	# 		command line.
	todos \
		--output "$${output}" \
		--todo-types="FIXME,Fixme,fixme,BUG,Bug,bug,XXX,COMBAK"

.PHONY: format-check
format-check: ## Check that files are properly formatted.
	@echo "Checking that files are properly formatted..."
	if [ -n "$$(git diff)" ]; then
		>&2 echo "The working directory is dirty. Please commit, stage, or stash changes and try again."
		exit 1
	fi
	$(MAKE) format
	exit_code=0
	if [ -n "$$(git diff)" ]; then
		>&2 echo "Some files need to be formatted. Please run '$(MAKE) format' and try again."
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then
			echo "::group::git diff"
		fi
		git --no-pager diff
		if [ "$(OUTPUT_FORMAT)" == "github" ]; then
			echo "::endgroup::"
		fi
		exit_code=1
	fi
	git restore .
	exit "$${exit_code}"

.PHONY: golangci-lint
golangci-lint: $(AQUA_ROOT_DIR)/.installed ## Runs the golangci-lint linter.
	@echo "Running golangci-lint..."
	golangci-lint run -c .golangci.yml ./...

..PHONY: hadolint
hadolint: $(AQUA_ROOT_DIR)/.installed ## Runs the hadolint linter.
	@echo "Running hadolint..."
	files=$$(
		git ls-files --deduplicate \
			'[Dd]ockerfile' \
			'[Cc]ontainerfile' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done
	)
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then
		hadolint -f checkstyle $${files}
	else
		hadolint $${files}
	fi

.PHONY: markdownlint
markdownlint: $(REPO_ROOT)/node_modules/.installed $(AQUA_ROOT_DIR)/.installed ## Runs the markdownlint linter.
	@echo "Running markdownlint..."
	# NOTE: Issue and PR templates are handled specially so we can disable
	# MD041/first-line-heading/first-line-h1 without adding an ugly html comment
	# at the top of the file.
	files=$$( \
		git ls-files --deduplicate \
			'*.md' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done
	)
	if [ "$${files}" == "" ]; then
		exit 0
	fi
	./node_modules/.bin/markdownlint-cli2 $${files}

.PHONY: renovate-config-validator
renovate-config-validator: $(REPO_ROOT)/node_modules/.installed ## Validate Renovate configuration.
	@echo "Validating Renovate configuration..."
	./node_modules/.bin/renovate-config-validator \
		--strict

.PHONY: textlint
textlint: $(REPO_ROOT)/node_modules/.installed $(AQUA_ROOT_DIR)/.installed ## Runs the textlint linter.
	@echo "Running textlint..."
	files=$$(
		git ls-files --deduplicate \
			'*.md' \
			'*.txt' \
			':!:requirements*.txt' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done
	)
	if [ "$${files}" == "" ]; then
		exit 0
	fi
	./node_modules/.bin/textlint $${files}

.PHONY: yamllint
yamllint: $(REPO_ROOT)/.venv/.installed ## Runs the yamllint linter.
	@echo "Running yamllint..."
	files=$$(
		git ls-files --deduplicate \
			'*.yml' \
			'*.yaml' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done
	)
	if [ "$${files}" == "" ]; then
		exit 0
	fi
	format="standard"
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then
		format="github"
	fi
	./.venv/bin/yamllint \
		--strict \
		--format "$${format}" \
		$${files}

.PHONY: zizmor
zizmor: $(REPO_ROOT)/.venv/.installed ## Runs the zizmor linter.
	@echo "Running zizmor..."
	# NOTE: On GitHub actions this outputs SARIF format to zizmor.sarif.json
	#       in addition to outputting errors to the terminal.
	files=$$(
		git ls-files --deduplicate \
			'.github/workflows/*.yml' \
			'.github/workflows/*.yaml' \
			| while IFS='' read -r f; do [ -f "$${f}" ] && echo "$${f}" || true; done
	)
	if [ "$${files}" == "" ]; then
		exit 0
	fi
	format="plain"
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then
		./.venv/bin/zizmor \
			--quiet \
			--pedantic \
			--format sarif \
			$${files} > zizmor.sarif.json
		format="github"
	fi
	./.venv/bin/zizmor \
		--quiet \
		--pedantic \
		--format "$${format}" \
		$${files}

## Documentation
#####################################################################

.PHONY: docs
docs: SUPPORTED_LANGUAGES.md ## Generate all documentation.

SUPPORTED_LANGUAGES.md: $(REPO_ROOT)/node_modules/.installed internal/scanner/languages.go ## Supported languages documentation.
	@# bash \
	go mod vendor; \
	go run \
		-mod=vendor \
		./internal/cmd/genlangdocs | \
			./node_modules/.bin/prettier \
				--parser markdown > $@

## Maintenance
#####################################################################

.PHONY: update-lockfiles
update-lockfiles: $(REPO_ROOT)/.aqua-checksums.json $(REPO_ROOT)/package-lock.json $(REPO_ROOT)/uv.lock ## Update lockfiles.

.PHONY: todos
todos: $(AQUA_ROOT_DIR)/.installed ## Print outstanding TODOs.
	@echo "Checking for outstanding TODOs..."
	output="default"
	if [ "$(OUTPUT_FORMAT)" == "github" ]; then
		output="github"
	fi
	# NOTE: todos does not use `git ls-files` because many files might be
	# 		unsupported and generate an error if passed directly on the command
	# 		line.
	todos \
		--output "$${output}" \
		--todo-types="TODO,Todo,todo,FIXME,Fixme,fixme,BUG,Bug,bug,XXX,COMBAK"

.PHONY: clean-node-modules
clean-node-modules:
	@echo "Cleaning up node_modules..."
	$(RM) -r node_modules

.PHONY: clean
clean: clean-node-modules ## Delete temporary files.
	@echo "Cleaning up temporary files..."
	$(RM) -r .bin
	$(RM) -r $(AQUA_ROOT_DIR)
	$(RM) -r .venv
	$(RM) -r .uv
	$(RM) *.sarif.json
	$(RM) -r vendor
	$(RM) coverage.out
	$(RM) todos
	$(RM) todos-*
	$(RM) ianlewis-todos-*.tgz
	$(RM) packages/todos-linux-amd64/todos
	$(RM) packages/todos-linux-amd64/*.tgz
	$(RM) packages/todos-linux-arm64/todos
	$(RM) packages/todos-linux-arm64/*.tgz
	$(RM) packages/todos-darwin-amd64/todos
	$(RM) packages/todos-darwin-amd64/*.tgz
	$(RM) packages/todos-darwin-arm64/todos
	$(RM) packages/todos-darwin-arm64/*.tgz
	$(RM) packages/todos-windows-amd64/todos.exe
	$(RM) packages/todos-windows-amd64/*.tgz
	$(RM) packages/todos-windows-arm64/todos.exe
	$(RM) packages/todos-windows-arm64/*.tgz

# Copyright 2026 Ian Lewis
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

# include.mk sets up the environment for the Makefile. It is included at the top
# of the Makefile. It sets up the shell, flags, and other variables that are
# used throughout the Makefile. This file can be included in other Makefiles
# (e.g. in subdirectories) to provide a consistent environment.

# Print a warning if undefined variables are used in the Makefile.
MAKEFLAGS += --warn-undefined-variables
# Disable built-in rules to avoid conflicts with custom rules.
MAKEFLAGS += --no-builtin-rules

# Set the initial shell so we can determine extra options.
SHELL := bash
# TODO(https://github.com/checkmake/checkmake/issues/263): use := assignment
# NOTE: We use recursive assignment to avoid triggering a bug in checkmake.
.SHELLFLAGS = -ueo pipefail -c

# DEBUG_LOGGING is set to true when debug logging is enabled. It will be
# automatically set to true when running in GitHub Actions with debug logging
# enabled. It can also be set manually to enable debug logging when running
# locally.
DEBUG_LOGGING ?=

# The GITHUB_ACTIONS environment variable is set to 'true' when running in
# GitHub Actions.
# https://docs.github.com/en/actions/reference/workflows-and-actions/variables
GITHUB_ACTIONS ?=

# The RUNNER_DEBUG environment variable is set to '1' when debug mode is
# enabled.
# https://docs.github.com/en/actions/reference/workflows-and-actions/variables
RUNNER_DEBUG ?=

# GitHub Actions debug logging environment variables.
# https://docs.github.com/en/actions/how-tos/monitor-workflows/enable-debug-logging
ACTIONS_RUNNER_DEBUG ?=
ACTIONS_STEP_DEBUG ?=

ifeq ($(DEBUG_LOGGING),)
  ifeq ($(GITHUB_ACTIONS),true)
    ifneq ($(RUNNER_DEBUG),)
      DEBUG_LOGGING := true
    else
      ifeq ($(ACTIONS_RUNNER_DEBUG),true)
        DEBUG_LOGGING := true
      else
        ifeq ($(ACTIONS_STEP_DEBUG),true)
          DEBUG_LOGGING := true
        endif
      endif
    endif
  endif
endif

ifeq ($(DEBUG_LOGGING),true)
  .SHELLFLAGS := -x $(.SHELLFLAGS)
endif

# .ONESHELL executes all commands in a recipe in a single shell instance. This
# allows us to use shell variables and functions across multiple lines in a
# recipe.
.ONESHELL:

# .DELETE_ON_ERROR deletes the recipe target files if the recipe fails. This
# makes sure that re-running the recipe will not use a partially created target
# file.
.DELETE_ON_ERROR:

uname_s := $(shell uname -s)
uname_m := $(shell uname -m)
arch.x86_64 := amd64
arch.aarch64 := arm64
arch.arm64 := arm64
arch := $(arch.$(uname_m))
kernel.Linux := linux
kernel.Darwin := darwin
kernel := $(kernel.$(uname_s))

OUTPUT_FORMAT ?= $(shell if [ "$(GITHUB_ACTIONS)" == "true" ]; then echo "github"; else echo ""; fi)
REPO_ROOT := $(shell git rev-parse --show-toplevel 2>/dev/null)
REPO_NAME := $(shell basename "$(REPO_ROOT)")
MAKEFILE_ROOT := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
MAKEFILE_NAME := $(shell basename "$(MAKEFILE_ROOT)")

# renovate: datasource=github-releases depName=aquaproj/aqua versioning=loose
AQUA_VERSION ?= v2.60.1
export AQUA_ROOT_DIR = $(MAKEFILE_ROOT)/.aqua

# Ensure that aqua and aqua installed tools are in the PATH.
export PATH := $(AQUA_ROOT_DIR)/bin:$(PATH)

# We want GNU versions of tools so prefer them if present.
GREP := $(shell command -v ggrep 2>/dev/null || command -v grep 2>/dev/null)
AWK := $(shell command -v gawk 2>/dev/null || command -v awk 2>/dev/null)
MKTEMP := $(shell command -v gmktemp 2>/dev/null || command -v mktemp 2>/dev/null)

# The help command prints targets in groups. Help documentation in the Makefile
# uses comments with double hash marks (##). Documentation is printed by the
# help target in the order it appears in the Makefile.
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
	@echo "$(MAKEFILE_NAME) Makefile"
	echo "Usage: $(MAKE) [COMMAND]"
	echo ""
	normal="";
	cyan=""
	if command -v tput >/dev/null 2>&1; then
		if [ -t 1 ]; then
			normal=$$(tput sgr0)
			cyan=$$(tput setaf 6)
		fi
	fi
	$(GREP) --no-filename -E '^([/a-z.A-Z0-9_%-]+:.*?|)##' $(MAKEFILE_LIST) | \
		$(AWK) \
			--assign=normal="$${normal}" \
			--assign=cyan="$${cyan}" \
			'BEGIN {FS = "(:.*?|)## ?"}; {
				if (length($$1) > 0) {
					printf("  " cyan "%-25s" normal " %s\n", $$1, $$2)
				} else {
					if (length($$2) > 0) {
						printf("%s\n", $$2)
					}
				}
			}'

# Node.js setup
#####################################################################

$(REPO_ROOT)/package-lock.json: $(REPO_ROOT)/package.json $(AQUA_ROOT_DIR)/.installed
	@echo "Updating Node.js dependencies..."
	loglevel="notice"
	if [ -n "$(DEBUG_LOGGING)" ]; then
		loglevel="verbose"
	fi
	# NOTE: npm install will happily ignore the fact that integrity hashes are
	# missing in the package-lock.json. We need to check for missing integrity
	# fields ourselves. If any are missing, then we need to regenerate the
	# package-lock.json from scratch.
	nointegrity=""
	noresolved=""
	if [ -f "$@" ]; then
		nointegrity=$$(jq '.packages | del(."") | .[] | select(has("integrity") | not)' < $@)
		noresolved=$$(jq '.packages | del(."") | .[] | select(has("resolved") | not)' < $@)
	fi
	if [ ! -f "$@" ] || [ -n "$${nointegrity}" ] || [ -n "$${noresolved}" ]; then
		# NOTE: package-lock.json is removed to ensure that npm includes the
		# integrity field. npm install will not restore this field if
		# missing in an existing package-lock.json file.
		rm -f $@
		# NOTE: We clean the node_modules directory to ensure that npm install
		#       will not desync between the package.json, package-lock.json
		#       and the node_modules directory. \
		$(MAKE) clean-node-modules
		npm --loglevel="$${loglevel}" install \
			--no-audit \
			--no-fund
	else
		npm --loglevel="$${loglevel}" install \
			--package-lock-only \
			--no-audit \
			--no-fund
	fi

$(REPO_ROOT)/node_modules/.installed: $(REPO_ROOT)/package.json
	@echo "Installing Node.js dependencies..."
	loglevel="silent"
	if [ -n "$(DEBUG_LOGGING)" ]; then
		loglevel="verbose"
	fi
	npm --loglevel="$${loglevel}" clean-install
	npm --loglevel="$${loglevel}" audit signatures
	touch $@

# Python setup
#####################################################################

$(REPO_ROOT)/.uv/venv/bin/activate:
	@echo "Creating Python virtual environment..."
	mkdir -p $(REPO_ROOT)/.uv
	python -m venv $(REPO_ROOT)/.uv/venv
	touch $@

$(REPO_ROOT)/.uv/.installed: $(REPO_ROOT)/requirements-dev.txt $(REPO_ROOT)/.uv/venv/bin/activate
	@echo "Installing Python dependencies..."
	$(REPO_ROOT)/.uv/venv/bin/pip install -r $< --require-hashes
	touch $@

$(REPO_ROOT)/uv.lock: $(REPO_ROOT)/pyproject.toml $(REPO_ROOT)/.uv/.installed
	@echo "Updating Python dependencies..."
	$(REPO_ROOT)/.uv/venv/bin/uv lock
	touch $@

$(REPO_ROOT)/.venv/.installed: $(REPO_ROOT)/pyproject.toml $(REPO_ROOT)/.uv/.installed
	@echo "Installing Python dependencies..."
	$(REPO_ROOT)/.uv/venv/bin/uv sync --locked
	touch $@

# Aqua setup
#####################################################################

$(AQUA_ROOT_DIR)/.$(AQUA_VERSION).installed:
	@echo "Installing aqua $(AQUA_VERSION)..."
	$(REPO_ROOT)/third_party/aquaproj/aqua-installer/aqua-installer -v "$(AQUA_VERSION)"
	touch $@

$(REPO_ROOT)/.aqua-checksums.json: $(REPO_ROOT)/.aqua.yaml $(AQUA_ROOT_DIR)/.$(AQUA_VERSION).installed
	@echo "Updating aqua checksums..."
	loglevel="info"
	if [ -n "$(DEBUG_LOGGING)" ]; then
		loglevel="debug"
	fi
	$(AQUA_ROOT_DIR)/bin/aqua \
		--config "$(REPO_ROOT)/.aqua.yaml" \
		--log-level "$${loglevel}" \
		update-checksum \
		--prune

$(AQUA_ROOT_DIR)/.installed: $(REPO_ROOT)/.aqua.yaml $(AQUA_ROOT_DIR)/.$(AQUA_VERSION).installed
	@echo "Installing aqua tools..."
	loglevel="info"
	if [ -n "$(DEBUG_LOGGING)" ]; then
		loglevel="debug"
	fi
	$(AQUA_ROOT_DIR)/bin/aqua \
		--config "$(REPO_ROOT)/.aqua.yaml" \
		--log-level "$${loglevel}" \
		install
	touch $@


# ========= Env loading  =========
ENV_FILES := .env
-include $(ENV_FILES)
export

# ========= Variables =========
GO_BUILD_ENVS     ?= CGO_ENABLED=1
INTERNAL_PATH     := github.com/ad9311/ninete/internal
SHELL             := /bin/bash
SHELL_FILES       := $(wildcard scripts/*.sh)
pkg               ?= ./...
func              ?=

# ========= Version =========
# Build identity, injected into internal/prog with -X. Derived from git so it
# cannot go stale; every value falls back to a literal when git is unavailable,
# because a build must never depend on the repository being readable.
VERSION           := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT            := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME        := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
VERSION_LDFLAGS   := -X $(INTERNAL_PATH)/prog.Version=$(VERSION) \
                     -X $(INTERNAL_PATH)/prog.Commit=$(COMMIT) \
                     -X $(INTERNAL_PATH)/prog.BuildTime=$(BUILD_TIME)

# ========= Phony =========
.PHONY: help dev build build-final deps lint lint-fix lint-sh build-static-js version snapshot

# ========= App / Dev =========
build: ## Build the application binary
	@echo "Building binary..."
	@mkdir -p ./build
	@mkdir -p ./data/db/dev
	@$(GO_BUILD_ENVS) go build -ldflags "$(VERSION_LDFLAGS)" -o ./build/dev ./cmd/ninete/main.go

dev: build-static-js build ## Run the app in development mode
	@echo "Starting application..."
	@ENV=development ./build/dev

build-migrate: ## Build the migrate binary
	@echo "Building migrate binary..."
	@mkdir -p ./build
	@mkdir -p ./data/db/dev
	$(GO_BUILD_ENVS) go build -ldflags "$(VERSION_LDFLAGS)" -o ./build/migrate ./cmd/migrate/main.go

migrate: build-migrate ## Run all migrations up
	@echo "Running migrations..."
	ENV=development ./build/migrate up

migrate-down: build-migrate ## Run all migrations up
	@echo "Running one migration down..."
	ENV=development ./build/migrate down

migrate-create: build-migrate ## Run all migrations up
	@echo "Creating migration file..."
	ENV=development ./build/migrate create $(name)

migrate-status: build-migrate ## Run all migrations up
	ENV=development ./build/migrate status

seed: build-migrate ## Seed the database
	ENV=development ./build/migrate seed

stamp: build-migrate ## Claim the database for the current ENV
	ENV=development ./build/migrate stamp

snapshot: build-migrate ## Write a snapshot of the development database
	ENV=development ./build/migrate snapshot

build-task: ## Build the task binary
	@echo "Building task binary..."
	@mkdir -p ./build
	@mkdir -p ./data/db/dev
	$(GO_BUILD_ENVS) go build -ldflags "$(VERSION_LDFLAGS)" -o ./build/task ./cmd/task/main.go

task: build-task ## Run a task
	@echo "Running $(name) task..."
	ENV=development ./build/task $(name)

clean: ## Removes compiled binaries
	@echo "Removing binaries..."
	@rm -rf ./build/*

clean-db: ## Removes dev database file
	@echo "Removing development database..."
	@rm -rf ./data/db/dev/*

clean-test-db: ## Removes test database files
	@echo "Removing test databases..."
	@rm -rf ./data/db/test/*

clean-test-cache: ## Cleans go test cache
	@echo "Removing go test cache..."
	@go clean -testcache

clean-full: clean clean-db clean-test-db clean-test-cache ## Runs `clean`, `clean-db`, `clean-test-db` and `clean-test-cache`
	@echo "Full clean done!"

deps: ## Install and tidy dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

build-static-js: ## Build the frontend entrypoints into web/static/js/build with bun
	@echo "Building static JS bundle..."
	bun run web/build.ts

# ========= Tests ===========
# build-static-js is a dependency because internal/serve asserts that /static/*
# serves the bundle, and the bundle is git-ignored.
test: build-static-js build clean-test-db ## Runs the tests
	@echo "Running tests..."
	@mkdir -p ./data/db/test
	ENV=test go test $(if $(func),-run $(func),) $(pkg)

test-verbose: build-static-js build clean-test-db ## Runs the tests in verbose mode
	@echo "Running tests in verbose mode"
	@mkdir -p ./data/db/test
	ENV=test go test -v $(if $(func),-run $(func),) $(pkg)

# Both zones, not just the configured default: a calendar date formatted with
# local getters still reads correctly east of UTC and only breaks west of it, so
# the Los Angeles run is the one that catches most of §3.6. Each is well under a
# second. CI runs them as separate jobs so a failure names its zone.
test-js: ## Runs the JS/Svelte tests in both configured zones
	@echo "Running JS tests (Pacific/Auckland)..."
	TEST_TZ=Pacific/Auckland bun run test:js
	@echo "Running JS tests (America/Los_Angeles)..."
	TEST_TZ=America/Los_Angeles bun run test:js

# ========= Linting =========
lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	golangci-lint run
	@$(MAKE) --no-print-directory lint-sh

# The bun steps run before golangci on purpose: make stops at the first failing
# recipe line, so a Go lint failure used to skip the static formatting entirely
# and leave web/ unformatted with no sign anything had been skipped.
lint-fix: ## Run golangci-lint with automatic fixes
	@echo "Running static formatter and linters with bun..."
	bun run format:static
	bun run lint:css
	bun run lint:js
	@echo "Running type checks..."
	bun run typecheck:ts
	bun run typecheck:svelte
	@echo "Running golangci-lint (with --fix)..."
	golangci-lint run --fix
	@$(MAKE) --no-print-directory lint-sh

# Runs last in lint/lint-fix, and skips itself when shellcheck is absent, so a
# missing optional tool cannot block the Go/CSS/JS formatting everyone runs. CI
# installs shellcheck in the step before calling this, so the skip cannot make
# the pipeline pass vacuously.
lint-sh: ## Run shellcheck over the deployment scripts
	@if ! command -v shellcheck >/dev/null 2>&1; then \
		echo "shellcheck not installed, skipping (brew install shellcheck)"; \
		exit 0; \
	fi; \
	echo "Running shellcheck..."; \
	shellcheck $(SHELL_FILES)

# ========= Version =========
version: ## Print the version this checkout would build
	@echo "$(VERSION) ($(COMMIT)) $(BUILD_TIME)"

# ========= Help =========
help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN { FS = ":.*##" } /^[a-zA-Z0-9_.-]+:.*##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

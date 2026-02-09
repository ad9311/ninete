
# ========= Env loading  =========
ENV_FILES := .env
-include $(ENV_FILES)
export

# ========= Variables =========
GO_BUILD_ENVS     ?= CGO_ENABLED=1
INTERNAL_PATH     := github.com/ad9311/ninete/internal
SHELL             := /bin/bash
pkg               ?= ./...
func              ?=

# ========= Phony =========
.PHONY: help dev build build-final deps lint lint-fix

# ========= App / Dev =========
build: ## Build the application binary
	@echo "Building binary..."
	@mkdir -p ./build
	@mkdir -p ./data/db/dev
	@$(GO_BUILD_ENVS) go build -o ./build/dev ./cmd/ninete/main.go

dev: build ## Run the app in development mode
	@echo "Starting application..."
	@ENV=development ./build/dev

build-migrate: ## Build the migrate binary
	@echo "Building migrate binary..."
	@mkdir -p ./build
	@mkdir -p ./data/db/dev
	$(GO_BUILD_ENVS) go build -o ./build/migrate ./cmd/migrate/main.go

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

build-task: ## Build the task binary
	@echo "Building task binary..."
	@mkdir -p ./build
	@mkdir -p ./data/db/dev
	$(GO_BUILD_ENVS) go build -o ./build/task ./cmd/task/main.go

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

# ========= Tests ===========
test: build clean-test-db ## Runs the tests
	@echo "Running tests..."
	@mkdir -p ./data/db/test
	ENV=test go test $(if $(func),-run $(func),) $(pkg)

test-verbose: build clean-test-db ## Runs the tests in verbose mode
	@echo "Running tests in verbose mode"
	@mkdir -p ./data/db/test
	ENV=test go test -v $(if $(func),-run $(func),) $(pkg)

# ========= Linting =========
lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	golangci-lint run

lint-fix: ## Run golangci-lint with automatic fixes
	@echo "Running golangci-lint (with --fix)..."
	golangci-lint run --fix

# ========= Help =========
help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN { FS = ":.*##" } /^[a-zA-Z0-9_.-]+:.*##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

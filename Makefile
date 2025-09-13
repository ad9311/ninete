
# ========= Env loading  =========
ENV_FILES := .env
-include $(ENV_FILES)
export

# ========= Variables =========
GO_BUILDENV       := CGO_ENABLED=1 GOOS=linux GOARCH=amd64
INTERNAL_PATH     := github.com/ad9311/ninete/internal
SHELL             := /bin/bash

# ========= Phony =========
.PHONY: help dev build build-final deps lint lint-fix

# ========= App / Dev =========
build: ## Build the application binary
	@echo "Building binary..."
	@mkdir -p ./build
	@$(GO_BUILDENV) go build -o ./build/dev ./cmd/ninete/main.go

dev: build ## Run the app in development mode
	@echo "Starting application..."
	@mkdir -p ./data/db/dev
	@ENV=development ./build/dev

build-migrate: ## Build the migrate binary
	@echo "Building migrate binary..."
	mkdir -p ./build
	$(GO_BUILDENV) go build -o ./build/dev_migrate ./cmd/migrate/main.go

migrate: build-migrate ## Run all migrations up
	@echo "Running migrations..."
	ENV=development ./build/dev_migrate up

migrate-down: build-migrate ## Run all migrations up
	@echo "Running one migration down..."
	ENV=development ./build/dev_migrate down

migrate-status: build-migrate ## Run all migrations up
	ENV=development ./build/dev_migrate status

clean: ## Removes compiled binaries
	@echo "Removing binaries..."
	@rm -rf ./build/*

clean-full: clean ## Runs `clean` and deletes all databases
	@echo "Removing databases..."
	@rm -rf ./data/db/*

deps: ## Install and tidy dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

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


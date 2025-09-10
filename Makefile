
# ========= Env loading  =========
ENV_FILES := .env
-include $(ENV_FILES)
export

# ========= Variables =========
MAIN_PATH         := cmd/ninete/main.go
BUILD_PATH        := build
BUILD_APP_NAME    := ninte

GO_BUILDENV       := CGO_ENABLED=1 GOOS=linux GOARCH=amd64

SHELL := /bin/bash

# ========= Phony =========
.PHONY: help dev build build-final deps lint lint-fix

# ========= App / Dev =========
dev: build ## Run the app in development mode
	@echo "Starting application..."
	ENV=development ./$(BUILD_PATH)/$(BUILD_APP_NAME) server

build: ## Build the application binary
	@echo "Building binary..."
	@mkdir -p $(BUILD_PATH)
	$(GO_BUILDENV) go build -o $(BUILD_PATH)/$(BUILD_APP_NAME) $(MAIN_PATH)

build-final: ## Build the application binary optimized for production
	@echo "Building optimized binary..."
	@mkdir -p $(BUILD_PATH)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -tags timetzdata -ldflags="-s -w" -o $(BUILD_PATH)/$(PRODUCTION_NAME) $(MAIN_PATH)

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


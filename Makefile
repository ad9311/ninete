# Makefile for go-api-base
# ========= Env loading (override-friendly) =========
# Load these if present, later files override earlier ones
ENV_FILES := .env .env.local .env.test
-include $(ENV_FILES)
export

# ========= Variables =========
MAIN_PATH         := cmd/main/main.go
BUILD_PATH        := build
BUILD_APP_NAME    := app
PRODUCTION_NAME   := ad9311app
DOCKER_COMPOSE    := docker-compose.yml
COMPOSE           := docker compose -f $(DOCKER_COMPOSE)

DB_SERVICE        := db_dev
TEST_DB_SERVICE   := db_test

# Default goal
.DEFAULT_GOAL := help

# Use bash for better control
SHELL := /bin/bash

# ========= Phony =========
.PHONY: help dev dev-test build clean deps test \
        db-start db-stop db-restart db-logs db-shell db-create db-drop \
        db-start-test db-stop-test db-shell-test db-create-test db-drop-test \
        migrate migrate-down \
        lint lint-fix db-reset print-env

# ========= Helpers =========
# Wait until Postgres is ready inside a running container
define wait_for_db
	@echo "Checking if database is ready (service: $(1))..."
	@for i in $$(seq 1 20); do \
		if $(COMPOSE) exec -T $(1) pg_isready -U $(DB_USER) > /dev/null 2>&1; then \
			echo "Database is ready."; \
			exit 0; \
		fi; \
		echo "Database not ready yet... ($$i/20)"; \
		sleep 1; \
	done; \
	echo "Database did not become ready in time." && exit 1
endef

print-env: ## Print key env vars loaded into Make
	@echo "Loaded env files (if present): $(ENV_FILES)"
	@printf "\n"
	@echo "DB_USER=$(DB_USER)"
	@echo "DB_NAME=$(DB_NAME)"
	@echo "DB_PORT=$(DB_PORT)"
	@printf "\n"
	@echo "TEST_DB_USER=$(TEST_DB_USER)"
	@echo "TEST_DB_NAME=$(TEST_DB_NAME)"
	@echo "TEST_DB_PORT=$(TEST_DB_PORT)"

# ========= App / Dev =========
dev: db-start build ## Run the app in development mode (starts dev database)
	@echo "Starting development server..."
	$(call wait_for_db,$(DB_SERVICE))
	ENV=development ./$(BUILD_PATH)/$(BUILD_APP_NAME) server

build: ## Build the application binary
	@echo "Building app..."
	@mkdir -p $(BUILD_PATH)
	go build -o $(BUILD_PATH)/$(BUILD_APP_NAME) $(MAIN_PATH)

build-final:
	@echo "Building optimized binary..."
	@mkdir -p $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -tags timetzdata -ldflags="-s -w" -o $(BUILD_PATH)/$(PRODUCTION_NAME) $(MAIN_PATH)

clean: db-stop test-db-stop ## Clean up build artifacts
	@echo "Cleaning up..."
	go clean -testcache
	rm -rf $(BUILD_PATH)/

deps: ## Install and tidy dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

test: test-db-start build ## Start test DB (if needed) and run tests
	@echo "Running tests..."
	$(call wait_for_db,$(TEST_DB_SERVICE))
	@status=0; \
  printf '\n'; \
	ENV=test go test -p 1 ./... || status=$$?; \
	exit $$status

test-force: ## Clean Go's test cache and then run all tests
	@echo "Cleaning test cache..."
	go clean -testcache; \
	$(MAKE) test

task: build ## Run a task by name
	@echo "Running task..."
	$(call wait_for_db,$(DB_SERVICE))
	ENV=maintenance ./$(BUILD_PATH)/$(BUILD_APP_NAME) task $(name)

print-tasks: build ## Prints all tasks
	./$(BUILD_PATH)/$(BUILD_APP_NAME) print-tasks

# ========= Linting =========
lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	golangci-lint run

lint-fix: ## Run golangci-lint with automatic fixes
	@echo "Running golangci-lint (with --fix)..."
	golangci-lint run --fix

# ========= Database: Dev =========
db-start: ## Start the development database container
	@echo "Starting development database..."
	$(COMPOSE) up -d $(DB_SERVICE)

db-stop: ## Stop the development database container
	@echo "Stopping development database..."
	$(COMPOSE) stop $(DB_SERVICE)

db-restart: db-stop db-start ## Restart the development database container

db-logs: ## Tail development database logs
	@echo "Showing database logs..."
	$(COMPOSE) logs -f $(DB_SERVICE)

db-shell: ## Open psql shell to development database
	@echo "Connecting to development database shell..."
	$(COMPOSE) exec $(DB_SERVICE) psql -U $(DB_USER) -d $(DB_NAME)

db-create: ## Create development database (idempotent)
	@echo "Creating development database..."
	$(COMPOSE) exec $(DB_SERVICE) createdb -U $(DB_USER) $(DB_NAME) || true

db-drop: ## Drop development database (if exists)
	@echo "Dropping development database..."
	$(COMPOSE) exec $(DB_SERVICE) dropdb -U $(DB_USER) --if-exists $(DB_NAME)

db-reset: db-start ## Reset dev DB: start container, drop DB, create DB
	@echo "Resetting development database (drop -> create)..."
	$(call wait_for_db,$(DB_SERVICE))
	$(MAKE) --no-print-directory db-drop
	$(MAKE) --no-print-directory db-create
	$(MAKE) --no-print-directory migrate

# ========= Database: Test =========
test-db-start: # Start the test database container
	@echo "Starting test database..."
	$(COMPOSE) up -d $(TEST_DB_SERVICE)

test-db-stop: ## Stop the test database container
	@echo "Stopping test database..."
	$(COMPOSE) stop $(TEST_DB_SERVICE)

test-db-shell: ## Open psql shell to test database
	@echo "Connecting to test database shell..."
	$(COMPOSE) exec $(TEST_DB_SERVICE) psql -U $(TEST_DB_USER) -d $(TEST_DB_NAME)

test-db-create: ## Create test database (idempotent)
	@echo "Creating test database..."
	$(COMPOSE) exec $(TEST_DB_SERVICE) createdb -U $(TEST_DB_USER) $(TEST_DB_NAME) || true

test-db-drop: ## Drop test database (if exists)
	@echo "Dropping test database..."
	$(COMPOSE) exec $(TEST_DB_SERVICE) dropdb -U $(TEST_DB_USER) --if-exists $(TEST_DB_NAME)

test-db-reset: db-start-test ## Reset test DB: drop -> create -> migrate
	$(call wait_for_db,$(TEST_DB_SERVICE))
	$(MAKE) --no-print-directory test-db-drop
	$(MAKE) --no-print-directory test-db-create
	$(MAKE) --no-print-directory test-migrate

# ========= Migrations =========
migrate: build ## Run all pending migrations against development DB
	@echo "Running migrations..."
	$(call wait_for_db,$(DB_SERVICE))
	ENV=development ./$(BUILD_PATH)/$(BUILD_APP_NAME) migrate

migrate-down: build ## Roll back one migration against development DB
	@echo "Running one migration down..."
	$(call wait_for_db,$(DB_SERVICE))
	ENV=development ./$(BUILD_PATH)/$(BUILD_APP_NAME) migrate-down

status: build ## Show migration status for development DB
	@echo "Migrations status..."
	$(call wait_for_db,$(DB_SERVICE))
	ENV=development ./$(BUILD_PATH)/$(BUILD_APP_NAME) status

test-migrate: build ## Run all pending migrations against test DB
	@echo "Running migrations..."
	$(call wait_for_db,$(TEST_DB_SERVICE))
	ENV=test ./$(BUILD_PATH)/$(BUILD_APP_NAME) migrate

test-migrate-down: build ## Roll back one migration against test DB
	@echo "Running one migration down..."
	$(call wait_for_db,$(TEST_DB_SERVICE))
	ENV=test ./$(BUILD_PATH)/$(BUILD_APP_NAME) migrate-down

test-status: build ## Show migration status for test DB
	@echo "Migrations status..."
	$(call wait_for_db,$(DB_SERVICE))
	ENV=test ./$(BUILD_PATH)/$(BUILD_APP_NAME) status

# ========= Help =========

help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN { FS = ":.*##" } /^[a-zA-Z0-9_.-]+:.*##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)


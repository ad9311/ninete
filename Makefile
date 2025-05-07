# Makefile

# Development
dev:
	bun run dev

# Build
build:
	bun run build

# Preview
preview:
	bun run preview

# Prepare
prepare:
	bun run prepare

# Check
check:
	bun run check

check-watch:
	bun run check:watch

# Format
format:
	bun run format

format-all:
	bun run format:all

# Lint
lint:
	bun run lint

# Test
test-unit:
	bun run test:unit

test:
	bun run test

test-server:
	bun run test:server

# Database
db-start:
	bun run db:start

db-push:
	bun run db:push

db-migrate:
	bun run db:migrate

db-studio:
	bun run db:studio

# Generates a new database migration with a specified name
# Usage: make db-generate-migration name=your_migration_name
db-generate-migration:
ifndef name
	@echo "Error: 'name' variable is not set."
	@echo "Usage: make db-generate-migration name=<migration_name>"
	@exit 1
endif
	bunx drizzle-kit generate --name=$(name)

# Machine Translate
machine-translate:
	bun run machine-translate

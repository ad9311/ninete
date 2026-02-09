# NINETE Architecture Guide

## Purpose
This document explains how the project is structured and how packages interact at runtime.

## Runtime Flow
1. `cmd/ninete/main.go` loads app config with `prog.Load()`.
2. It opens SQLite through `db.Open()`.
3. It creates data access with `repo.New(app, sqlDB)`.
4. It creates business logic with `logic.New(app, queries)`.
5. It creates HTTP server with `serve.New(app, store)`.
6. It loads templates with `server.LoadTemplates()`.
7. It starts the HTTP server via `server.Start()`.

## Request Flow
1. Incoming request enters Chi router in `internal/serve`.
2. Middlewares run in order:
- Recoverer, request ID, timeout.
- CSRF middleware (`nosurf`).
- Auth gate (`AuthMiddleware`).
- Template context setup (`setTmplData`).
3. Route dispatches to handlers in `internal/handlers`.
4. Handlers call `logic.Store` methods.
5. Logic calls `repo.Queries` methods.
6. Repo executes SQL against SQLite.
7. Handler renders templates via injected render function.

## Layering
- `cmd/*`: composition and process entrypoints.
- `internal/serve` + `internal/handlers`: HTTP transport.
- `internal/logic`: business/application rules.
- `internal/repo`: SQL persistence layer.
- `internal/db`: database connection, migration, seed plumbing.
- `internal/prog`: runtime config and logging primitives.

## Package Reference

### `cmd/ninete`
- **Role**: Main web app entrypoint.
- **Key file**: `cmd/ninete/main.go`.
- **Responsibilities**:
- Bootstrap app and dependencies.
- Start web server lifecycle.

### `cmd/migrate`
- **Role**: Migration/seed CLI entrypoint.
- **Key file**: `cmd/migrate/main.go`.
- **Responsibilities**:
- Registers migration commands (`up`, `down`, `create`, `status`, `seed`).
- Delegates command execution to `internal/cmd`.

### `internal/cmd`
- **Role**: Small CLI command registry and dispatcher.
- **Key files**: `internal/cmd/cmd.go`, `internal/cmd/errs.go`.
- **Responsibilities**:
- Register command handlers.
- Parse command name from `os.Args`.
- Print usage/help.
- Execute selected command and return exit codes.

### `internal/prog`
- **Role**: Shared runtime primitives.
- **Key files**: `internal/prog/prog.go`, `internal/prog/logger.go`, `internal/prog/utility.go`, `internal/prog/errs.go`.
- **Responsibilities**:
- Load environment configuration.
- Validate `ENV` (`production`, `development`, `test`).
- Load `.env` for non-production.
- Provide structured logger (`Logger`) including SQL timing output.
- Utility conversions like `ToLowerCamel`.

### `internal/db`
- **Role**: Database setup and maintenance operations.
- **Key files**: `internal/db/db.go`, `internal/db/migrate.go`, `internal/db/seed.go`, `internal/db/migrations/*.sql`, `internal/db/init/init.sql`.
- **Responsibilities**:
- Open SQLite DB with connection limits and startup PRAGMAs.
- Execute Goose migrations (embedded SQL files).
- Create migrations interactively.
- Run seed routines using logic layer APIs.

### `internal/repo`
- **Role**: Data access layer (SQL).
- **Key files**:
- Core: `internal/repo/repo.go`, `internal/repo/query_options.go`, `internal/repo/err.go`.
- Entities: `internal/repo/user.go`, `internal/repo/category.go`, `internal/repo/expense.go`, `internal/repo/recurrent_expense.go`.
- **Responsibilities**:
- Encapsulate SQL CRUD/query operations.
- Provide transactional API (`WithTx`, `TxQueries`).
- Build validated query fragments (`Filters`, `Sorting`, `Pagination`).
- Convert nullable DB fields to domain-facing shapes where needed.
- Emit query timing logs through `prog.Logger`.

### `internal/logic`
- **Role**: Application/business logic.
- **Key files**: `internal/logic/logic.go`, `internal/logic/logic_auth.go`, `internal/logic/logic_user.go`, `internal/logic/logic_category.go`, `internal/logic/err.go`.
- **Responsibilities**:
- Own use-case methods exposed to handlers.
- Validate inputs with `go-playground/validator`.
- Handle authentication flow (`Login` + bcrypt password compare).
- Keep route layer free of SQL and validation details.

### `internal/serve`
- **Role**: HTTP infrastructure and server lifecycle.
- **Key files**: `internal/serve/serve.go`, `internal/serve/middleware.go`, `internal/serve/routes.go`, `internal/serve/render.go`, `internal/serve/template.go`, `internal/serve/errs.go`.
- **Responsibilities**:
- Create and configure Chi router and SCS session manager.
- Register routes and middleware stack.
- Handle auth redirect behavior and CSRF setup.
- Build and inject template data map into context.
- Parse and cache templates.
- Graceful start/shutdown with timeouts and signal handling.
- Return hardcoded error when layout template is missing (`ErrLayoutNotFound`).

### `internal/handlers`
- **Role**: HTTP endpoint handlers.
- **Key files**: `internal/handlers/handler.go`, `internal/handlers/auth.go`, `internal/handlers/root.go`, `internal/handlers/dashboard.go`.
- **Responsibilities**:
- Implement endpoint behavior (`/`, `/login`, `/logout`, `/dashboard`).
- Call logic methods and session operations.
- Render responses through injected render function.
- Keep route methods decoupled from `serve.Server` internals.

### `internal/webkeys`
- **Role**: Shared context/session key constants.
- **Key file**: `internal/webkeys/keys.go`.
- **Responsibilities**:
- Provide canonical key names used across `serve` and `handlers`.

### `internal/webtmpl`
- **Role**: Shared template name constants.
- **Key file**: `internal/webtmpl/template_names.go`.
- **Responsibilities**:
- Define typed template identifiers (`webtmpl.Name`).
- Prevent hardcoded template string duplication.

## UI/Assets Structure
- Templates: `web/views/layout.html`, `web/views/*/index.html`, `web/views/common/_*.html`.
- Static files: `web/static/css`, `web/static/js`, `web/static/img`.
- Server serves static assets under `/static/*`.

## Data Model Overview
- `users`: auth identity and password hash.
- `categories`: normalized spending categories with stable `uid`.
- `expenses`: one-off transactions linked to user + category.
- `recurrent_expenses`: recurring templates with month period and optional `last_copy_created_at`.

## Current Exposure vs Domain Scope
- Exposed HTTP routes currently cover auth and dashboard.
- Expense and recurrent-expense domain logic exists at repository level and is ready for additional handlers/routes.

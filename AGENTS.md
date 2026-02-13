# NINETE Architecture Guide

## Purpose
This document gives high-level context so agents can navigate the codebase quickly and make consistent changes.

## Runtime Flow (`cmd/ninete`)
1. `cmd/ninete/main.go` loads application config using `prog.Load()`.
2. It opens SQLite via `db.Open()`.
3. It creates repository queries via `repo.New(app, sqlDB)`.
4. It creates business logic via `logic.New(app, queries)`.
5. It creates the HTTP server via `serve.New(app, store)`.
6. It loads templates via `server.LoadTemplates()`.
7. It starts HTTP serving via `server.Start()`.

## Request Flow (`internal/serve` -> `internal/handlers`)
1. Request enters Chi router in `internal/serve/routes.go`.
2. Global middleware order:
- Recoverer, request ID, timeout.
- CSRF middleware (`nosurf`).
- Auth gate (`AuthMiddleware`).
- Template/context setup (`setTmplData`).
3. Route-level context middleware may run for resource-specific lookups.
4. Handler executes endpoint behavior in `internal/handlers`.
5. Handler calls `logic.Store` methods.
6. Logic calls `repo.Queries` methods.
7. Repo executes SQL against SQLite.
8. Handler renders templates through handler-owned render helpers (`internal/handlers/render.go`), using template lookup/reload callbacks injected by `serve.Server`.

## Layering
- `cmd/*`: process entrypoints and composition.
- `internal/serve` + `internal/handlers`: HTTP transport.
- `internal/logic`: business rules and validation.
- `internal/repo`: SQL persistence.
- `internal/db`: DB open/migrations/seeds.
- `internal/prog`: config/logging/shared utilities.
- `internal/task`: app-level task hooks executed by `cmd/task`.

## Engineering Workflow
- Use `Makefile` targets as the default way to run project commands.
- After implementing changes, run `make lint-fix`.
- After implementing changes, run tests via `make test` (or `make test-verbose` when needed).
- Temporary note: there are no tests yet. For now, keep the test step in the workflow but mark it as skipped when it only reports `[no test files]`.
- Do not create ad-hoc/dynamic errors inline. Define reusable errors in the nearest `errs.go` file to where they are used.
- Use those `errs.go` errors directly or wrap them (for example: `fmt.Errorf("%w", err)`).
- Any temporary file should go under `./tmp/`

## Package Reference

### `cmd/ninete`
- **Role**: Main web app entrypoint.
- **Key file**: `cmd/ninete/main.go`.
- **Responsibilities**:
- Bootstrap dependencies.
- Start server lifecycle.

### `cmd/migrate`
- **Role**: Migration/seed CLI entrypoint.
- **Key file**: `cmd/migrate/main.go`.
- **Responsibilities**:
- Register migration commands (`up`, `down`, `create`, `status`, `seed`).
- Delegate execution to `internal/db` functions via `internal/cmd`.

### `cmd/task`
- **Role**: Task CLI entrypoint.
- **Key file**: `cmd/task/main.go`.
- **Responsibilities**:
- Register task commands.
- Bootstrap app/db/store and run task functions from `internal/task`.

### `internal/cmd`
- **Role**: CLI command registry/dispatcher.
- **Key files**: `internal/cmd/cmd.go`..
- **Responsibilities**:
- Register command handlers.
- Parse command names from args.
- Print usage/help.
- Execute selected command and return exit codes.

### `internal/prog`
- **Role**: Runtime primitives.
- **Key files**: `internal/prog/prog.go`, `internal/prog/logger.go`, `internal/prog/utility.go`.
- **Responsibilities**:
- Load environment configuration.
- Validate `ENV` (`production`, `development`, `test`).
- Load `.env` outside production.
- Provide app logger (`Logger`) with query timing support.
- Shared utility parsing/conversion helpers.

### `internal/db`
- **Role**: Database setup and maintenance.
- **Key files**: `internal/db/db.go`, `internal/db/migrate.go`, `internal/db/seed.go`, `internal/db/migrations/*.sql`, `internal/db/init/init.sql`.
- **Responsibilities**:
- Open SQLite with startup PRAGMAs.
- Execute Goose migrations.
- Create new migration files.
- Run seed routines.

### `internal/repo`
- **Role**: SQL data access layer.
- **Key files**:
- Core: `internal/repo/repo.go`, `internal/repo/query_options.go`..
- Domain query files follow `internal/repo/*.go` by resource.
- **Responsibilities**:
- Implement SQL CRUD and query operations.
- Provide transaction API (`WithTx`, `TxQueries`).
- Validate/filter sorting/pagination query options.
- Emit query timing logs through `prog.Logger`.
- Enforce ownership constraints where applicable (example: expense update/delete scoped by user).

### `internal/logic`
- **Role**: Application/business logic.
- **Key files**: `internal/logic/logic.go`.
- **Responsibilities**:
- Expose use-cases to handlers.
- Validate inputs (`go-playground/validator`).
- Handle auth flows.
- Keep route layer free of SQL details.
- Logic files must be named with the `logic_` prefix.

### `internal/serve`
- **Role**: HTTP server infrastructure/lifecycle.
- **Key files**: `internal/serve/serve.go`, `internal/serve/middleware.go`, `internal/serve/routes.go`, `internal/serve/template.go`.
- **Responsibilities**:
- Configure Chi router and SCS session manager.
- Register global middleware and routes.
- Configure CSRF and auth redirection.
- Build and inject template/request context data.
- Parse/cache templates and expose lookup callback to handlers.
- Start and gracefully shut down HTTP server.

### `internal/handlers`
- **Role**: HTTP handlers and rendering.
- **Key files**: `internal/handlers/handler.go`, `internal/handlers/render.go`, `internal/handlers/constants.go`.
- **Responsibilities**:
- Implement endpoint behavior.
- Use `logic.Store` + session manager for app actions.
- Own template rendering helpers and render error paths.
- Provide context-key and template-name constants.
- Handler endpoint files must be named with the `handle_` prefix.

### `internal/task`
- **Role**: Task hooks used by `cmd/task`.
- **Key file**: `internal/task/task.go`.
- **Responsibilities**:
- Define task entrypoints executed with initialized app/store dependencies.

## Handler File Structure Convention
- ALL handler endpoint files must use the `handle_` prefix (`internal/handlers/handle_*.go`).
- Apply this structure to each handler file:
1. Context middleware(s).
2. Handlers in action order: `index`, `show`, `new`, `edit`, `create`, `update`, `delete`.
3. Exported functions that are not handlers.
4. Unexported functions/helpers.
- Use section-separator comments between blocks.
- If an action is intentionally missing (for example `show`), leave a short note comment in the handler section.

## UI/Assets Structure
- Views follow a resource/action pattern: `web/views/<resource>/<action>.html`.
- Shared layout lives in `web/views/layout.html`.
- Shared partials live in `web/views/common/_*.html`.
- Static assets live under `web/static/` (for example css/js/img).
- Route definitions are the source of truth in `internal/serve/routes.go`.

## Data Model Overview
- `users`: authentication identity and password hash.
- `categories`: normalized category catalog (`name`, `uid`).
- `expenses`: one-off transactions linked to `user_id` and `category_id`.
- `recurrent_expenses`: recurring templates with period and copy-tracking metadata.

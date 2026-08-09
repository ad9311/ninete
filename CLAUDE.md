# NINETE Architecture Guide

## Purpose
This document gives high-level context so agents can navigate the codebase quickly and make consistent changes.

## Project Scope
NINETE is a personal tracking app (expenses, macros/nutrition, foods, moods). It has one user — the owner — and will almost certainly never have two people using it at the same time. Treat that as a fixed design constraint, not a temporary stage the project will grow out of.

The app is still built multi-user and must stay that way: auth flows exist, and every table holding personal data is scoped by `user_id` (`categories` is the one shared lookup table). That scoping is an ownership and correctness boundary — a query that forgets `user_id` is a bug that leaks or destroys another account's data — and it is cheap to maintain. It is not an ambition to serve many people at once.

The distinction matters most when deciding what to optimize. See Performance Priorities.

## Performance Priorities
The goal is an app that feels instant for one person, not one that sustains throughput for many. When those two goals conflict, choose responsiveness.

**Worth the effort**
- Reduce how many queries a page costs, and keep each one index-backed. Any user-scoped query bounded by a timestamp should have a matching `(user_id, <time column>)` index; `internal/db/index_test.go` pins the important ones with `EXPLAIN QUERY PLAN`.
- Batch lookups instead of one query per row. `repo.SelectTagRows` is the pattern: pass a slice of ids, get rows back, group them in memory with `repo.TagNamesByTargetID`.
- Keep work off paths that should be free. `/static/*` is mounted outside the app middleware chain specifically so serving an asset never loads a session or queries the database.
- Remove anything unbounded in rows, memory, or SQL parameters. SQLite rejects a statement with more than 32766 parameters, which is why `SelectTagRows` batches its `IN (...)` list.

**Not worth the effort**
- Concurrency capacity work: connection-pool tuning, read/write pool splits, caching layers, queues. `MAX_OPEN_CONNS` defaults to 1 and that is fine — one user produces almost no overlapping requests, so serializing them costs microseconds nobody can perceive. Do not propose a pool split as a performance improvement.
- If a pool larger than 1 is ever introduced anyway, three things matter. The PRAGMAs in `internal/db/init/connection.sql` are already applied to every connection through the driver connect hook, so those are fine. `WithTx` opens a deferred transaction, which can fail with `SQLITE_BUSY_SNAPSHOT` under concurrent writers — `_txlock=immediate` would be needed. And `db.Optimize` runs `PRAGMA optimize` on whichever single connection `database/sql` hands it at shutdown; the statistics it refreshes come from that connection's own query history, so with several connections in play it would only ever see a fraction of the workload.

**Measuring**
- Query shape: `EXPLAIN QUERY PLAN`, asserted in tests rather than eyeballed.
- Wall time: the app logs every repo query with its duration outside `ENV=test` (`prog.Logger.Query`), so `make dev` shows what a page actually costs.

## Runtime Flow (`cmd/ninete`)
1. `cmd/ninete/main.go` loads application config using `prog.Load()`.
2. It opens SQLite via `db.Open()`.
3. It creates repository queries via `repo.New(app, sqlDB)`.
4. It creates business logic via `logic.New(app, queries)`.
5. It creates the HTTP server via `serve.New(app, store, sqlDB)` — the `*sql.DB` is handed over because the session store (`scs/sqlite3store`) persists sessions in the same database.
6. It loads templates via `server.LoadTemplates()`.
7. It starts HTTP serving via `server.Start()`.

## Request Flow (`internal/serve` -> `internal/handlers`)
1. Request enters Chi router in `internal/serve/routes.go`.
2. Root middleware, paid by every request including static assets (`setUpMiddlewares`):
- Logger (non-test), Recoverer, request ID.
- Base security headers (`nosniff`, HSTS in production).
3. `/static/*` is mounted on the root router, outside the app chain, and adds only a `Cache-Control` header. Serving an asset must never load a session or query the database — keep it that way.
4. App middleware, only for rendered routes (`setUpAppMiddlewares`, applied to a `chi` group):
- Session load/save (`scs`).
- Request body limit, timeout.
- CSP nonce and headers (`contentSecurityPolicy`).
- CSRF middleware (`nosurf`).
- Template/context setup (`setTmplData`) — this is what makes `h.tmplData(r)` available, so anything calling a render helper must sit inside this group. `NotFound`/`MethodNotAllowed` are registered on the group for that reason.
- Auth gate (`AuthMiddleware`) — redirects guests from protected routes and authenticated users from guest-only routes (`/login`, `/register`).
5. Route-level context middleware may run for resource-specific lookups.
6. Handler executes endpoint behavior in `internal/handlers`.
7. Handler calls `logic.Store` methods.
8. Logic calls `repo.Queries` methods.
9. Repo executes SQL against SQLite.
10. Handler renders templates through handler-owned render helpers (`internal/handlers/render.go`), using template lookup/reload callbacks injected by `serve.Server`.

## Layering
- `cmd/*`: process entrypoints and composition.
- `internal/serve`: HTTP server lifecycle, router/middleware/session wiring, template loading.
- `internal/handlers`: HTTP transport adapter (request parsing, context middleware, response rendering).
- `internal/logic`: business rules/use-cases and validation.
- `internal/repo`: SQL persistence.
- `internal/db`: DB open/migrations/seeds.
- `internal/prog`: config/logging/shared utilities.
- `internal/task`: app-level task hooks executed by `cmd/task`.
- `internal/spec`: test setup/factories for integration-style package tests.
- Preferred dependency direction: handlers -> logic -> repo -> db.

## Feature and Route Map
`internal/serve/routes.go` is the source of truth; this is the orientation map.

| Area | Routes | Handlers | Logic |
| --- | --- | --- | --- |
| Auth | `/login`, `/register`, `/logout` | `handle_auth.go` | `logic_auth.go`, `logic_invitation_code.go` |
| Dashboard | `/dashboard` | `handle_dashboard.go` | reuses expense + macro stores |
| Expenses | `/expenses`, `/expenses/quick`, `/expenses/stats`, `/expenses/{id}` | `handle_expenses.go`, `handle_quick_expense.go`, `expense_search.go` | `logic_expense.go`, `logic_quick_expense.go` |
| Recurrent expenses | `/recurrent-expenses`, `/recurrent-expenses/{id}` | `handle_recurrent_expenses.go` | `logic_recurrent_expense.go` |
| Macros | `/macros`, `/macros/goals`, `/macros/stats`, `/macros/{id}` | `handle_macros.go`, `macro_shared.go` | `logic_macro.go` |
| Foods | `/foods`, `/foods/{id}` | `handle_foods.go` | `logic_food.go` |
| Moods | `/moods`, `/moods/stats`, `/moods/{id}` | `handle_mood_entries.go` | `logic_mood_entry.go`, `mood.go` |
| Account | `/account` and its `delete-all` endpoints | `handle_account.go` | `logic_account.go` |
| Exports | `/exports`, `/exports/expenses.json` | `handle_exports.go` | `logic_export.go` |
| Infrastructure | `/`, `/static/*`, `/csp-report` | `handle_root.go`, `handle_csp_report.go` | — |

Cross-cutting: tags attach to expenses and mood entries (`logic_tag.go`, `repo/tagging.go`); categories are global, not user-scoped (`logic_category.go`).

## Engineering Workflow
- Use `Makefile` targets as the default way to run project commands.
- After implementing changes, run `make lint-fix`.
- After implementing changes, run tests via `make test` (or `make test-verbose` when needed).
- Do not create ad-hoc/dynamic errors inline. Define reusable errors in the nearest `errs.go` file to where they are used.
- Use those `errs.go` errors directly or wrap them (for example: `fmt.Errorf("%w", err)`).
- Any temporary file should go under `./tmp/`

## Testing Conventions
- Test files live in the same directory as the code they test, using external test packages (e.g., `package logic_test`).
- Use package-level `TestMain` when database bootstrapping is needed and run setup through `internal/spec`.
- Write table-driven tests with a `cases` struct containing `name` and `fn func(*testing.T)`.
- Keep test functions uncluttered by delegating repeated setup and record creation to test factories/helpers.
- Use `github.com/stretchr/testify/require` for assertions in tests.
- Ensure test records are unique when sharing a package-level database. Package tests share one database file, so name records after the test (`export_chunk_user`, `pagination_category`) instead of reusing generic names.
- `spec.New(t)` builds app, store and server for a test. For HTTP-level tests, `spec.WrappedHandler()` returns the handler `Start()` serves, `spec.NewGetRequest` / `spec.NewPostRequest` build requests, and `spec.AuthCookies(t, email, password)` performs a real login and returns the cookies for authenticated requests.
- Prefer external test packages, but an internal test file (`package repo`, `package handlers`) is acceptable when the invariant under test is unexported — see `internal/repo/columns_internal_test.go` and `internal/handlers/helpers_internal_test.go`. Name these `*_internal_test.go`.
- When a change fixes a bug, make sure the test fails without the fix before committing. Several tests in this repo carry a comment saying which ones are genuine reproductions and which are only invariant guards; keep that distinction honest.

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
- **Key files**: `internal/cmd/cmd.go`.
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
- **Key files**: `internal/db/db.go`, `internal/db/migrate.go`, `internal/db/seed.go`, `internal/db/migrations/*.sql`, `internal/db/init/database.sql`, `internal/db/init/connection.sql`.
- **Responsibilities**:
- Open SQLite and apply PRAGMAs.
- Execute Goose migrations.
- Create new migration files.
- Run seed routines.
- **PRAGMA split — do not merge these two files back together**: `init/database.sql` holds settings SQLite persists in the database file (encoding, page size, journal mode) and runs once when the pool opens. `init/connection.sql` holds per-connection settings and runs from a driver connect hook registered under the driver name `sqlite3_ninete`, so every connection gets them. Applying `foreign_keys` once at startup leaves later connections with SQLite's default of OFF, which silently skips `ON DELETE CASCADE`. The hook reads the value back and fails the connection if it did not take.
- **Migration index convention**: Simple single-column FK columns use only the inline `REFERENCES "table"("col") ON DELETE CASCADE` declaration — do NOT add a separate `CREATE INDEX` for them. Only add explicit `CREATE INDEX` statements for composite or complex indexes (e.g. `CREATE UNIQUE INDEX … ON "table" ("col_a", lower("col_b"))`). Example of correct inline FK: `"user_id" INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE`.
- **Migration file convention**: named `YYYYMMDDHHMMSS_description.sql`, with `-- +goose Up` and `-- +goose Down` sections. Every migration must set `PRAGMA user_version` — incremented in `Up`, restored to the previous value in `Down`. Read the newest migration to find the current number rather than guessing.
- **Adding a column**: append it at the end of the table. If a migration rebuilds a table or inserts a column mid-list, the matching columns constant in `internal/repo` must be updated in the same change; `TestColumnConstantsMatchSchema` will fail until it is.

### `internal/repo`
- **Role**: SQL data access layer.
- **Key files**:
- Core: `internal/repo/repo.go`, `internal/repo/query_options.go`.
- Domain query files follow `internal/repo/*.go` by resource.
- **Responsibilities**:
- Implement SQL CRUD and query operations.
- Provide transaction API (`WithTx`, `TxQueries`).
- Validate/filter sorting/pagination query options.
- Emit query timing logs through `prog.Logger`.
- Enforce ownership constraints where applicable (example: expense update/delete scoped by user).
- **Query patterns to follow rather than reinvent**:
- `QueryOptions` (`query_options.go`) composes a `WHERE`/`ORDER BY`/`LIMIT OFFSET` tail from `Filters`, `Sorting` and `Pagination`. Callers pass column names, which are validated against the table's `validXFields()` list before reaching SQL. A filter needing real SQL sets `FilterField.Expr` with its own `Args` — that fragment must be repo-defined, never user input (see `ExpenseTagFilter`).
- `Sorting.Build` appends `"id"` as a tiebreaker. Sort columns hold duplicates, and `LIMIT/OFFSET` over a non-deterministic order repeats rows on one page and drops them from another.
- Every file declares a columns constant (`expenseColumns`, `macroEntryColumns`, …) naming its table's columns in physical order, and queries concatenate it. **Do not go back to `SELECT *` or `RETURNING *`**: the `Scan` calls read positionally, so a reordered table would put values in the wrong struct fields with no error from SQLite or the driver.
- Reads and deletes hang off `*Queries`; inserts and updates that participate in a transaction hang off `*TxQueries`. Multi-step writes go through `queries.WithTx`.
- Tags are polymorphic: `taggings` rows carry `taggable_type` + `taggable_id`, with types listed as `TaggableType*` constants. Bulk tag reads batch through `SelectTagRows` + `TagNamesByTargetID`.

### `internal/logic`
- **Role**: Application/business logic.
- **Key files**: `internal/logic/logic.go`, `internal/logic/logic_*.go`.
- **Responsibilities**:
- Expose use-cases to handlers.
- Validate inputs (`go-playground/validator`).
- Handle auth flows.
- Keep route layer free of SQL details.
- The `logic_` prefix is reserved for service/business-use-case files.

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
- **Key files**: `internal/handlers/handler.go`, `internal/handlers/render.go`, `internal/handlers/constants.go`, `internal/handlers/shared.go`, `internal/handlers/expense_shared.go`.
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

### `internal/spec`
- **Role**: Test support package for DB-backed setup and factories.
- **Key files**: `internal/spec/setup.go`, `internal/spec/factory.go`, `internal/spec/spec.go`, `internal/spec/http.go`.
- **Responsibilities**:
- Initialize isolated test DB state.
- Provide reusable factories/helpers for logic tests.
- Provide HTTP test helpers (request builders, CSRF extraction, login cookies).

## File Structure Convention
- ALL handler endpoint files must use the `handle_` prefix (`internal/handlers/handle_*.go`).
- Logic service/business-use-case files must use the `logic_` prefix (`internal/logic/logic_*.go`).
- The `logic_` prefix is ONLY for service-like business logic files (for example: create/update/delete model workflows). Non-service files in `internal/logic` must not use it.
- Unprefixed files in these packages are shared infrastructure, and new code belongs in one of them rather than in a new prefixed file: `handler.go` (dependencies/struct), `render.go` (render helpers), `constants.go` (context keys, template names), `shared.go` and `*_shared.go` (form parsing, pagination, helpers used by several endpoints), `errs.go` (sentinel errors).

## UI/Assets Structure
- Views follow a resource/action pattern: `web/views/<resource>/<action>.html`.
- Shared layout lives in `web/views/layout.html`.
- Shared partials live in `web/views/common/_*.html`.
- Static assets live under `web/static/` (for example css/js/img).
- Route definitions are the source of truth in `internal/serve/routes.go`.
- **Frontend JS**: Uses `@hotwired/turbo` for SPA-like navigation and `@hotwired/stimulus` for lightweight controllers.
- Stimulus entrypoint: `web/static/js/index.ts`. Controllers live in `web/static/js/controllers/`.

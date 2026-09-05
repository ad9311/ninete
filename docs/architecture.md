# Architecture

Orientation reference for the codebase: how a process boots, how a request
travels, and what each package owns. `CLAUDE.md` holds the rules that must not
be broken; this file holds the description that helps you find your way.

## Runtime Flow (`cmd/ninete`)

1. `cmd/ninete/main.go` loads application config using `prog.Load()`.
2. It opens SQLite via `db.Open()`.
3. It creates repository queries via `repo.New(app, sqlDB)`.
4. It creates business logic via `logic.New(app, queries)`.
5. It creates the HTTP server via `serve.New(app, store, sqlDB)` — the `*sql.DB` is handed over because the session store (`scs/sqlite3store`) persists sessions in the same database.
6. It loads the shell template via `server.LoadTemplates()`, then the asset manifest via
   `server.LoadAssetManifest()` — in that order, because the manifest's missing-file branch
   reads the loaded template map to tell "no `web/` tree at all" from "the bundle was never
   built" (`internal/serve/manifest.go`).
7. It starts HTTP serving via `server.Start()`.

## Request Flow (`internal/serve` -> `internal/handlers`)

1. Request enters Chi router in `internal/serve/routes.go`.
2. Root middleware, paid by every request including static assets (`setUpMiddlewares`):
   - `realClientIP`, Logger (non-test), Recoverer, request ID.
   - Base security headers (`nosniff`, HSTS in production).
   - `realClientIP` runs before the logger so access logs record the client rather
     than the proxy. See the "Client IP" invariant in `CLAUDE.md` for why it must
     not be swapped for chi's `middleware.RealIP`.
3. `/static/*` is mounted on the root router by `setUpFileServer`, outside the app
   chain, and adds only a `Cache-Control` header (`staticCacheControl`, five
   minutes — the JS bundle and the stylesheet carry a content hash, but the images
   do not, so the window stays short for their sake and `http.FileServer`
   answers the revalidation with a 304 off `Last-Modified`).
4. App middleware, for the three non-API, non-static routes — `GetApp` (the SPA shell catch-all),
   `POST /logout` and `POST /csp-report` (`setUpAppMiddlewares`, applied to a `chi` group), in
   order:
   - Session load/save (`scs`).
   - Request body limit, then a five-second timeout.
   - CSP nonce and headers (`contentSecurityPolicy`).
   - CSRF middleware (`nosurf`), which exempts `/csp-report`: browsers post violation reports
     automatically and carry no token.
   - Template/context setup (`setTmplData`) — this is what makes `h.tmplData(r)` available, so anything calling a render helper must sit inside this group, which is why `GetApp` is registered here rather than on the root router. Nothing registers `NotFound`/`MethodNotAllowed` on this group any more: the `/*` catch-all answers every unmatched path with the shell, and the client router shows its own "Not found". Only the `/api/*` group registers them (see below).
   - Auth gate (`AuthMiddleware`) — redirects guests from protected routes and authenticated users from guest-only routes (`/login`, `/register`), and lets `/csp-report` through unauthenticated.
5. Route-level context middleware may run for resource-specific lookups.
6. Handler executes endpoint behavior in `internal/handlers`.
7. Handler calls `logic.Store` methods.
8. Logic calls `repo.Queries` methods.
9. Repo executes SQL against SQLite.
10. `GetApp` renders the SPA shell through handler-owned render helpers (`internal/handlers/render.go`), using template lookup/reload callbacks injected by `serve.Server`. It is the only render call left in the codebase — every other route answers JSON.

## The `/api/*` group (`internal/serve/routes.go:setUpAPIRoutes`)

A sibling of the app group above, not a child, registered on its own `chi` sub-router. Shares the
session, body-cap, timeout and CSRF middleware of the app chain and drops the two pieces that
assume HTML:

- **No `setTmplData`.** Nothing under `/api/*` renders a template, so there is no template map.
  `apiAuth` (`internal/serve/middleware.go`) is the API's replacement for the piece of
  `setTmplData` that matters to handlers: it puts the signed-in user into `KeyCurrentUser`, since
  every resource handler opens with `getCurrentUser(r)`, which panics if the key is absent.
- **No `AuthMiddleware`.** `apiAuth` answers an unauthenticated request with `401` and a JSON
  body, never a `Location` header — `AuthMiddleware`'s redirect would otherwise arrive at a
  `fetch` call as a same-origin `200` full of login-page HTML.

`POST /api/login` and `POST /api/register` carry `authRateLimit()`, built once in
`setUpAPIRoutes` so both draw on one shared budget — see the invariant in `CLAUDE.md`. Every
response goes through `internal/handlers/api.go`'s JSON writers (`WriteJSON`, `WriteJSONError`,
`WriteAPIError`), which map validation failures to `422` with `{"error", "fields"}` and unexpected
failures to a generic `500` that never quotes `err.Error()`. There is no CSP on this chain by
design — a JSON response has no document to constrain.

The session cookie is configured in `setUpSession` (`internal/serve/routes.go`):
seven-day lifetime, `HttpOnly`, `SameSite=Lax`, persistent, named
`ninete_session`, and `Secure` only in production.

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
- Register migration commands (`up`, `down`, `create`, `status`, `seed`, `stamp`).
- Delegate execution to `internal/db` functions via `internal/cmd`.

### `cmd/task`
- **Role**: Task CLI entrypoint.
- **Key file**: `cmd/task/main.go`.
- **Responsibilities**:
- Register task commands (`create_invitation_code`, `copy_due_recurrent_expenses`, `test`).
- Bootstrap app/db/store and run task functions from `internal/task`.

### `internal/cmd`
- **Role**: CLI command registry/dispatcher.
- **Key files**: `internal/cmd/cmd.go`.
- **Responsibilities**:
- Register command handlers.
- Parse command names from args.
- Print usage/help.
- Execute selected command and return exit codes.
- Register a `version` command on every binary using `Run`, printing `prog.VersionString()` without touching config or the database.

### `internal/prog`
- **Role**: Runtime primitives.
- **Key files**: `internal/prog/prog.go`, `internal/prog/logger.go`, `internal/prog/utility.go`, `internal/prog/version.go`.
- **Responsibilities**:
- Load environment configuration.
- Validate `ENV` (`production`, `development`, `test`).
- Load `.env` outside production.
- Provide app logger (`Logger`) with query timing support.
- Shared utility parsing/conversion helpers (`SetInt`, `LoadList`, `FindRelativePath`).
- Expose the link-time build identity (`Version`, `Commit`, `BuildTime`, `VersionString`), stamped with `-X` and falling back to `dev`/`unknown` when unstamped.

### `internal/db`
- **Role**: Database setup and maintenance.
- **Key files**: `internal/db/db.go`, `internal/db/migrate.go`, `internal/db/seed.go`, `internal/db/stamp.go`, `internal/db/migrations/*.sql`, `internal/db/init/database.sql`, `internal/db/init/connection.sql`.
- **Responsibilities**:
- Open SQLite and apply PRAGMAs.
- Verify the environment stamp before handing the pool back.
- Execute Goose migrations.
- Create new migration files.
- Run seed routines.
- `Optimize` runs `PRAGMA optimize` at shutdown to refresh query planner statistics.

The invariants that govern this package — the PRAGMA split, the environment
stamp, and the migration conventions — are in `CLAUDE.md`, because breaking one
corrupts data or silently disables `ON DELETE CASCADE`.

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
- Tags are polymorphic: `taggings` rows carry `taggable_type` + `taggable_id`. Each kind is a `repo.Taggable` value built by `TaggableExpense()` / `TaggableRecurrentExpense()`, pairing the `taggable_type` string with the table that owns those records — a tag read joins that table to scope the row to a user, and the table name is interpolated rather than bound, so it must not come from a caller. Bulk tag reads batch through `SelectTagRows` + `TagNamesByTargetID`.

### `internal/logic`
- **Role**: Application/business logic.
- **Key files**: `internal/logic/logic.go`, `internal/logic/logic_*.go`.
- **Responsibilities**:
- Expose use-cases to handlers.
- Validate inputs (`go-playground/validator`).
- Handle auth flows.
- Keep route layer free of SQL details.

### `internal/serve`
- **Role**: HTTP server infrastructure/lifecycle.
- **Key files**: `internal/serve/serve.go`, `internal/serve/middleware.go`, `internal/serve/routes.go`, `internal/serve/template.go`, `internal/serve/manifest.go`.
- **Responsibilities**:
- Configure Chi router and SCS session manager.
- Register global middleware and routes.
- Configure CSRF and auth redirection.
- Build and inject template/request context data.
- Parse/cache the shell template and expose lookup/reload callbacks to handlers.
- Read the asset manifest (`manifest.go`) and resolve the content-hashed bundle and stylesheet paths handed to the shell.
- Start and gracefully shut down HTTP server.

The shell template and static assets this package serves are documented in
`web/README.md`, including the shell's data contract, the CSP nonce rule and the
Svelte build chain.

### `internal/handlers`
- **Role**: HTTP handlers and rendering.
- **Key files**: `internal/handlers/handler.go`, `internal/handlers/render.go`, `internal/handlers/constants.go`, `internal/handlers/shared.go`, `internal/handlers/expense_shared.go`, `internal/handlers/expense_search.go`, `internal/handlers/api.go`.
- **Responsibilities**:
- Implement endpoint behavior.
- Use `logic.Store` + session manager for app actions.
- Own the SPA shell's render helper (`render.go`) and the `/api/*` JSON writers/error mapper (`api.go`).
- Provide context-key and template-name constants.

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

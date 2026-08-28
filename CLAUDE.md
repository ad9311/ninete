# NINETE Architecture Guide

## Purpose
This document holds the rules and invariants that must not be broken, plus the
map needed to find things.

### Documentation map
Every document in the repository, and when to read it. This list is the index —
if a document is added, add it here too, or nobody will find it.

| Document | What it holds | Read it when |
| --- | --- | --- |
| `CLAUDE.md` (this file) | Rules, invariants, conventions, route map | Always. It is loaded for you |
| `docs/architecture.md` | Runtime flow, request flow, per-package reference | Orienting in unfamiliar packages |
| `docs/spa-migration.md` | The staged plan for replacing the templates + Turbo frontend with a Svelte SPA: inventory, cross-cutting concerns (auth, CSRF, CSP, **dates**), where `.svelte` sources live, why Tailwind and component libraries are deferred, phases, decisions already made, **and the scope reduction dropping macros/foods/moods (§0)** | Before any frontend work while the migration is running |
| `docs/performance.md` | What optimization work pays off here and what does not | Before proposing any performance change |
| `docs/deployment.md` | How the app runs in production: deploy scripts, systemd unit, Caddy, migrations, versioning, backups, rollback | Answering anything about production, or editing `scripts/` |
| `docs/deployment.local.md` | Host specifics: paths, service account, hostname, scheduled jobs, known gaps. Git-ignored, exists only on the maintainer's machine and the host | Touching the deploy account or the host config. Assume it exists even if you cannot read it |
| `web/README.md` | How the three directories under `web/` work and how code reaches the browser: partial namespace, template data contract, CSP nonce rule, Stimulus controller registration, the Svelte build chain | **Before editing anything under `web/`** |
| `web/app/README.md` | Working rules for the Svelte sources: what belongs in `lib/`, `components/`, `routes/<resource>/`, where tests go, and what lint and formatting cover | Before adding a file under `web/app/` |
| `TODO.md` | Known bugs and follow-up work deliberately left out of the change that surfaced them | Before reporting a bug as new, and before "fixing" something adjacent |
| `README.md` | Setup, prerequisites, commands, troubleshooting | Running the project locally for the first time |

## Project Scope
NINETE is a personal tracking app for expenses. It has one user — the owner — and will almost certainly never have two people using it at the same time. Treat that as a fixed design constraint, not a temporary stage the project will grow out of.

**Macros, foods and moods have been dropped.** The decision is recorded in `docs/spa-migration.md` §0 and the code was removed in that plan's Phase 0B. The *tables* survive until Phase 8, along with `internal/repo/{macro_entry,macro_goal,food,mood_entry}.go` and their column constants, so `TestColumnConstantsMatchSchema` keeps guarding the schema; those repo files have no callers on purpose. Do not revive the features and do not port them to the SPA. Expenses and what hangs off them — recurrent expenses, budgets, tags, categories, dashboard, exports, delete-data, auth — are what the app keeps.

The app is still built multi-user and must stay that way: auth flows exist, and every table holding personal data is scoped by `user_id` (`categories` is the one shared lookup table). That scoping is an ownership and correctness boundary — a query that forgets `user_id` is a bug that leaks or destroys another account's data — and it is cheap to maintain. It is not an ambition to serve many people at once.

The distinction matters most when deciding what to optimize. **Do not propose
connection-pool tuning, read/write pool splits, caching layers or queues as a
performance improvement** — `MAX_OPEN_CONNS` defaults to 1 and that is correct
for this app. See `docs/performance.md` for what is worth the effort instead.

## Deployment
Production runs on a single Linux VPS with a standard FHS layout — no containers. `docs/deployment.md` covers the deploy script chain, the systemd unit, Caddy, migrations, and rollback. Read it before answering anything about how the app runs in production.

Host specifics — exact paths, the service account, the hostname, cron, and the list of known gaps — are in `docs/deployment.local.md`, which is git-ignored because this repository is public. Keep it that way: a note that helps someone write code belongs in `deployment.md`, a note that helps someone reach the box belongs in `deployment.local.md`.

Four facts from it that change how code should be written:
- `prog.Load()` skips `.env` when `ENV=production`, so production config comes only from the process environment (systemd `EnvironmentFile=/etc/ninete/env`). A new config value must be added there, not to a `.env` on the host.
- Templates and static assets resolve through relative paths (`./web/views/...`, `./web/static/`), so the binary only works when started from the repository root. `WorkingDirectory` in the unit is load-bearing — do not introduce more relative-path dependencies without noting it.
- Migrations are applied while the previous binary is still serving. A migration that removes or renames something the running code reads will error until the restart lands.
- The unit runs under `ProtectSystem=strict`, so the app process can write to exactly one directory — the one holding the SQLite file. Runtime writes to the working directory, `/tmp`, or anywhere else fail with `EPERM`. The `./tmp/` convention below is for development only. The sandbox covers only the web process: `migrate.sh` and `task.sh` run outside the unit and are unconstrained, so a file write proven through a task will still fail from a handler. Put anything a task writes beside the database.

The deploy account's privilege setup is deliberate, settled, and documented in `deployment.local.md`. It is not an open weakness, and the trade-offs behind it — including what was considered and rejected — are recorded there. Do not propose changes to it as a security improvement; if something genuinely looks wrong, raise it with the owner rather than acting. The one part that touches code: the privileged deploy steps are granted per exact command line, so changing which commands the deploy runs, or their arguments, breaks the grant until the host config is updated to match. Read `deployment.local.md` before touching anything in that area.

## Invariants
Each of these breaks something real if undone. `docs/architecture.md` describes
the surrounding machinery; this is the part you must not get wrong.

**Client IP** — `realClientIP` (`internal/serve/middleware.go`) rewrites
`RemoteAddr` from the **last** `X-Forwarded-For` entry, which is the only one
Caddy itself wrote. Do not replace it with chi's `middleware.RealIP`: that reads
`True-Client-IP`, then `X-Real-IP`, then the *first* `X-Forwarded-For` entry, and
a stock Caddy sets neither of the first two and appends to the third — so all
three arrive client-controlled and let a caller mint a fresh rate-limit bucket
per request by rotating a header. Binding loopback (`defaultHost` in
`internal/serve/serve.go`) stops direct connections; it does nothing about forged
headers arriving through the proxy.

**Static assets stay off the app chain** — `/static/*` is mounted on the root
router by `setUpFileServer`, outside `setUpAppMiddlewares`. Serving an asset must
never load a session or query the database.

**Auth rate limit is one shared value** — `POST /login` and `POST /register`
carry the same `authRateLimit()` middleware value, applied with
`root.With(...)`, so a client draws on a single budget across both. Calling
`authRateLimit()` twice would build two independent counters and double the
allowance. It is a no-op under `ENV=test` — the suite logs in ~100 times from one
address — and is covered directly in
`internal/serve/middleware_internal_test.go` instead.

**Render helpers need `setTmplData`** — anything calling a render helper must sit
inside the app middleware group, which is why `NotFound`/`MethodNotAllowed` are
registered on the group rather than the root router.

**PRAGMA split — do not merge `internal/db/init/database.sql` and
`internal/db/init/connection.sql` back together.** `database.sql` holds settings
SQLite persists in the database file (encoding, page size, journal mode) and runs
once when the pool opens. `connection.sql` holds per-connection settings and runs
from a driver connect hook registered under the driver name `sqlite3_ninete`, so
every connection gets them. Applying `foreign_keys` once at startup leaves later
connections with SQLite's default of OFF, which silently skips
`ON DELETE CASCADE`. The hook reads the value back and fails the connection if it
did not take.

**Environment stamp** — `ENV` and `DATABASE_URL` are independent, so nothing else
stops a development command from opening the production database.
`verifyEnvStamp` (`internal/db/stamp.go`) writes the owning environment into the
SQLite header's `application_id` on first open and fails `Open` when a later open
disagrees. That is why `init/database.sql` must not set `application_id` itself. A
database created before stamping reads back as unstamped and is claimed by
whichever command opens it first, so claim it deliberately once with
`ENV=<env> ./build/migrate stamp` (or `make stamp`).

**No `SELECT *` or `RETURNING *` in `internal/repo`.** Every file declares a
columns constant (`expenseColumns`, `macroEntryColumns`, …) naming its table's
columns in physical order, and queries concatenate it. The `Scan` calls read
positionally, so a reordered table would put values in the wrong struct fields
with no error from SQLite or the driver.

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
| Dashboard | `/dashboard` | `handle_dashboard.go` | reuses the expense stores |
| Expenses | `/expenses`, `/expenses/quick`, `/expenses/stats`, `/expenses/budgets`, `/expenses/{id}` | `handle_expenses.go`, `handle_quick_expense.go`, `handle_expense_budgets.go`, `expense_search.go` | `logic_expense.go`, `logic_quick_expense.go`, `logic_expense_budget.go` |
| Recurrent expenses | `/recurrent-expenses`, `/recurrent-expenses/archived`, `/recurrent-expenses/{id}` | `handle_recurrent_expenses.go` | `logic_recurrent_expense.go` |
| Account | `/account` | `handle_account.go` | — |
| Delete data | `/account/delete-data` and the `delete-all` endpoints | `handle_delete_data.go` | `logic_account.go` |
| Exports | `/account/exports`, `/account/exports/expenses.json` | `handle_exports.go` | `logic_export.go` |
| Infrastructure | `/`, `/static/*`, `/csp-report` | `handle_root.go`, `handle_csp_report.go` | — |
| API (SPA) | `/api/login`, `/api/register`, `/api/session`, `/api/categories`, `/api/dashboard`, `/api/exports/expenses.json`, `/api/delete-data` (+ `/expenses`, `/recurrent-expenses`, `/expense-budgets`, `/tags`), `/api/recurrent-expenses`, `/api/expenses` (+ `/quick`, `/stats`, `/budgets`), and the resource routes later phases add | `api.go`, `handle_api_auth.go`, `handle_api_session.go`, `handle_api_categories.go`, `handle_api_dashboard.go`, `handle_api_exports.go`, `handle_api_delete_data.go`, `handle_api_recurrent_expenses.go`, `handle_api_expenses.go`, `handle_api_quick_expense.go`, `handle_api_expense_stats.go`, `handle_api_expense_budgets.go` | reuses the page handlers' stores |
| SPA shell | `/app`, `/app/*` (moves to `/` in Phase 7) — `/app/login`/`/app/register` are the guest-reachable exception, mirroring `/login`/`/register` | `handle_app.go` | — |

**The `/api/*` group is a sibling of the page group, not a child** (`setUpAPIRoutes`). The two
middleware chains differ, and an `/api` route must never fall through to a rendered template.
`setUpAPIMiddlewares` shares the session, body cap, timeout and CSRF of the page chain and drops
the two pieces that assume HTML: `setTmplData` (no templates) and `AuthMiddleware` (which
redirects rather than answering `401`). Three consequences worth knowing before adding a route:

- **`apiAuth` must keep putting the user in `KeyCurrentUser`.** Dropping `setTmplData` drops the
  only other place that happens, and `getCurrentUser` *panics* when the key is absent — so a
  handler opening with `user := getCurrentUser(r)` would 500 in a way that looks like a client
  bug.
- **`api.NotFound`/`api.MethodNotAllowed` are registered on the group** so an unmatched path
  answers with the JSON envelope. They only take effect once the group has at least one real
  route: a chi sub-router with none never builds its middleware chain.
- **Responses go through `api.go`** — `WriteJSON`, `WriteJSONError`, `WriteAPIError` — which maps
  validation to `422` with `{"error", "fields"}` and unexpected failures to a generic `500` that
  never quotes `err.Error()`. There is no CSP on this chain by design; a JSON response has no
  document to constrain.

Cross-cutting: tags attach to expenses and recurrent expenses (`logic_tag.go`, `repo/tagging.go`); a recurrent expense copies its tags onto every expense it generates, and archives itself once it has generated `occurrence_limit` copies (0 means unlimited), staying out of the cron job until unarchived by hand; categories are global, not user-scoped (`logic_category.go`).

## Engineering Workflow
- **Always write in English.** Code, comments, identifiers, commit messages, PR titles and
  descriptions, documentation, migration names, log messages, and user-facing strings are all
  English, with no exceptions. This holds regardless of the language a request is written in.
- Use `Makefile` targets as the default way to run project commands.
- After implementing changes, run `make lint-fix`. It covers Go, CSS, JS, type
  checking, and the shell scripts under `scripts/` (`make lint-sh` runs shellcheck alone). The shell
  step runs last and skips with a warning when shellcheck is not installed, so
  install it (`brew install shellcheck`) before editing `scripts/` — CI runs a
  pinned version and will not skip.
- **Type checking is two tools, not one.** `bun run typecheck:ts` (`tsc --noEmit`)
  covers `.ts`; `bun run typecheck:svelte` (`svelte-check`) covers `.svelte`,
  which `tsc` cannot parse at all. Both run in `make lint-fix` and in the
  `typecheck` CI job. Dropping either half leaves that language unchecked while
  the run still goes green.
- After implementing changes, run tests via `make test` (or `make test-verbose` when needed).
- `make test` is the Go suite only. Anything touching `web/app/` also needs `make test-js`,
  which runs the frontend suite in both configured time zones — see `web/app/README.md`.
- Do not create ad-hoc/dynamic errors inline. Define reusable errors in the nearest `errs.go` file to where they are used.
- Use those `errs.go` errors directly or wrap them (for example: `fmt.Errorf("%w", err)`).
- **Build identity is stamped, not stored.** `internal/prog.Version`/`Commit`/`BuildTime` are set with `-X` from `git describe`, by `scripts/build.sh` and the Makefile. There is no `VERSION` file to bump, and `package.json`'s `version` field describes nothing — do not treat it as the app's version. Unstamped builds fall back to `dev`/`unknown`, so nothing may depend on the stamp being present. `docs/deployment.md` has the details.
- Any temporary file should go under `./tmp/`
- Adding a document to the repository means adding a row to the Documentation map above. A document nobody can find is a document nobody reads.

## Migration Conventions
- **File naming**: `YYYYMMDDHHMMSS_description.sql` under `internal/db/migrations/`, with `-- +goose Up` and `-- +goose Down` sections. Every migration must set `PRAGMA user_version` — incremented in `Up`, restored to the previous value in `Down`. Read the newest migration to find the current number rather than guessing.
- **Indexes**: Simple single-column FK columns use only the inline `REFERENCES "table"("col") ON DELETE CASCADE` declaration — do NOT add a separate `CREATE INDEX` for them. Only add explicit `CREATE INDEX` statements for composite or complex indexes (e.g. `CREATE UNIQUE INDEX … ON "table" ("col_a", lower("col_b"))`). Example of correct inline FK: `"user_id" INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE`.
- **Adding a column**: append it at the end of the table. If a migration rebuilds a table or inserts a column mid-list, the matching columns constant in `internal/repo` must be updated in the same change; `TestColumnConstantsMatchSchema` will fail until it is.

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

## File Structure Convention
- ALL handler endpoint files must use the `handle_` prefix (`internal/handlers/handle_*.go`).
- Logic service/business-use-case files must use the `logic_` prefix (`internal/logic/logic_*.go`).
- The `logic_` prefix is ONLY for service-like business logic files (for example: create/update/delete model workflows). Non-service files in `internal/logic` must not use it.
- Unprefixed files in these packages are shared infrastructure, and new code belongs in one of them rather than in a new prefixed file: `handler.go` (dependencies/struct), `render.go` (render helpers), `constants.go` (context keys, template names), `shared.go` and `*_shared.go` (form parsing, pagination, helpers used by several endpoints), `errs.go` (sentinel errors), `api.go` (JSON writers and the `/api/*` error mapping).
- Reads and deletes hang off `*Queries`; inserts and updates that participate in a transaction hang off `*TxQueries`. Multi-step writes go through `queries.WithTx`.
- `scripts/*.sh` holds the production deploy scripts, run on the host through a symlink. `rollback.sh` is the exception: never part of a deploy, only run by hand, and the only script that can destroy data (`--with-database` replaces the live database with a snapshot). They carry two constraints that are easy to undo by accident — the `main()` wrap ending in `main "$@"; exit`, and `cd -P` for paths into the checkout — because a deploy rewrites these files while they are running. Read the "Individual scripts" section of `docs/deployment.md` before editing one. They have no test coverage; `make lint-sh` (shellcheck) is the only check.

## UI/Assets Structure
**`web/README.md` is the reference for everything under `web/`. Read it before
editing a template, a stylesheet, a controller or a Svelte component** — it
covers the template data contract, the partial namespace, the Turbo/Stimulus
wiring, the Svelte build chain and the editing loop. What follows is the shape,
plus the things that fail silently if you do not know them going in.

- Views follow a resource/action pattern: `web/views/<resource>/<action>.html`.
- Shared layout lives in `web/views/layout.html`.
- Shared partials live in `web/views/common/_*.html`.
- Static assets live under `web/static/` (for example css/js/img).
- Route definitions are the source of truth in `internal/serve/routes.go`.
- **Frontend JS**: Uses `@hotwired/turbo` for SPA-like navigation and `@hotwired/stimulus` for lightweight controllers.
- Stimulus entrypoint: `web/static/js/index.ts`. Controllers live in `web/static/js/controllers/`.
- **Svelte sources live in `web/app/`, and are never served.** The SPA migration
  (`docs/spa-migration.md`) is replacing the templates and Turbo with Svelte, phase by phase, on
  the `spa-base` branch. `web/app/README.md` has the layout rules.
- **The bundle is generated and git-ignored: run `make build-static-js` after editing any `.ts`
  or `.svelte`.** `make dev` and `make test` do it for you; running `go test` directly does not,
  and `internal/serve` asserts `/static/*` serves the bundle — so a skipped rebuild surfaces as
  a red Go suite on a path unrelated to the change.
- **`make test` does not run the frontend suite.** `make test-js` does, in two time zones. See
  `docs/spa-migration.md` §3.6 before touching anything dated.
- **Loading feedback is already global.** A spinner covers every Turbo visit
  and form submission; a new form or listing needs nothing added. It is Turbo's
  `.turbo-progress-bar` element restyled in `layout.css`, not an overlay of
  ours. Only markup that opts out of Turbo loses it.

Frontend failures that produce no build error and no obvious symptom:

- **Partial `define` names share one global namespace.** Every `_*.html` in a
  resource directory under `web/views` is parsed into the same base template
  (`parseTemplates` in `internal/serve/template.go`), so a duplicate name
  silently overwrites another page's partial.
- **Templates live exactly one directory below `web/views`.** Both globs in
  `template.go` use `**`, which Go's `filepath.Glob` treats as a single path
  segment, not a recursive match. A view or partial placed directly in
  `web/views`, or nested two levels deep, is never parsed and surfaces only as a
  failed render.
- **A view needs a matching `handlers.TemplateName` constant.** Nothing checks
  the two agree at compile time; a mismatch logs `missing template` and returns a
  500 at request time.
- **Inline `<script>` and `<style>` must carry `nonce="{{ .cspNonce }}"`.**
  Without it the browser drops the tag and posts to `/csp-report`.
- **A new Stimulus controller is inert until registered in `index.ts`.**

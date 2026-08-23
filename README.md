# NINETE

A personal tracking app, built around four areas:

- **Expenses** — categories and tags, quick entry, search by description, tag or
  date range, per-category monthly budgets, and stats.
- **Recurrent expenses** — copied into real expenses on a schedule by a task,
  carrying their tags, and archived once they hit an optional occurrence limit.
- **Nutrition** — macro entries against daily goals, plus a personal food library
  used to prefill them.
- **Moods** — tagged daily entries with stats.

Alongside those: a dashboard summarizing spend and macro progress, a JSON export
of expenses, and an account page for bulk-deleting any of the data above.

In practice it runs single-user. Data stays user-scoped for correctness, but the app is tuned for one person's responsiveness rather than for concurrent capacity — see the Project Scope section of [`CLAUDE.md`](CLAUDE.md) and [`docs/performance.md`](docs/performance.md) before optimizing anything.

## Prerequisites

- **Go** 1.25.6 or higher
- **Bun** (for installing JS deps and building static assets)
- **golangci-lint** (for linting)
- **shellcheck** (for linting the deploy scripts under `scripts/`) — `make lint-sh`
  skips with a warning when it is absent, but CI runs a pinned version and will
  not skip, so install it (`brew install shellcheck`) before editing `scripts/`
- **A C compiler** — the app uses `mattn/go-sqlite3`, which requires CGO (`CGO_ENABLED=1`) and
  a C toolchain:
  - **macOS**: Xcode Command Line Tools (`clang`) — install with `xcode-select --install`
  - **Linux**: a C toolchain. `.env.example` targets `musl-gcc` (install `musl-tools`); or use
    `gcc` and adjust `CC` in `GO_BUILD_ENVS` accordingly

## Setup

### 1. Clone the repository

```bash
git clone <repo-url>
cd ninete
```

### 2. Install dependencies

Fetch the Go dependencies:

```bash
make deps
```

This runs `go mod download` and `go mod tidy`.

Then install the JavaScript dependencies (needed for `make build-static-js` and `make lint-fix`):

```bash
bun install
```

### 3. Configure environment

Copy `.env.example` to `.env` and configure as needed:

```bash
cp .env.example .env
```

Key variables:
- `GO_BUILD_ENVS`: **Required.** CGO build environment for the host OS. Set the C compiler and
  target for your platform, for example:
  - **macOS (Apple Silicon)**: `CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 CC=clang`
  - **Linux (amd64, musl)**: `CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GOAMD64=v3 CC=musl-gcc`
- `DATABASE_URL`: **Required.** Absolute path to the SQLite database file (e.g.
  `/path/to/ninete/data/db/dev/main.db`). The file is created on first migration, but the
  variable itself must be set.
- `PORT`: HTTP server port (default: 8080)
- `HOST`: Listen address. Empty or unset binds loopback (`127.0.0.1`), which is
  what the production reverse proxy connects to. That default is a security
  boundary — see [`docs/deployment.md`](docs/deployment.md) before changing it.
- `MAX_IDLE_CONNS`: Max idle database connections (optional; defaults to 1)
- `MAX_OPEN_CONNS`: Max open database connections (optional; defaults to 1 — see
  [`docs/performance.md`](docs/performance.md) for why raising it is not an
  improvement here)

`ENV` is not read from `.env`: the `Makefile` targets set it (`development` for
`make dev`, `test` for `make test`), and production supplies it through the
process environment. Valid values are `production`, `development` and `test`.

### 4. Initialize the database

```bash
make migrate
```

This runs all pending migrations. Optionally seed with sample data:

```bash
make seed
```

On first open the database is stamped with the environment that owns it, and any
later open from a different `ENV` fails. A fresh `make migrate` stamps it for
you; a database file created before stamping existed reads back as unstamped and
is claimed by whichever command opens it first, so claim it deliberately:

```bash
make stamp
```

`make stamp` claims the database for `development`. For another environment, run
the binary directly: `ENV=<env> ./build/migrate stamp`.

## Running the App

Start the development server:

```bash
make dev
```

The app will start at `http://localhost:8080` (or the port configured in `.env`).

The development build:
- Rebuilds the static JS bundle
- Compiles the Go binary and runs it with `ENV=development`
- Re-parses templates server-side on render (throttled), so template edits show up on the next
  page load without recompiling — refresh the browser to see them. Navigation itself is
  SPA-style via Turbo.

### Other useful commands:

- `make build` — Build the binary without running
- `make version` — Print the version this checkout would build
- `make snapshot` — Write a snapshot of the development database
- `make build-static-js` — Build only the static JS bundle
- `make test-js` — Run the frontend tests in both configured time zones
- `make lint` — Run golangci-lint and shellcheck without fixing
- `make lint-fix` — Run all formatters and linters with automatic fixes
- `make lint-sh` — Run shellcheck over `scripts/*.sh` alone
- `make task name=<task>` — Run a task (see below)
- `make clean` — Remove compiled binaries
- `make clean-db` — Reset the development database
- `make help` — List every target with its description

## Tasks

Tasks are one-off or scheduled jobs that run outside the web process, registered
in `cmd/task/main.go`:

```bash
make task name=create_invitation_code   # prompts on stdin for a code
make task name=copy_due_recurrent_expenses
```

`copy_due_recurrent_expenses` materializes due recurrent expenses into real
expenses and is the one meant to run on a schedule in production — see
[`docs/deployment.md`](docs/deployment.md).

## Running Tests

Run the full test suite:

```bash
make test
```

Run tests in verbose mode:

```bash
make test-verbose
```

Run a specific test function, or restrict the run to one package:

```bash
make test func=TestName
make test pkg=./internal/logic/...
```

Tests use an isolated test database (`./data/db/test/`), which is automatically cleaned before each run.

`make test` covers the Go suite only. The frontend tests under `web/app/` run
separately:

```bash
make test-js
```

That runs the suite twice, once in `Pacific/Auckland` and once in
`America/Los_Angeles`. Both signs of UTC offset are needed: a calendar date
formatted with local getters still reads correctly east of UTC and only breaks
west of it, so neither run alone is enough. Set `TEST_TZ` to run a single zone
by hand (`TEST_TZ=Europe/Madrid bun run test:js`). CI runs both as two steps of
one job.

`TEST_TZ=UTC` is the one value that does not work: `dates.test.ts` asserts the
zone is not UTC, deliberately, because every date bug the suite exists to catch
passes silently at offset zero. That guard failing is the suite telling you it
has been neutered.

## Development Workflow

After implementing changes:

1. **Lint and format** your code:
   ```bash
   make lint-fix
   ```

2. **Run tests**:
   ```bash
   make test
   make test-js
   ```

3. **Verify the app** runs without errors:
   ```bash
   make dev
   ```

## Database Migrations

Create a new migration:

```bash
make migrate-create name=add_new_table
```

Check migration status:

```bash
make migrate-status
```

Rollback one migration:

```bash
make migrate-down
```

Migrations live in `internal/db/migrations/` and are managed with Goose.

## Deployment

Production runs on a single Linux VPS (systemd + Caddy, no containers). See
[`docs/deployment.md`](docs/deployment.md) for the deploy procedure, environment
handling, migrations, tasks, and rollback.

## Project Structure

`CLAUDE.md` holds the conventions and the invariants that must not be broken, and
opens with a map of every document in the repository.
[`docs/architecture.md`](docs/architecture.md) describes the runtime flow,
request flow, and what each package owns;
[`web/README.md`](web/README.md) covers the frontend half.

Quick overview:
- `cmd/` — CLI entrypoints (app, migrations, tasks)
- `internal/serve/` — HTTP server, router, middleware, templates
- `internal/handlers/` — HTTP handlers
- `internal/logic/` — Business logic
- `internal/repo/` — Data access (SQL)
- `internal/db/` — Database setup and migrations
- `internal/task/` — Task entrypoints run by `cmd/task`
- `internal/spec/` — Test setup and factories
- `web/` — HTML templates and static assets (JS, CSS), documented in [`web/README.md`](web/README.md)
- `scripts/` — Production deploy scripts, run on the host
- `docs/` — Architecture, performance and deployment references
- `TODO.md` — Known bugs and follow-up work left out of the change that surfaced them

## Troubleshooting

**"CGO_ENABLED=1 is required"**
Make sure `CGO_ENABLED=1` is set. On macOS, Xcode Command Line Tools may be required.

**Database locked**
If you see "database is locked", ensure no other instances of the app are running. Clean and reinitialize:
```bash
make clean-full
make migrate
```

**Static assets not updating**
Rebuild the JS bundle:
```bash
make build-static-js
```

## License

Licensed under the GNU General Public License v3.0. See the [`LICENSE`](LICENSE) file for the
full text.

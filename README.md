# NINETE

A personal expense and nutrition tracking app. Supports multiple users with auth, expenses (with categories and tags), recurring expenses, macro tracking with daily goals, and a personal food library.

## Prerequisites

- **Go** 1.25.6 or higher
- **Bun** (for installing JS deps and building static assets)
- **golangci-lint** (for linting)
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
- `MAX_IDLE_CONNS`: Max idle database connections (optional; defaults applied when unset)
- `MAX_OPEN_CONNS`: Max open database connections (optional; defaults applied when unset)

### 4. Initialize the database

```bash
make migrate
```

This runs all pending migrations. Optionally seed with sample data:

```bash
make seed
```

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
- `make build-static-js` — Build only the static JS bundle
- `make clean` — Remove compiled binaries
- `make clean-db` — Reset the development database

## Running Tests

Run the full test suite:

```bash
make test
```

Run tests in verbose mode:

```bash
make test-verbose
```

Run a specific test function:

```bash
make test func=TestName
```

Tests use an isolated test database (`./data/db/test/`), which is automatically cleaned before each run.

## Development Workflow

After implementing changes:

1. **Lint and format** your code:
   ```bash
   make lint-fix
   ```

2. **Run tests**:
   ```bash
   make test
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

## Project Structure

See `CLAUDE.md` for detailed architecture, layering, and conventions.

Quick overview:
- `cmd/` — CLI entrypoints (app, migrations, tasks)
- `internal/handlers/` — HTTP handlers
- `internal/logic/` — Business logic
- `internal/repo/` — Data access (SQL)
- `internal/db/` — Database setup and migrations
- `web/` — HTML templates and static assets (JS, CSS)

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

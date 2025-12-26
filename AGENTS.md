# Repository Guidelines

## Project Structure & Module Organization
- `cmd/` contains entrypoints for binaries: `ninete` (API/server), `migrate` (DB migrations), and `task` (one-off jobs).
- `internal/` holds application packages (db, logic, serve, repo, task) and helpers; these are not meant for external import.
- `data/db/` stores SQLite files for dev/test; the Makefile creates `data/db/dev` and `data/db/test`.
- `build/` is generated output for compiled binaries.

## Build, Test, and Development Commands
- Prefer Makefile targets for builds, tests, migrations, and linting; avoid running Go tools directly.
- `make build`: builds the `ninete` binary to `build/dev`.
- `make dev`: builds and runs the server with `ENV=development`.
- `make migrate`, `make migrate-down`, `make migrate-status`: run migrations using the `migrate` binary.
- `make migrate-create name=add_users`: create a new migration file.
- `make seed`: seed the database.
- `make test` / `make test-verbose`: run Go tests with `ENV=test` and a fresh test DB.
- `make lint` / `make lint-fix`: run `golangci-lint`.

## Coding Style & Naming Conventions
- Go 1.25.1 module: `github.com/ad9311/ninete`.
- Use `gofmt`-formatted Go code and standard Go naming (Exported identifiers in `CamelCase`, packages in lowercase).
- Prefer small, focused packages in `internal/` with clear boundaries (db, repo, logic, serve).

## Testing Guidelines
- Tests live alongside code under `internal/` and use the standard `*_test.go` pattern.
- Run tests via Make targets only; do not invoke `go test` directly.
- Run the full suite with `make test`; use `make test func=TestName pkg=./internal/serve` to scope.
- Test DB setup/teardown is handled by the Make targets; do not create or manage it manually.

## Commit & Pull Request Guidelines
- Commit messages are short, imperative, and descriptive (e.g., "Fix broken query for expense").
- PRs should describe the change, list any new commands or env vars, and include test results.

## Configuration Tips
- `.env` is loaded by the Makefile when present; keep secrets out of version control.
- `ENV=development` and `ENV=test` are used by the binaries to select config and DB paths.

# Deployment

How NINETE runs in production. The app is deployed on a single Linux VPS with a
standard FHS layout — no containers, no orchestration. One host, one app, one user.

This document covers the parts of the deployment that constrain how code is
written. Host-specific details — paths, account names, the hostname, scheduled
jobs, and the current list of known gaps — live in `docs/deployment.local.md`,
which is git-ignored and exists only on the maintainer's machine and the host.

## Shape of the deployment

The host holds a git clone of this repository and builds from it. There is no
cross-compilation and no artifact shipped from a developer machine; everything is
built on the VPS from the checked-out source. CGO is required
(`mattn/go-sqlite3`), so the host needs a working `gcc`.

Three binaries are built: `ninete`, `migrate`, and `task`. Build output lands in a
directory private to the service account, and the finished `ninete` binary is then
copied to a root-owned location on the system path — which is what the systemd
unit executes. Rebuilding therefore never disturbs the running process: the swap
happens in one `cp`, immediately before the restart.

Caddy terminates TLS and reverse-proxies to the app on loopback. The app listens
only on `PORT` and is never exposed directly. Certificates are issued and renewed
automatically.

## Environment

`prog.Load()` (`internal/prog/prog.go`) **skips `.env` entirely when `ENV=production`.**
It reads the process environment and nothing else. That means:

- The service gets its configuration from systemd's `EnvironmentFile`.
- The `migrate.sh` and `task.sh` wrappers `source` that same file explicitly before
  exec'ing their binary, because a shell invocation has no systemd to do it for them.
- There is no `.env` file on the production host and there must not be one. Adding
  one would have no effect and would only create a second, silently-ignored source
  of truth.

**A new config value must be added to the systemd environment file on the host**,
not to a `.env`. See `.env.example` for what each key does.

`MAX_IDLE_CONNS` and `MAX_OPEN_CONNS` are present but **empty** in production, which
is deliberate. `prog.SetInt` treats an empty value as unset and falls back to
`DefaultMaxOpenConns` / `DefaultMaxIdleConns` — both `1` (`internal/db/db.go`). A
single connection is the intended production configuration; see the Performance
Priorities section of `CLAUDE.md` for why raising it is not an improvement here, and
what would have to change first if it were ever raised anyway.

`GO_BUILD_ENVS` is *not* set in production. `build.sh` sets its own build flags
inline (`CC=gcc`, not the `musl-gcc` from `.env.example`).

### Environment stamping

`ENV` and `DATABASE_URL` are independent variables, so nothing structurally
prevents a development command from opening the production database.
`verifyEnvStamp` (`internal/db/stamp.go`) guards against it: the owning environment
is written into the SQLite header's `application_id` on first open, and `db.Open`
fails when a later open disagrees.

The production database is already stamped. Do not re-stamp it.

## Working directory

The unit is `Type=simple`, `Restart=always`, running as an unprivileged service
account with `WorkingDirectory` set to the repository checkout.

**The working directory is load-bearing.** `internal/serve/template.go` and
`internal/serve/routes.go` resolve templates and static assets through relative
paths — `./web/views/**/*.html` and `./web/static/`. The app only finds its own
frontend when started from the repository root. Running the binary from anywhere
else produces a server that boots and then fails to render.

Do not introduce more relative-path dependencies without noting it here.

## Deploying

A single script on the host runs the whole procedure, in order:

1. `pull.sh` — `git pull --ff-only`. Fails rather than merging, so a dirty or
   diverged checkout stops the deploy instead of resolving itself.
2. `build-js.sh` — `bun install` then `make build-static-js`. The JS bundle is
   git-ignored, so it must be built on the host.
3. `build.sh` — builds `migrate`, `task`, and `ninete` with
   `CGO_ENABLED=1 CC=gcc -trimpath -ldflags=-s -ldflags=-w -buildvcs=false`, then
   re-applies restrictive permissions to the output directory.
4. `migrate.sh up` — applies pending migrations.
5. Copy the new binary into place as root and restart the unit.

Every script refuses to run as root (`$EUID` check at the top).

Note the ordering: **migrations are applied while the old binary is still serving.**
For this app that is fine — it is single-user and migrations are additive — but a
migration that removes or renames something the running binary still reads will
error for the few seconds before the restart lands.

### Individual scripts

Each step is also runnable on its own: `pull.sh`, `build-js.sh`, `build.sh`,
`migrate.sh <cmd>`, `task.sh <name>`. The last two print help when given no
arguments. The scripts live on the host and are **not under version control**.

## Database

Single SQLite file, WAL mode. The `-wal` file settles around 4 MB because
`wal_autocheckpoint = 1000` pages at a 4096-byte page size
(`internal/db/init/connection.sql`, `internal/db/init/database.sql`) — a WAL much
larger than the database is expected here, not a symptom.

Migration commands always go through the `migrate.sh` wrapper so the env file is
loaded: `status`, `up`, `down` (one step).

### Backups

If adding backups, use SQLite's own backup API rather than copying the file — a
plain `cp` of a database with a live WAL can capture a torn state:

```bash
sqlite3 <db-path> ".backup '<backup-path>'"
```

## Tasks

Available tasks are registered in `cmd/task/main.go`:

- `copy_due_recurrent_expenses` — materializes due recurrent expenses into real
  expenses (`internal/task/task.go`, `CopyDueRecurrentExpenses`). Run on a schedule.
- `create_invitation_code` — interactive, prompts on stdin. Run by hand.

## Rollback

There is no automated rollback. To revert: check out the previous commit in the
host's clone, re-run `build.sh`, copy the binary into place as root, and restart
the unit.

The full deploy script cannot be used for this — its first step is
`git pull --ff-only`, which would drag the checkout straight back to the tip of the
branch.

Two caveats. `git checkout <commit>` leaves a detached HEAD, so the next `pull.sh`
fails until you `git checkout main`. And **migrations are not rolled back by this
procedure** — if the bad deploy applied one, decide deliberately whether the old
binary tolerates the new schema, or run `migrate.sh down` first.

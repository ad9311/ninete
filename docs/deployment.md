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

The listener binds `127.0.0.1` unless `HOST` is set to something else (an empty
`HOST` falls back to the default too). That default is a security boundary, not a
convenience: the root middleware chain runs `realClientIP`, which identifies the
client from the **last** `X-Forwarded-For` entry, and the login rate limiter
buckets on the result. Caddy appends the address it saw, so the last entry is the
one Caddy wrote and everything before it is whatever the client sent. A directly
reachable port would let a client append that last entry itself and get a fresh
rate-limit budget per request.

`realClientIP` deliberately ignores `True-Client-IP` and `X-Real-IP`. Caddy's
`reverse_proxy` sets neither, so both would arrive verbatim from the client;
chi's `middleware.RealIP` trusts them and must not be substituted here. **Do not
set `HOST` to `0.0.0.0` without putting a proxy in front that overwrites, rather
than appends to, `X-Forwarded-For`.**

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
single connection is the intended production configuration; see `docs/performance.md`
for why raising it is not an improvement here, and what would have to change first if
it were ever raised anyway.

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

## Service sandbox

The unit runs under systemd's filesystem sandbox: `ProtectSystem=strict` makes the
entire filesystem read-only for the process, and a single `ReadWritePaths=` entry
punches the database directory back open. **The app process can write to exactly
one directory — the one holding the SQLite file — and nowhere else.**

That is a constraint on code, not just on the host:

- Anything the app writes at runtime must live beside the database. Writing a
  temp file, a cache, an upload, or an export to the working directory, `/tmp`,
  or the service account's home fails with `EPERM`. `PrivateTmp=yes` also gives
  the service its own `/tmp`, so nothing written there is visible from a shell.
- The working directory — the repository checkout — is **read-only** to the running
  app. Reading templates and static assets from it still works, which is all the
  app needs. Nothing may write there.
- `UMask=0077` means files the app creates are `0600`, owned by the service account.
- The capability bounding set is empty, so the app cannot bind a privileged port.
  It binds `PORT` on loopback and Caddy owns 80/443; do not change that expecting
  the app to listen on 443 directly.
- A syscall filter (`@system-service`) and `MemoryDenyWriteExecute=yes` are in
  effect. Both are fine for a Go binary — no JIT, no exotic syscalls — but a
  future dependency that needs one would fail at runtime rather than at build time.

The exact stanza lives in `docs/deployment.local.md`.

### What the sandbox does not cover

The sandbox is a property of the long-running web process, not of the app as a
whole. `migrate.sh` and `task.sh` run as plain shell processes outside the unit,
so **none of the constraints above apply to them**. This is deliberate and settled
— the reasoning is in `docs/deployment.local.md` — but it has one consequence for
code:

Something invoked through `task.sh` can write wherever the service account can
write. The same logic called from an HTTP handler gets `EPERM` outside the
database directory. So a task that ran clean in production proves nothing about
the same code reached from a request, and "it worked when I ran the task" is not
evidence the write path is legal.

Nothing diverges today — `internal/task/task.go` does database work only, and
migrations are `go:embed`ed, so `goose.Up` never touches the filesystem. The first
task that needs to write a file (an import, an export dump, a generated report) is
the point at which this stops being trivia: put its output beside the database, or
raise it with the owner before assuming a directory is writable.

## Deploying

A single script on the host runs the whole procedure, in order:

1. `pull.sh` — `git pull --ff-only`. Fails rather than merging, so a dirty or
   diverged checkout stops the deploy instead of resolving itself.
2. `confirm_version` (in `deploy.sh`) — prints the version this deploy will stamp,
   and asks for confirmation when the checkout is dirty or `HEAD` is not tagged.
   Placed here deliberately: the version is a property of the code `pull.sh` just
   fetched, and declining before `build.sh` and `migrate.sh` costs nothing but an
   advanced checkout. See **Versioning** below.
3. `build-js.sh` — `bun install` then `make build-static-js`. The JS bundle is
   git-ignored, so it must be built on the host.
4. `build.sh` — builds `migrate`, `task`, and `ninete` with
   `CGO_ENABLED=1 CC=gcc -trimpath -buildvcs=false` and an `-ldflags` argument
   carrying `-s -w` plus the version stamp (see **Versioning** below), then
   re-applies restrictive permissions to the output directory.

   The flags are held in bash arrays rather than strings. `-ldflags` is a single
   flag taking one quoted argument, and a string expanded unquoted for word
   splitting cannot carry one — passing `-ldflags=-s -ldflags=-w` instead means Go
   keeps only the last, silently dropping `-s`. That was the case until
   2026-08-17 and cost ~1.1 MB of symbol table in every production binary.
5. `migrate.sh up` — applies pending migrations.
6. Copy the new binary into place as root and restart the unit.

Every script refuses to run as root (`$EUID` check at the top).

Note the ordering: **migrations are applied while the old binary is still serving.**
For this app that is fine — it is single-user and migrations are additive — but a
migration that removes or renames something the running binary still reads will
error for the few seconds before the restart lands.

### Individual scripts

Each step is also runnable on its own: `pull.sh`, `build-js.sh`, `build.sh`,
`migrate.sh <cmd>`, `task.sh <name>`. The last two print help when given no
arguments.

The scripts live in this repository under `scripts/`, and the host reaches them
through a symlink so the path it invokes is stable. Three consequences for anyone
editing them:

- **They deploy themselves.** `pull.sh` updates the scripts along with the code,
  so a change to `deploy.sh` takes effect on the *next* deploy, not the one that
  pulls it.
- **Each body is wrapped in `main()`, and the last line is `main "$@"; exit`.**
  Both halves are load-bearing, not style. Bash parses one command at a time and
  seeks back to just past it before executing, and step 1 of a deploy rewrites the
  very file that is running. The wrap gets the body parsed in one piece, so the
  work itself is safe. It is not sufficient on its own: on return from `main`,
  bash seeks to the old offset of that final line and keeps reading, so a
  rewritten file that grew hands it a fragment of new content to execute. The
  `exit` prevents that read — but only while it shares the line, because an
  `exit` on a line of its own sits in the old bytes bash never reaches.
- **Paths into the checkout are derived, not hardcoded**, via
  `cd -P "$(dirname "${BASH_SOURCE[0]}")"`. `-P` is required: the host invokes them
  through the symlink, and without it the derived parent directory is wrong.

They are linted: `make lint-sh` runs shellcheck over `scripts/*.sh`, and `make lint`
and `make lint-fix` both end by calling it. These scripts have no test coverage and
a failure only surfaces mid-deploy, so the linter is the only check standing between
an edit and production. It skips with a warning when shellcheck is not installed,
so an editor without it silently loses that check — CI installs a pinned version
(`v0.11.0`, matching the pin in `.github/workflows/linters.yml`) and always runs it.

Paths *outside* the checkout (the build output directory, the env file, the
installed binary) stay literal — they are host facts, not repository facts, and
one of them is matched verbatim by the sudo policy. See `deployment.local.md`
before changing any of those strings.

## Database

Single SQLite file, WAL mode. The `-wal` file settles around 4 MB because
`wal_autocheckpoint = 1000` pages at a 4096-byte page size
(`internal/db/init/connection.sql`, `internal/db/init/database.sql`) — a WAL much
larger than the database is expected here, not a symptom.

Migration commands always go through the `migrate.sh` wrapper so the env file is
loaded: `status`, `up`, `down` (one step).

### Backups

Production is backed up nightly to off-site object storage on a systemd timer.
Host specifics — paths, credentials, retention, and the restore procedure — are
in `deployment.local.md`.

The one rule that constrains code: never snapshot the database with `cp`. A plain
copy of a database with a live WAL can capture a torn state. Use SQLite's own
mechanism, which the backup script does:

```bash
sqlite3 <db-path> "VACUUM INTO '<snapshot-path>'"
```

`VACUUM INTO` writes a consistent, compacted copy — expect the snapshot to be
smaller than the source. `.backup '<path>'` is equally safe if you want a
byte-for-byte copy instead.

A restored database arrives with the copying process's umask rather than the
original's, so modes need re-checking after any restore.

## Tasks

Available tasks are registered in `cmd/task/main.go`:

- `copy_due_recurrent_expenses` — materializes due recurrent expenses into real
  expenses (`internal/task/task.go`, `CopyDueRecurrentExpenses`). Run on a schedule.
  Each copy carries the recurrent expense's tags and bumps its occurrence
  counter; when that counter reaches a non-zero `occurrence_limit` the row
  archives itself in the same statement and drops out of the task's selection
  until someone unarchives it from `/recurrent-expenses/archived`. A run that
  copies fewer rows than the month before is therefore expected, not a fault.
  One failing row is logged and skipped, and the task still exits 0 — check the
  count in the log line, not just the exit status.
- `create_invitation_code` — interactive, prompts on stdin. Run by hand.
- `test` — a no-op hook for development. Not for production use.

## Versioning

Every binary carries the identity of the commit it was built from. Three
variables in `internal/prog/version.go` — `Version`, `Commit`, `BuildTime` — are
set at link time with `-X`, by `scripts/build.sh` on the host and by the Makefile
locally. Both derive the values from git:

```
Version    git describe --tags --always --dirty    v0.1.0-7-gdf3dbdc-dirty
Commit     git rev-parse --short HEAD              df3dbdc
BuildTime  date -u +%Y-%m-%dT%H:%M:%SZ             2026-08-22T14:58:37Z
```

`Version` is the primary identifier and needs no bookkeeping: the nearest tag,
how many commits have landed since it, the commit itself, and `-dirty` when the
checkout had uncommitted changes. An untagged commit still gets an exact
identity, so there is no state to forget to bump and nothing that can go stale.

Tags are SemVer-shaped (`v0.1.0`) and cut by hand when a release feels like one,
with `git tag -a` on `main` followed by `git push --tags`. They are a convenience
for reading `Version`, not a compatibility contract — this app has one user and
no consumers. **Schema compatibility is tracked separately** by `PRAGMA
user_version` in the migrations; do not conflate the two.

### Release order

Tag before deploying, not after:

```
git checkout main && git pull
git tag -a v0.1.0 -m "..."      # on the merge commit
git push --tags
<deploy on the host>
```

`pull.sh` runs `git pull --ff-only`, and git auto-follows tags that point into the
history it fetches, so a tag pushed beforehand arrives with the commit. Tag
afterwards and the host will not see it until some later fetch — the binary is
identical either way, only the label it reports differs.

Not every deploy needs a tag. An untagged commit still gets an exact identity
(`v0.1.0-3-gabc1234`); that is the point of deriving from `git describe`.

`deploy.sh` prints the version it is about to build and stops for confirmation
when the stamp says something is off:

- the checkout has uncommitted changes (the build would be stamped `-dirty`)
- `HEAD` is not tagged

Neither is an error, so it asks rather than refusing — untagged deploys are the
normal case. `deploy.sh --yes` (or `-y`) skips the prompt, and a run with no
terminal prints the warnings and continues rather than blocking, so a timer or a
piped invocation cannot hang on it. Declining exits 1 before anything is built or
migrated.

A `-dirty` version in production means the running binary matches no commit.
Investigate rather than ignore it.

Reading the version:

- The web binary logs it on boot, on the `Booting up application...` line.
- All three binaries answer `version` as a subcommand
  (`/usr/local/bin/ninete version`, `migrate version`, `task version`). It runs
  before any config or database access, so it still answers on a host whose env
  file or database is broken.
- `make version` prints what the current checkout *would* build, without building.

Two things that keep this honest:

- **`-buildvcs=false` stays.** Go's own VCS stamping is off and these `-X` values
  are the only source of build identity. Turning `buildvcs` on would add a second,
  differently-derived answer to the same question.
- **Unstamped builds are legal.** The defaults (`dev`, `unknown`, `unknown`) mean
  `go test`, `go run`, and a hand-typed `go build` all work with no flags. Nothing
  may depend on the stamp being present — it is diagnostic, never load-bearing.

## Rollback

There is no automated rollback, and no archive of previous binaries — `deploy.sh`
overwrites `/usr/local/bin/ninete` in place. To revert: check out the previous
commit in the host's clone, re-run `build.sh`, copy the binary into place as root,
and restart the unit.

The version stamp helps you decide *what* to revert to, not how: read the running
build with `ninete version` or from the boot log, then pick the commit or tag to
check out. It does not shorten the procedure below.

The full deploy script cannot be used for this — its first step is
`git pull --ff-only`, which would drag the checkout straight back to the tip of the
branch.

Two caveats. `git checkout <commit>` leaves a detached HEAD, so the next `pull.sh`
fails until you `git checkout main`. And **migrations are not rolled back by this
procedure** — if the bad deploy applied one, decide deliberately whether the old
binary tolerates the new schema, or run `migrate.sh down` first.

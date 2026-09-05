# Performance

The goal is an app that feels instant for one person, not one that sustains
throughput for many. When those two goals conflict, choose responsiveness. The
reasoning behind that constraint is in the Project Scope section of `CLAUDE.md`.

## Worth the effort

- Reduce how many queries a page costs, and keep each one index-backed. Any
  user-scoped query bounded by a timestamp should have a matching
  `(user_id, <time column>)` index; `internal/db/index_test.go` pins the
  important ones with `EXPLAIN QUERY PLAN`.
- Batch lookups instead of one query per row. `repo.SelectTagRows` is the
  pattern: pass a slice of ids, get rows back, group them in memory with
  `repo.TagNamesByTargetID`.
- Keep work off paths that should be free. `/static/*` is mounted outside the app
  middleware chain specifically so serving an asset never loads a session or
  queries the database.
- Remove anything unbounded in rows, memory, or SQL parameters. SQLite rejects a
  statement with more than 32766 parameters, which is why `SelectTagRows` batches
  its `IN (...)` list.

## Not worth the effort

Concurrency capacity work — connection-pool tuning, read/write pool splits,
caching layers, queues — buys nothing here. `MAX_OPEN_CONNS` defaults to 1
(`db.DefaultMaxOpenConns`) and that is fine: one user produces almost no
overlapping requests, so serializing them costs microseconds nobody can
perceive.

### JSON encoding, and `encoding/json/v2`

Measured 2026-09-04, on the Go 1.27 upgrade. Do not reopen this without a
measurement that contradicts it.

Go 1.27 reimplements `encoding/json` on top of the v2 engine by default —
`go list -f '{{.GoFiles}}' encoding/json` selects the `v2_*.go` files, and only
`GOEXPERIMENT=nojsonv2` brings the old implementation back. **The speedup
therefore arrived with the toolchain bump, with no code change.** On 2000
expense-shaped structs, encoding went from 879 µs and 2001 allocations to 670 µs
and 2; decoding roughly halved.

Rewriting call sites onto `encoding/json/v2` directly buys close to nothing on
top of that: it is the same engine underneath. The v2 API's real advantages —
`omitzero`, case-sensitive field matching, explicit options — are semantic, and
nothing here needs them.

There are three JSON sites outside tests, and none of them is near a bottleneck:
`handlers/api.go` (API responses, one page of rows at a time),
`handlers/handle_exports.go` (the export, the only one with volume, at 0.67 ms
for 2000 rows), and `serve/manifest.go` (one `Unmarshal` at startup).

The engine swap is a behavior surface as well as a speed one, since these
responses feed the SPA. The Go suite asserts payload shape across the handler
tests and passes on 1.27, which is the evidence that it held. If a
JSON-shaped bug ever does appear, `GOEXPERIMENT=nojsonv2` at build time is the
bisect switch — worth knowing it exists, not worth setting.

## If a pool larger than 1 is ever introduced anyway

Three things would matter:

- The PRAGMAs in `internal/db/init/connection.sql` are already applied to every
  connection through the driver connect hook, so those are fine.
- `WithTx` opens a deferred transaction, which can fail with
  `SQLITE_BUSY_SNAPSHOT` under concurrent writers — `_txlock=immediate` would be
  needed.
- `db.Optimize` runs `PRAGMA optimize` on whichever single connection
  `database/sql` hands it at shutdown; the statistics it refreshes come from that
  connection's own query history, so with several connections in play it would
  only ever see a fraction of the workload.

## Measuring

- Query shape: `EXPLAIN QUERY PLAN`, asserted in tests rather than eyeballed.
- Wall time: the app logs every repo query with its duration outside `ENV=test`
  (`prog.Logger.Query`), so `make dev` shows what a page actually costs.

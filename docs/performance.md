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

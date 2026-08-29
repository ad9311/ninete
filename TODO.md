# TODO

Known issues and follow-up work that is deliberately out of scope for the change
that surfaced it. Remove an entry once it is fixed.

## No endpoint currently needs a raw unique-constraint mapped to 422 — keep it that way

Every table with a unique index lets SQLite's `UNIQUE constraint failed: ...`
propagate untouched out of the logic layer if nothing catches it first. The
API's `WriteAPIError` won't recognize that string, so an uncaught collision
falls through to a generic `500` — correct about not leaking the driver's
message, wrong about whose fault it is.

This was a real gap against `Store.CreateFood`/`UpdateFood`, which Phase 0B
deleted. Checked again after the SPA migration (2026-08-29): every live path
that could still hit one of these constraints already avoids the raw case, each
for its own reason, not by a shared fix —

- `POST /api/register` is the only reachable plain insert on a unique column
  (`users.email`/`username`). `SignUp` (`internal/logic/logic_auth.go`) calls
  `repo.IsUniqueViolation(err)` and maps it to `ErrAccountExists`, answered as
  a `422` via `WriteAPIError`'s `userErrors` list.
- Tags are only ever created through `INSERT OR IGNORE`
  (`ensureTagsForUserTx`, `internal/logic/logic_tag.go`) — a collision reuses
  the existing tag instead of erroring. `Store.CreateTag`, the plain-insert
  version that could hit the constraint, has no HTTP route.
- Categories have no create/update API route at all.
- Expense budgets upsert (`INSERT ... ON CONFLICT DO UPDATE`,
  `internal/repo/expense_budget.go`), so a "collision" just updates the row.

So there is nothing to fix today. Kept as a note rather than deleted: the
pattern to reuse, if a future create endpoint does a plain insert on a unique
column, is `repo.IsUniqueViolation` + a sentinel in `internal/logic/errs.go`
named in that endpoint's `userErrors` — the same shape `SignUp` already uses.
Don't add a shared abstraction on top of it pre-emptively; reject-and-report
(registration), silently-reuse (tags), and upsert (budgets) are different
conflict semantics, not the same code three times.

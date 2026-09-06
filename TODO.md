# TODO

Known issues and follow-up work that is deliberately out of scope for the change
that surfaced it. Remove an entry once it is fixed.

## Quick-add still parses a day-level billed date

The billed date is picked and displayed as a month everywhere else: the expense
form is an `<input type="month">` writing day 01, and `formatMonthUTC` renders
`Sep 2026`. Quick-add did not move with it. `parseQuickDate`
(`internal/logic/logic_quick_expense.go`) still accepts `today`, `yesterday`,
`tomorrow`, `next month`, `12 July 2026` and `12/07/2026`, and
`QuickAddForm.svelte`'s help popover still advertises all of them.

Nothing is broken: whatever day it resolves lands in the right month, which is
the only part the app shows or filters on. It is incoherent rather than wrong —
the field asks for a precision the app then discards, and `tz_offset` is sent
on every quick-add solely to resolve a relative day that no longer matters.

Deliberately out of scope for the change that made the billed date a month
(#146), which was about display and search. Deciding what to do here is a
product call, not a cleanup: narrowing the grammar to months would break muscle
memory for `today`, and the shortest input is the whole point of quick-add.
Options, roughly in increasing order of disruption — keep parsing days and
silently snap to the 1st; keep the grammar but add month forms (`Sep 2026`,
`this month`); or drop day forms entirely and retire the `tz_offset` field with
them.

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

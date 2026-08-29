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

## An expired session on an `/api/*` link fails silently instead of redirecting

`web/app/routes/exports/Index.svelte` points a plain `<a download>` at
`/api/exports/expenses.json`. That path runs `apiAuth`, which answers `401` with
the JSON envelope and no `Location` — correct for the API chain, which must never
redirect. But a browser honours `download` whatever the status, so a user whose
session lapsed while the SPA sat open clicks Download and silently saves an
`expenses.json` holding `{"error":"…"}`. No sign-in prompt, no visible failure.

The deployed version does the right thing: `/account/exports/expenses.json` goes
through `AuthMiddleware` and redirects to `/login`. Phase 6 settled the general
case — `lib/api.ts` recognises the `401` and sends the user to `AppLoginPath` —
but the export link is the one call that does not go through it, because §5 of
`docs/spa-migration.md` keeps export links plain anchors and a plain anchor is a
browser navigation, not a `fetch`.

So what is left is this anchor alone, and every future `<a download>` at an
`/api/*` path. Whoever fixes it has to keep the plain-anchor decision or get it
revisited: a `HEAD` or `/api/session` probe before letting the click through is
the cheap version, and Phase 7 is where the honest fix lives, since moving the
SPA to `/` lets the download hang off a route that redirects like the page chain
does.

# TODO

Known issues and follow-up work that is deliberately out of scope for the change
that surfaced it. Remove an entry once it is fixed.

## A unique-name collision will read as a server fault on the API

Every table with a unique index lets SQLite's `UNIQUE constraint failed: ...`
propagate untouched out of the logic layer. The pages render that string; the
API's `WriteAPIError` deliberately will not, so an unnamed collision becomes a
generic `500` instead — correct about not leaking, wrong about whose fault it is.

The case this was first written against was `Store.CreateFood`/`UpdateFood`,
which Phase 0B deleted. It survives on the tables expenses still uses: `tags`
is unique per `(user_id, lower(name))`, and `categories` on `lower(name)`.

Whoever ports the first endpoint that can hit one — Phase 2 of
`docs/spa-migration.md` — should turn the collision into a sentinel in
`internal/logic/errs.go`, name it in the endpoint's `userErrors`, and let it
answer `422` with a message that does not quote the constraint. Fixing it in the
logic layer improves the existing pages too.

## A nested validation failure publishes the inner struct's field name

`ValidationError.Fields` (`internal/logic/logic.go`) is keyed by
`validator.FieldError.Field()`, which is the leaf Go field name — it says nothing
about which struct the field belongs to. Every `ValidateStruct` call a use-case
makes on a *secondary* params struct therefore lands in the same flat map as the
request's own fields.

The reachable case today is tags. `ensureTagsForUserTx` (`internal/logic/logic_tag.go`)
validates one `TagParams{Name: name}` per tag, and `normalizeTagNames` does not
truncate, so `POST /expenses` (or a recurrent expense) carrying a
tag longer than 20 characters fails with `[Name:max]` — `fields` comes back as
`{"name": "max"}`. None of those payloads has a `name` key; the offending value
arrived under `tags`. A client that highlights inputs by field key marks nothing,
and on an endpoint that *does* own a `name` field it would mark the wrong input.

The same shape would appear for any future nested params struct sharing a field
name with its parent, where the two entries would silently overwrite each other.

Phase 2 should decide how a nested failure names itself — map the tag case onto
`tags`, or key nested entries by a path rather than the leaf name — before the
first endpoint that accepts tags is ported.

## `SelectTagsForTaggable` takes its owner table as a plain string

`SelectTagsForTaggable` (`internal/repo/tagging.go`) interpolates its
`ownerTable` argument straight into the `INNER JOIN` of
`selectTagsForTaggableBase` with `fmt.Sprintf`. Every caller today passes a
literal — `"expenses"` and `"recurrent_expenses"` — so nothing is
wrong at runtime, but the safety is a convention rather than something the type
system or a whitelist enforces, unlike every other query in the package, where
`QueryOptions` validates column names against `validExpenseFields()` and friends.

The gosec exclusion added for `internal/repo` in `.golangci.yml` covers this
call site along with the `QueryOptions` ones, so a future caller that passed a
non-literal here would not be flagged.

Give `ownerTable` a defined type with constants beside `TaggableTypeExpense`,
so the pairing of taggable type and owner table is expressed once and a caller
cannot supply an arbitrary string.

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

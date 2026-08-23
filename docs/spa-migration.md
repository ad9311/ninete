# Plan — Migrate NINETE to a Svelte SPA

Status: approved 2026-08-22. Scope reduced 2026-08-23 — read §0 first. Phases 0.1–0.3 landed.
Audience: the maintainer reading it once, and an agent picking up any single phase later
without the conversation that produced it. Section 7 records the decisions already made — do
not re-litigate them.

**Starting point.** `main` has no Svelte tooling at all. A throwaway spike on
`experiment/svelte-api` (commit `84b9d4a`) proved the pieces work together — Svelte 5 under
`bun-plugin-svelte`, mounted from a Stimulus controller, reading a Go JSON endpoint — and the
facts this plan states about the bundler were verified there. Treat that branch as a reference
to read, not a branch to merge: Phase 0 re-lands only the build tooling, deliberately, without
the experiment route.

**Before anything else, read §0 (Scope reduction).** Macros, foods and mood entries are being
dropped from the app. Roughly a third of the inventory this plan was originally sized against
is not going to be ported, and the phase order changed because of it.

**Before writing any view code, read §3.6 (Dates).** It is the one section where a mistake is
invisible in review, invisible under the default `TZ=UTC` test run, and wrong for the user.
Expenses are the app's core feature and every one of them hangs off a calendar date.

**Freeze rule while this runs:** no new features and no behavior changes to existing views.
Every phase is a like-for-like port. If a view looks wrong after porting, the bug is in the
port, not in the design — that is the only reason this migration is checkable at all.

---

## 0. Scope reduction — macros, foods and moods are being dropped

Decided 2026-08-23, after Phase 0.3 landed. **The app keeps expenses and what hangs off them.**

| Kept | Dropped |
| --- | --- |
| Expenses, recurrent expenses, expense budgets, tags, categories, dashboard, account, delete-data, exports, auth | Macros (entries, goals, stats), Foods, Mood entries |

This is not a migration decision — those features are going away whether or not the SPA
happens. It is recorded here because it changes what the migration has to port, and because
doing it in the wrong order wastes a phase.

**Removal happens in two steps, deliberately split.**

1. **Code removal — Phase 0B, before any resource is ported.** Routes, handlers, logic, views,
   Stimulus controllers, nav links, and the macro half of the dashboard.
2. **Table removal — Phase 8, at the very end.** `foods`, `macro_entries`, `macro_goals`,
   `mood_entries`, and the `taggings` rows pointing at mood entries. Irreversible, and gated on
   an export.

**Why the code cannot wait for the cleanup phase.** Phase 2's pilot resource *was* Foods, and
Phase 4 ported macros and moods with their Chart.js views. Deferring removal means porting 18
of 47 views and 6 of 18 controllers into a codebase that then deletes them — and worse, setting
the conventions every other resource copies on a resource nobody will maintain. Phase 7's
catch-all route cannot coexist with template-only routes either, so "the cleanup phase" was
never actually available: Phase 7 was the real deadline.

**Why the tables can wait, and must.** Dropping them destroys the macro and mood history, and
`/account/exports/expenses.json` exports expenses only — there is no backup path for that data
today. Keeping the tables through the migration costs nothing: the `internal/repo` files stay
compiled with no callers, and `TestColumnConstantsMatchSchema` keeps guarding the schema.
**Before Phase 8 drops them, export that history to a file and confirm with the owner that it
is kept.** If the data is not wanted, say so explicitly in Phase 8's PR rather than letting it
go silently.

**Freeze-rule exception.** The freeze rule above forbids behavior changes to existing views.
Phase 0B is its one deliberate exception, and it is a deletion rather than a change. After it,
the freeze rule applies again in full to everything left.

**What this removes, counted** — so §2's inventory can be checked rather than trusted:

| | Before | Dropped | After |
| --- | --- | --- | --- |
| Template files | 47 | 18 | 29 |
| Stimulus controllers | 18 | 6 | 12 |
| Routes | 67 | 32 | 35 |
| Handler test files | 14 | 4 | 10 |

---

## 1. Goal and end state

Today: Go `html/template` renders every page, Turbo intercepts navigation and form posts,
Stimulus adds per-page interactivity. 47 template files, 18 Stimulus controllers, 67 routes —
29, 12 and 35 once Phase 0B removes the dropped features (§0). Every count below this line is
the post-§0 one.

End state:

- Go serves one HTML shell plus a JSON API. No per-view templates.
- Svelte 5 owns rendering, routing and client state.
- Turbo and Stimulus are gone, along with `@hotwired/*` from `package.json`.
- Session cookie auth, CSRF, and the CSP survive unchanged in shape — see §3.

Non-goals: server-side rendering of components, SvelteKit (Go serves the HTML), offline
support, multi-user features, any visual redesign, and reviving macros, foods or moods in any
form (§0).

---

## 2. Inventory — everything that must be migrated

Nothing below is optional. A phase is done when its rows are ticked.

### 2.1 Views (29 files after §0, `web/views/`)

Macros, foods and mood entries are absent from this table on purpose: Phase 0B deletes them
(§0). If you are looking at a row that is not here, it is not being ported.

| Resource | Files | Notes |
| --- | --- | --- |
| `layout.html` | 1 | Becomes the shell. Last thing to change, §Phase 7 |
| `common/` | `_csrf`, `_footer`, `_form_buttons`, `_form_error`, `_header`, `_pagination` | Become shared components |
| `login`, `register` | 2 | Guest routes. Special: CSRF/session rotation, §3.2 |
| `dashboard` | 1 | Expense summary only; the macro half goes in Phase 0B |
| `expenses` | `index`, `new`, `edit`, `show`, `stats`, `budgets`, `_form`, `_quick_form` | Largest resource. Search, filters, sort, pagination |
| `recurrent_expenses` | `index`, `archived`, `new`, `edit`, `show`, `_form` | |
| `account`, `delete_data`, `exports` | 3 | Destructive posts; file download |
| `error`, `not_found` | 2 | Become client-side error states + a server fallback |


### 2.2 Stimulus controllers (12 after §0, `web/static/js/controllers/`)

Each becomes either component-local state or a shared util. None survive as controllers.

`amount`, `chart`, `dashboardDate`, `date`, `dateHelp`, `filter`, `localDate`, `nav`,
`quickExpense`, `searchPanel`, `sort`, `theme`.

Deleted unported in Phase 0B: `macroCalc`, `macroDate`, `macroSelect`, `macroTrend`,
`moodChart`, `submitOnChange`. The last one is the trap — `submitOnChange` reads as a generic
utility but its only two uses are `macros/index` and `macros/stats`, so it goes with them;
`filter` is the equivalent control expenses keeps.

Two of the survivors need decisions rather than a port:
- `theme` — writes `localStorage`, and `layout.html:9-22` has an inline anti-FOUC script that
  must stay server-rendered even in the SPA.
- `chart` — a Chart.js instance; in Svelte it becomes a component with `$effect` lifecycle.
  After §0 its only remaining use is `expenses/stats`, so Chart.js arrives in Phase 3 rather
  than Phase 4. It stays as a dependency; it is not the problem being solved here.

### 2.3 Server-side behavior that has no client equivalent yet

| Behavior | Where | What replaces it |
| --- | --- | --- |
| Auth redirect `303 → /login` | `internal/serve/middleware.go:41` | `401` JSON on `/api/*`, §3.1 |
| CSRF token in every form | `common/_csrf.html`, `nosurf` at `middleware.go:79` | `X-CSRF-Token` header, §3.2 |
| Error re-render preserving input | `render.go:renderErr` | Client keeps form state; API returns `422` + field errors |
| `tz_offset` query injection | `web/static/js/index.ts:11-17` (a Turbo event) | Retired for listings — client computes explicit bounds, §3.6. Quick-add still needs a client zone, sent explicitly |
| Pagination / sort / filter via query string | `handlers/shared.go`, `expense_search.go` | Client router owns the query string, API reads it unchanged |
| Icons rendered on `turbo:load` | `index.ts` + `icons.ts` | Per-component icon rendering |
| Loading spinner | Turbo's `.turbo-progress-bar`, restyled in `layout.css` | Own component, §3.7 |
| Export download | `/account/exports/expenses.json`, `data-turbo="false"` | Plain anchor, unchanged |
| 32 `http.Redirect` calls after POST | `internal/handlers/*.go` | API returns the created/updated resource; client navigates |

### 2.4 Styles

`web/static/css/layout.css`, 1450 lines, hand-written, one file. Decision in §4.1; component libraries in §4.2.

### 2.5 Tests

Handler tests currently assert status codes and rendered pages (`handle_*_test.go`, 14 files;
10 after Phase 0B deletes the macro, food and mood ones).
They keep working against JSON with changed assertions. There is **no** frontend test tooling
today — adding it is Phase 0's optional tail, see §5.

---

## 3. Cross-cutting concerns — decide these before writing view code

### 3.1 Auth

`AuthMiddleware` (`internal/serve/middleware.go:41-77`) answers unauthenticated requests with
`303 → /login`. `fetch` follows redirects silently, so an SPA calling an API under that
middleware receives the login page's HTML with status `200`. Every "why is my JSON parse
failing" bug in this migration will be this.

**Resolution:** `/api/*` gets its own middleware stack that returns `401` with a JSON body and
no `Location`. The client's fetch wrapper turns a `401` into `window.location = "/login"`.
Guest API routes (`POST /api/login`, `POST /api/register`) are exempt.

Session storage itself does not change: `scs` server-side session, cookie `ninete_session`,
`HttpOnly`, `SameSite=Lax`, 7-day lifetime (`routes.go:setUpSession`). **Do not move auth to a
JWT in `localStorage`.** It is strictly worse — a token readable by JS turns any XSS into
account takeover, while the current cookie is unreadable by script — and it buys nothing here,
because the API is same-origin.

### 3.2 CSRF

`nosurf` (`middleware.go:79-92`) sets `ninete_csrf` as `HttpOnly`, so **JS cannot read the
token from the cookie**. It accepts the token from either the `csrf_token` form field or the
`X-CSRF-Token` header.

**Resolution:** the shell HTML carries `<meta name="csrf-token" content="{{ .csrfToken }}">`.
The fetch wrapper reads it once and sends `X-CSRF-Token` on every non-GET request.

**The token survives login and logout.** Worth stating because the opposite is the intuitive
guess: nosurf keeps its token in its own cookie (`ninete_csrf`, set by `SetBaseCookie` at
`middleware.go:83-89`) and regenerates only when that cookie is missing or malformed
(`nosurf@v1.2.0/handler.go:143` — `len(realToken) != tokenLength`). `handle_auth.go:50/90/105/110`
call `session.RenewToken` / `session.Destroy`, which touch the `scs` session cookie only. So the
`<meta>` token stays valid across a session boundary and a SPA that never reloads is fine. Do a
full page load after login/logout if it simplifies resetting client state — but not because the
token demands it.

Two traps that are real:
- **`ensureSameOrigin` runs on every unsafe request** (`handler.go:155`, `:187`), before the
  token is even looked at. It accepts `Sec-Fetch-Site: same-origin`, else falls back to `Origin`,
  else `Referer`. `isTLS` defaults to always-true (`handler.go:104`) and the app never calls
  `SetIsTLSFunc`, so `selfOrigin` is forced to `https` and an `Origin`/`Referer` of
  `http://localhost:3000` does **not** match. Browsers send `Sec-Fetch-Site` so real requests
  pass; anything hand-built does not. `internal/spec/http.go:81` already sets the header for this
  reason — every API handler test must do the same, or it fails CSRF with a 400 that looks like a
  bad token. Today only form posts exercise this path; every `fetch` will.
- `GET /api/*` must stay side-effect free, since nosurf only guards unsafe methods.

### 3.3 Session

Unchanged. Worth stating explicitly because it is the piece people break "while they are in
there": `SameSite=Lax` is fine — every API call is same-origin — and `Secure` stays tied to
`IsProduction()`. The 5-second request timeout (`WithTimeout`, defined at
`internal/serve/middleware.go:24`, applied at `middleware.go:326` in `setUpAppMiddlewares`)
applies to API routes too, and the `/api/*` stack must apply it as well.

**Exports do not escape that deadline.** `handle_exports.go:17` calls
`h.store.ExportExpenses(r.Context(), user.ID)` and materializes the whole slice before the first
byte is written, under the same 5-second context; only the `json.Encoder` write is streamed. If
an API export ever needs longer than 5s, that is unsolved work, not something the current design
already handles.

### 3.4 Security and CSP

Current policy (`middleware.go:153-166`): `default-src 'self'`, `script-src 'self' 'nonce-…'`,
`style-src 'self' 'nonce-…'`, no `unsafe-inline`, no `unsafe-eval`, `frame-ancestors 'none'`,
violations posted to `/csp-report`. Svelte 5 compiles to plain JS and needs no relaxation. Keep
it exactly as strict; if a phase seems to need `unsafe-inline`, the port is wrong.

Three rules that follow:

1. **Component `<style>` blocks are fine, but the emitted CSS file must be linked.**
   `bun-plugin-svelte` compiles component CSS into a sibling `.css` file next to the JS bundle
   (verified), so nothing is injected at runtime and the nonce never enters the picture. What
   *would* break the CSP is a Svelte config or library that injects `<style>` at runtime
   (`css: "injected"`), so keep the bundler emitting a file, and add the `<link>` to the shell
   the moment the first component carries styles — otherwise those styles silently do nothing.
2. **Bootstrapped JSON must be escaped.** If the shell embeds initial state in
   `<script type="application/json" nonce="…">`, a `</script>` inside user text (an expense
   description) breaks out of the tag. Escape `<`, or skip the problem entirely by having the
   client fetch its initial state — one extra round trip, no injection surface. **Prefer
   fetching.**
3. **`{@html}` is banned.** Svelte escapes `{expr}` by default; `{@html}` is the one way to
   reintroduce XSS, and nothing in this app needs it.

Unchanged and still required: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`,
HSTS in production, the auth rate limit on `/login`+`/register` (which must be re-pointed at
the new API routes, sharing **one** middleware value — see the invariant in `CLAUDE.md`), and
the 1 MB request body cap.

### 3.5 Validation and error display

Today a failed create re-renders the page with `error` set, which is why forms read their
values back from the template map. In the SPA the client holds the form state, so:

- API returns `422` with `{"error": "...", "fields": {"amount": "required"}}`.
- `500` stays a generic message; do not leak `err.Error()` for unexpected failures.
- Keep the message text identical to today's, so the port is verifiable by eye.

**Built in Phase 0.3.** `logic.ValidationError` (`internal/logic/logic.go`) carries the failed
fields beside the flat message the templates already print — `Error()` is byte-for-byte what it
was, so no page changed. `Handler.WriteAPIError` (`internal/handlers/api.go`) is the only
mapping from a store error to a response:

| Error | Answer |
| --- | --- |
| `*logic.ValidationError` | `422`, `fields` filled from the validator |
| `sql.ErrNoRows` | `404` |
| An error the endpoint named in `userErrors` | `422` with that error's own message |
| Anything else | logged in full, `500` with a generic message |

**Nothing is user-facing by default.** That is the point of the last row: today a failed insert
re-renders the form with the driver's text, so a `UNIQUE` violation names the table and column
on a public page. An endpoint opts a message in by naming its sentinel, which also documents
what that endpoint expects to go wrong.

**`fields` keys are snake_case** — `category_id`, `occurrence_limit` — matching the existing
form field names and the columns behind them, derived from the Go field by `snakeFieldName`. The
value is the validator's rule (`required`, `gte`, `email`), not a sentence: the client phrases
it, and no message text is invented server-side. **JSON request and response bodies use the same
snake_case names**, so the client sends back exactly the keys it was told about.

### 3.6 Dates — the part that must not break

**Read this section before touching anything dated: expenses, budgets, the dashboard, or the
recurrent-expense cron.** Dates are the one area where a port can look completely correct,
pass every test, and still be wrong by one day for the user — silently, and only sometimes.
Expenses are the app's core feature and every expense is filed under a calendar date that
decides which billing window it lands in. An off-by-one here does not look like a bug, it looks
like the user misremembering.

#### The two kinds of value, which must never be confused

| Kind | Examples | Stored as | Means |
| --- | --- | --- | --- |
| **Calendar date** | `expenses.date`, `recurrent_expenses.last_copy_created_at`, budget month bounds | epoch seconds at **UTC midnight** of that date | "the 21st", a label on a calendar. Has no time and no zone |
| **Instant** | `created_at`, `updated_at`, export `exported_at` | epoch seconds of a real moment | a point in time, displayed in the viewer's zone |

After §0, calendar dates survive in exactly two places — `expenses.date` and
`recurrent_expenses.last_copy_created_at` — which narrows this section's blast radius but does
not soften any of its rules. `last_copy_created_at` is the one to watch: the name says instant,
the value is a calendar date.

Both are `INTEGER` in SQLite and both are `int64` in Go, so **the type system will not catch a
mix-up**. The distinction lives only in how each value is formatted and compared. Every date bug
in this migration will be an instant treated as a calendar date, or the reverse.

#### The contract as it works today (verified, keep it)

Write path — `dateController.ts:20`: the `<input type="date">` yields `YYYY-MM-DD` in the user's
local calendar; the controller appends `T00:00:00Z` and posts that; `prog.StringToUnixDate`
(`internal/prog/utility.go:44`) parses it as RFC3339 and stores `.Unix()`. Net effect: the
calendar date the user picked, pinned to UTC midnight, independent of their zone.

Read path — `localDateController.ts`: calendar dates format with `getUTC*` getters; instants
(`datetime: true`) format with local getters and carry a full local timestamp in `title`. The
two code paths exist precisely because the two kinds differ.

Ranges — `computeDateRange` (`expense_shared.go:86`) builds `[start, end)` half-open bounds at
UTC midnight, using `tz_offset` **only** to decide which month "now" falls in for the client.

Search — `expense_search.go:115` parses `YYYY-MM-DD` with `time.Parse` (UTC) and takes `.Unix()`.
Recurrent copies — `logic_recurrent_expense.go:180` files every generated expense at
`time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)` and stores the same value in
`last_copy_created_at`. That is a calendar date computed server-side, with no client zone
involved, and it stays that way — the SPA must not start sending an offset into it.

**Half-open `[start, end)` is the convention everywhere. Do not introduce an inclusive end.**

#### Rules for the Svelte port

1. **Format calendar dates with UTC getters only.** `new Date(epoch * 1000).getDate()` returns
   the previous day for any user west of UTC. Every calendar-date render must use `getUTCDate`,
   `getUTCMonth`, `getUTCFullYear` — the same split `localDateController` already makes. Port
   that file's **three** formatters verbatim into one shared module and use it everywhere; do not
   hand-roll formatting per component. All three are live: `formatDateUTC` (line 45, calendar
   dates), `formatDate` (line 37, local, the visible text of an instant) and `formatDateTime`
   (line 53, local, the `title` tooltip an instant carries alongside it — see line 15 of the
   controller). Dropping `formatDateTime` silently removes that tooltip everywhere.
2. **Prefer carrying calendar dates as `YYYY-MM-DD` strings in client state**, converting to
   epoch only at the API boundary. A string cannot be accidentally shifted by a zone; an epoch
   can. This is the single highest-value habit in this section.
3. **Never `toISOString()` a `Date` built from local getters** — it re-applies the offset and
   moves the day.
4. **Beware JS's own inconsistency:** `new Date("2026-08-22")` parses as **UTC** midnight, but
   `new Date("2026-08-22T00:00:00")` parses as **local** midnight. Same-looking strings, one
   day apart in output. Always parse explicitly.
5. **Instants keep local formatting.** `created_at` should read in the user's zone. Do not
   "fix" it to UTC for consistency — the two kinds are supposed to differ.

#### Retiring `tz_offset` on the API side (Phase 3) — but not deleting it (Phase 7)

`tz_offset` has **two** consumers, and only one of them is a range calculation. Getting this
wrong is the exact failure class this section exists to prevent, so both are spelled out.

**Consumer 1 — named date ranges.** `computeDateRange` (`expense_shared.go:86`) uses the offset
only to decide which month "now" falls in for the client. In the SPA the client already knows its
zone, so the **API** takes explicit bounds instead:

```
GET /api/expenses?start=1754006400&end=1756684800     ← client computed, UTC midnight, half-open
```

instead of `?date_range=this_month&tz_offset=-120`. The named ranges (`this_month`,
`last_month`, `six_months`, `this_year`, `this_week`, `next_month`) stay as client-side labels
that resolve to bounds. That deletes the "did I remember to attach tz_offset to this fetch" bug
class and stops the server from guessing what "now" means for someone else.

**Consumer 2 — quick-add relative dates.** `handle_quick_expense.go:31` passes `parseTZOffset(r)`
into `logic.ParseQuickExpense`, and `parseQuickDate` (`logic_quick_expense.go:131`) builds
`time.FixedZone("client", -tzOffsetMinutes*60)` to resolve `today`, `yesterday` and weekday names
against the *client's* calendar. This has nothing to do with `computeDateRange`. Drop the offset
here and the zone silently becomes UTC, shifting every relative quick-add by a day for any
non-UTC user. **The API quick-add endpoint must keep taking a client zone offset**, sent
explicitly in the request rather than injected by a Turbo hook. Quick-add is Phase 3 scope, so
this lands in the same PR as the range change — do not let "retire `tz_offset`" delete it.

**Both server paths stay alive until Phase 7.** Phase 3 adds explicit bounds to the API; it does
**not** turn `computeDateRange` into pure validation and does **not** delete `parseTZOffset`.
Four template-side call sites keep computing from `tz_offset` while the templates exist:
`handle_dashboard.go:50-52` (dashboard is Phase 4), `handle_expenses.go:363`, `shared.go:120`,
`handle_expense_budgets.go:122`. Phase 3's exit criterion is running the template version beside
the SPA as an oracle, which requires the template path to still *compute*; retiring it in Phase 3
would destroy the oracle in the phase that leans on it hardest.

The retirement itself — `computeDateRange` reduced to bounds validation, `parseTZOffset` deleted
once quick-add's explicit offset is the only zone input left — happens in Phase 7 with the rest
of the HTML handlers. Keep the boundary rule `computeDateRange` encodes wherever it ends up:
`start` inclusive, `end` exclusive, both UTC midnight. `/expenses/budgets` reads its month
buckets off the same bounds, so its month/months mode selection moves client-side with the
labels.

#### How this gets verified (not optional)

- **Run the frontend tests under a non-UTC zone.** `TZ=Pacific/Auckland` (UTC+12/+13, and it
  observes DST) makes every "local getter used on a calendar date" bug fail immediately. Under
  `TZ=UTC` — the CI default — all of them pass. Add `TZ` to the vitest config in Phase 0, and a
  second CI run at `TZ=America/Los_Angeles` for the negative-offset direction.
- **Boundary cases each dated resource must cover:** the 1st and last day of a month; a date in
  a DST transition week for the local zone; a range crossing a year boundary; an expense
  entered near local midnight (the case where the user's calendar day and UTC's disagree).
- **Manual acceptance for Phase 3:** with the machine's zone set to UTC+13 and again to UTC-8,
  create an expense dated the 1st of the month and confirm it appears in `this_month`, reads
  back as the 1st on the list, the show page, the stats bucket, and the budgets page.
- **Cross-check against the old implementation while both exist.** Until Phase 7, the template
  version is a working oracle: same user, same data, same range, compare the two listings. This
  is the strongest reason to keep dates in an early phase rather than a late one.

### 3.7 Navigation and loading feedback

The Turbo progress bar disappears with Turbo. Replacement: a router-level pending flag driving
the same `.turbo-progress-bar` styling in `layout.css` (rename it). Also gone: Turbo's scroll
restoration and link prefetching — both need explicit handling in the router if wanted.

Router: **hand-rolled, path-based** (decided). A `$state` holding `location.pathname`, a
`popstate` listener, a click handler intercepting internal `<a>` clicks, and a match table —
roughly 80–120 lines for this app's flat route shape (nothing nests deeper than
`/resource/:id/action`). Path-based keeps URLs identical to today's, which means Go must serve
the shell for every non-API, non-static path (a catch-all) so a hard refresh on
`/expenses/12/edit` works. Scroll restoration and any link prefetching are ours to write.

### 3.8 Build and deploy

Today `make build-static-js` runs `bun build` on the command line. That cannot compile `.svelte`
files — the CLI takes no plugin — so Phase 0 replaces it with a `web/build.ts` script calling
`Bun.build()` with `bun-plugin-svelte`. Production runs the same target through
`scripts/build-js.sh`, which does `bun install` first, so no deploy-script change is needed.
Two more things to add during this migration:

- **Minification.** Not enabled today. The bundle is already 0.65 MB unminified (`make
  build-static-js`, 1745 modules) and Svelte plus a router will not shrink it.
- **A content hash in the filename**, or `/static/*`'s `max-age=300` (`routes.go`) keeps
  serving a stale `index.js` after every deploy for five minutes. Today this only bites during
  development; in an SPA the whole app is that file.

### 3.9 Where the `.svelte` files live

Decided here rather than discovered in Phase 2, because Phase 1 already writes the shell, the
router and the chrome components — by the time a phase "sets the conventions" the layout is
whatever Phase 1 happened to do.

**Sources move out of `web/static/`.** `setUpFileServer` mounts the whole directory
(`http.FileServer(http.Dir("./web/static/"))`, `internal/serve/routes.go:148`), so everything
under it is publicly fetchable — `GET /static/js/index.ts` serves TypeScript source today, and
`.svelte` files placed there would serve the same way. Nothing secret lives in a component, so
this is tidiness rather than a vulnerability, but there is no reason to ship sources once the
build step exists. Only build output stays under `web/static/`.

```
web/
  app/                      ← sources; never served, never linked from a template
    index.ts                entry point (Bun.build input; replaces web/static/js/index.ts)
    App.svelte              shell: chrome + <Router/>
    router.ts               §3.7's hand-rolled path router + the match table
    lib/
      api.ts                §3.2 fetch wrapper, X-CSRF-Token, 401 handling
      dates.ts              §3.6 formatters and YYYY-MM-DD ⇄ epoch helpers
      icons.ts              moved from web/static/js/
    components/             shared, resource-agnostic: FormButtons, FormError,
                            Pagination, Header, Footer, Spinner, LocalDate
    routes/
      expenses/             Index.svelte, New.svelte, Edit.svelte, Show.svelte,
                            Stats.svelte, Budgets.svelte, Form.svelte, QuickForm.svelte
      recurrent_expenses/ account/ auth/ dashboard/
  static/
    js/build/               ← emitted bundle + CSS sibling + manifest; the only served JS
  views/
    layout.html             the shell template, and after Phase 7 the only template
```

Four rules that follow:

1. **`routes/<resource>/` mirrors `web/views/<resource>/`, one file per action.** The port stays
   checkable by eye against the template it replaces, which is the whole basis of the freeze
   rule. `_form.html` becomes `Form.svelte` in the same directory, not a shared component.
2. **`components/` is for things used by more than one resource.** A component used by exactly
   one resource lives beside that resource's routes. Resist a flat `components/` bucket — it is
   how the partial namespace problem in `web/README.md` reappears in a new form.
3. **`lib/` holds no components.** Plain `.ts` modules only, so they are unit-testable under
   vitest without a DOM.
4. **Filenames: `PascalCase.svelte` for components, `camelCase.ts` for modules** — matching the
   existing `web/static/js/` convention (`localDateController.ts`, `icons.ts`).

`web/static/js/index.ts` and `web/static/js/controllers/` stay where they are until Phase 7,
which deletes them. During the coexistence phases there are two entry points; `web/build.ts`
builds both.

---

## 4. Tailwind and component libraries — decide separately, and later

### 4.1 Tailwind

Tailwind v4 builds fine under bun, scans `.svelte`, emits a static file, and does not touch the
CSP. That is not the question. The question is that `layout.css` is 1450 hand-written,
currently-working lines, and Tailwind does not reuse them — adopting it means rewriting the
app's entire visual layer *at the same time as* changing its rendering model. Two variables,
one bisect.

**Recommendation:** port to Svelte with the existing CSS untouched (components reference the
same class names). If a redesign is wanted afterwards, do it as its own project against a
codebase where every view is already a component — which is exactly the codebase that makes a
Tailwind migration cheap. Interim mixed state (Tailwind on new views, `layout.css` on old) is
possible but leaves the app looking like two apps for however many months this takes.

### 4.2 Component libraries (Flowbite, Melt UI, Bits UI, shadcn-svelte)

Same answer as Tailwind — **not during the migration** — but for a different and stronger reason,
so it gets its own section rather than a footnote.

**Why not during.** The freeze rule at the top of this plan says every phase is a like-for-like
port, and that rule is the only thing making the migration checkable: if a ported view looks
wrong, the bug is in the port. A library's `<Select>` is not the app's current `<select>`. It has
its own markup, its own keyboard behavior, its own focus handling and its own open/close
animation. Adopting one mid-migration means every view diff contains two changes — a rendering
model change and a UI change — and the side-by-side oracle in Phase 3 stops being an oracle.
That is the whole argument; the rest is detail for whoever picks this up later.

**The Tailwind coupling is not optional, and it decides most of this.** Component libraries split
into two kinds:

- **Styled, Tailwind-based** — Flowbite-Svelte, shadcn-svelte, DaisyUI. Their components *are*
  Tailwind class strings. Adopting one is adopting Tailwind, so §4.1 applies in full and this is
  not a separate decision. Do not let "we're only adding a few components" smuggle in a
  visual-layer rewrite.
- **Headless / unstyled** — Melt UI, Bits UI. They ship behavior (focus trap, roving tabindex,
  ARIA wiring, floating positioning) and no CSS, so they compose with the existing `layout.css`
  class names. If a library is ever adopted here, it should be one of these: it is additive
  rather than a rewrite, and it buys the part that is actually hard to hand-write.

**What to check against this app before adopting any of them.** These are the questions this
codebase makes non-obvious, not a general evaluation checklist:

1. **CSP.** `style-src 'self' 'nonce-…'` with no `unsafe-inline` (`buildCSP`,
   `middleware.go:153-166`), and a nonce does not cover `style="…"` attributes. Styles written
   through CSSOM from JS (`el.style.transform = …`, how floating-ui positions a dropdown) are not
   subject to CSP and are fine; a library that injects a `<style>` element at runtime, or renders
   inline style attributes into parsed markup, is not. **Verify, do not reason about it:** mount
   the library's most dynamic component — a dropdown or a tooltip, the ones that position
   themselves — under the real policy and watch `/csp-report`. The report endpoint already exists
   for exactly this. §3.4 rule 1 is the same trap in the Svelte-compiler direction.
2. **`{@html}`.** §3.4 rule 3 bans it. A library that takes a `content`/`label` prop and renders
   it as HTML reintroduces the XSS surface that ban exists to close.
3. **Svelte 5 runes.** Libraries written against Svelte 4's builder/store API are not drop-in.
   Check the version actually targets runes rather than running in legacy mode.
4. **Bundle size.** 0.65 MB unminified before Svelte, a router, Chart.js and a component library
   (§3.8). Minification and the content hash are prerequisites, not follow-ups.
5. **Accessibility is the thing being bought.** If the library is adopted for looks rather than
   for focus management and ARIA wiring, it is the Tailwind decision wearing a different hat.

**Recommendation:** finish the migration on `layout.css` and hand-written markup. Afterwards,
adopting a headless library is a small, reversible, component-at-a-time change against a codebase
where every view is already a component — and by then it will be clear which widgets actually
needed it (the searchable selects and the date pickers, most likely) instead of guessing now.

---

## 5. Phases

Each phase ends with a green `make test`, `make lint-fix`, and a working app. No phase leaves
its base branch unshippable. Every phase is one PR.

**Base branch: `spa-base`, not `main`.** `spa-base` is cut from `main` and every `spa/*` phase
branch targets it; `main` receives the whole migration in one merge at the end. It is not called
`spa` because git refs are a directory hierarchy and `spa/phase-0-groundwork` already occupies
that name. Phase PRs land against `spa-base`, so `git merge main` into `spa-base` is the way an
unrelated `main` change reaches the migration — not the other way around. Do not open a phase PR
against `main`.

### Phase 0 — Groundwork (no view changes)

Branch: `spa/phase-0-groundwork`. Nothing a user can see changes in this phase. Its job is to
make every later phase mechanical.

**0.1 Build tooling** — landed (#113)
- `bun add svelte bun-plugin-svelte` and `bun add -d vitest @testing-library/svelte`.
- `web/build.ts`: `Bun.build()` with `SveltePlugin`, outdir `web/static/js/build`. Point
  `build-static-js` in the `Makefile` at `bun run web/build.ts`. Entry in this phase is still the
  Stimulus `web/static/js/index.ts`; Phase 1 adds `web/app/index.ts` as a second entry and
  Phase 7 removes the first (§3.9).
- Create `web/app/` with the `lib/` modules below. It is not bundled yet — nothing imports it
  until Phase 1 — but `dates.ts` is unit-tested from 0.5 without a bundle.
- Turn on minification. Keep the output filename `web/static/js/build/index.js` in this phase.
- **Content hashing waits for Phase 1**, where the shell is being rewritten anyway. It cannot
  land here as a filename change alone: `web/views/layout.html:31` hardcodes
  `src="/static/js/build/index.js"` and `internal/serve/routes_test.go:99` asserts that exact
  path returns 200, so a hashed name in Phase 0 fails `make test` and ships a shell with no JS —
  the opposite of this phase's exit criterion. The mechanism it needs, in Phase 1: `web/build.ts`
  writes a manifest next to the bundle, Go reads the manifest at startup and exposes the filename
  through `setTmplData` the way `version` already is, the shell reads it from the template map,
  and `routes_test.go` reads the same manifest instead of a literal path.
- Reference implementation: `web/build.ts` on `experiment/svelte-api`.

**0.2 API route group** — landed (#113)
- `/api/*` under its own middleware stack in `internal/serve/routes.go`: session load, body
  limit, timeout, CSRF — but **not** `setTmplData`, and **not** the HTML `AuthMiddleware`.
- API auth middleware: `401` + JSON body, never a `Location` header (§3.1).
- **It must also put the signed-in user into `KeyCurrentUser`.** Dropping `setTmplData` drops the
  only place that happens (`internal/serve/middleware.go:125`), and `getCurrentUser`
  (`handle_auth.go:119-127`) **panics** when the key is absent. Every resource handler opens with
  `user := getCurrentUser(r)`, so without this the first `GET /api/recurrent-expenses` in
  Phase 2 panics into a 500 that looks like a Svelte or fetch bug. Do the `FindUser` half of
  `setTmplData` — session lookup, `sql.ErrNoRows` handled the same way — and skip the template
  map.
- The auth rate limit currently guarding `POST /login` and `POST /register` must guard their
  API equivalents when those land in Phase 6, still sharing **one** middleware value — see the
  invariant in `CLAUDE.md`.

**0.3 JSON handler plumbing** — landed (#114)
- Naming: `internal/handlers/handle_api_<resource>.go`, keeping the `handle_` convention.
- One shared JSON write helper and one error mapper: `422` with `{"error", "fields"}` for
  validation, `404`, `500` with a generic message (never `err.Error()` for unexpected
  failures), all in `internal/handlers/` alongside `render.go`.
- Sentinel errors go in the existing `errs.go`, not inline.

**0.4 Client fetch wrapper**
- `web/app/lib/api.ts`: `X-CSRF-Token` from `<meta name="csrf-token">`, `401` →
  `window.location.assign("/login")`, JSON error envelope parsing, typed helpers.
- The shell gains the `<meta>` tag in Phase 1; until then read it from `layout.html`, which
  already has `cspNonce` and can carry `csrfToken` the same way.

**0.5 Dates module (see §3.6 before writing it)**
- `web/app/lib/dates.ts`: all three formatters from `localDateController.ts` ported verbatim —
  `formatDateUTC` (UTC getters, calendar dates), `formatDate` and `formatDateTime` (local
  getters, an instant's text and its `title` tooltip) — plus `YYYY-MM-DD` ⇄ epoch helpers for the
  API boundary.
- Tests for all three, including the month-boundary and DST cases named in §3.6.

**0.6 Test setup**
- `vitest` + `@testing-library/svelte`, with `TZ=Pacific/Auckland` in the config and a second
  CI run at `TZ=America/Los_Angeles`. Under CI's default `TZ=UTC` every date bug in §3.6
  passes silently, which is the entire reason this is Phase 0 work and not Phase 5 work.
- A `make test-js` target, wired into CI next to `make test`.

**0.7 Documentation**
- `web/README.md`: how the Svelte build works, the `web/app/` layout from §3.9 (including why
  sources sit outside the served `web/static/` tree), the `.svelte` files are
  not covered by prettier/eslint caveat.
- `CLAUDE.md`: the `/api/*` group in the route map.

Exit criteria:
- `make build-static-js`, `make test`, `make test-js` and `make lint-fix` all green.
- A signed-out request to any `/api/*` route returns `401` JSON with no `Location`.
- A signed-in POST to an `/api/*` route without `X-CSRF-Token` is rejected.
- `dates.ts` tests pass under both configured zones.
- The app looks and behaves exactly as it does today.

### Phase 0B — Retire macros, foods and moods (code only)

Branch: `spa/retire-macros-foods-moods`. One PR. The only phase that removes a feature instead
of porting one, and the only deliberate exception to the freeze rule (§0). **It must land
before Phase 2**, which was going to pilot on Foods.

Delete, in one change:

- Routes: the `/macros`, `/foods` and `/moods` groups in `internal/serve/routes.go`, plus the
  four `POST /account/{macro-entries,macro-goals,foods,moods}/delete-all` endpoints. 32 routes.
- Handlers: `handle_macros.go`, `macro_shared.go`, `handle_foods.go`, `handle_mood_entries.go`
  and their four test files.
- Logic: `logic_macro.go`, `logic_food.go`, `logic_mood_entry.go`, `mood.go` and their tests.
- Views: `web/views/macros/`, `web/views/foods/`, `web/views/mood_entries/` — 18 files — and the
  three nav links at `common/_header.html:25-27`.
- Controllers: `macroCalc`, `macroDate`, `macroSelect`, `macroTrend`, `moodChart`,
  `submitOnChange`, with their `window.Stimulus.register` lines in `web/static/js/index.ts`
  (see §2.2 on why `submitOnChange` goes too).
- The `TemplateName` constants for every deleted view. A leftover constant compiles fine and
  fails at request time, which is the failure mode `CLAUDE.md` warns about.
- `repo.TaggableTypeMoodEntry`, and the mood branch of `logic_tag.go` / `repo/tagging.go`.
  `TaggableTypeExpense` and `TaggableTypeRecurrentExpense` stay.
- The macro half of the dashboard: `dashboardMacros`, `buildDashboardMacros`,
  `computeMacroProgress` and the `macros` key in `handle_dashboard.go:20-42,135-163`, plus the
  matching block in `dashboard/index.html`. **This is what makes the dashboard a valid Phase 4
  oracle again** — a like-for-like port cannot be checked against a page whose right half is
  about to disappear.
- The delete-data counts for the dropped resources: `MacroEntries`, `MacroGoals`, `Foods`,
  `MoodEntries` in `logic_account.go`, and their rows in `delete_data/index.html`.

Keep, deliberately:

- **The tables and every existing migration.** No migration is written in this phase (§0).
- `internal/repo/{food,macro_entry,macro_goal,mood_entry}.go` with their column constants, so
  `TestColumnConstantsMatchSchema` keeps guarding the schema until Phase 8 drops it. They end
  up with no callers; that is intended, and `unused` does not flag exported methods. If the
  linter does complain, delete the method rather than reviving a caller.
- `internal/spec` factories only where a surviving repo test still uses them.

One consequence to state rather than discover: **`POST /account/delete-all` stops clearing the
dropped tables.** Rows in them outlive "delete all my data" until Phase 8 drops the tables.
For a single-user app with the owner making the call that is acceptable, but it should be a
decision, not a surprise.

Also in this PR: `CLAUDE.md`'s feature/route map and its tags cross-cutting note (the
project-scope paragraph already carries the decision); `web/README.md` wherever it names a
deleted controller; and `docs/architecture.md` if it maps the dropped packages.

Exit criteria:
- `make test`, `make test-js` and `make lint-fix` green.
- `/macros`, `/foods` and `/moods` return 404, and nothing in the nav links to them.
- The dashboard renders the expense summary alone.
- `grep -riE "macro|food|mood" internal/handlers internal/logic internal/serve web/views` comes
  back empty. `internal/repo` and `internal/db` are expected to still match — that is §0's split.

### Phase 1 — Shell and router, coexisting with templates

Serve the SPA under `/app/*` while every existing route keeps working. Shell template (with the
`<meta name="csrf-token">` tag from §3.2), hand-rolled path-based router, layout chrome
(`_header`, `_footer`, nav) as components, loading indicator, theme handling (the anti-FOUC
inline script in `layout.html:9-22` stays server-rendered and nonce'd), icon rendering.

This phase creates `web/app/index.ts`, `App.svelte`, `router.ts` and the first entries in
`components/` — the layout in §3.9, which is settled before this PR opens, not discovered in it.
`web/build.ts` gains it as a second entry point alongside the Stimulus bundle.

The content-hashed bundle filename deferred from Phase 0.1 lands here, with the manifest
mechanism described there: `web/build.ts` writes the manifest, Go reads it at startup and passes
the filename through `setTmplData`, both shells read it from the template map, and
`routes_test.go:99` reads the manifest instead of the literal `/static/js/build/index.js`.

Exit: `/app` renders the real chrome around a placeholder route, a hard refresh on a nested
path like `/app/recurrent-expenses/1/edit` still resolves, every template route is untouched,
and no template or test hardcodes the bundle filename.

### Phase 2 — First real resource: Recurrent expenses

Foods was the pilot because it was the smallest full CRUD; §0 deletes it, so the pilot moves to
recurrent expenses — now the only resource left with a complete `index`/`new`/`edit`/`show`/
`_form` set and neither charts nor search. It costs one thing Foods did not: **tags arrive in
Phase 2 instead of Phase 3.** That is a fair trade rather than a regression, because tags are
shared with expenses and are easier to get right against the simpler resource.

It is the first resource to fill in `web/app/routes/<resource>/` (§3.9), so it sets the
conventions that live *inside* that shape — form handling, validation display, list/detail data
flow, the tag input, and the `filter` control that `recurrent_expenses/index` and `archived`
both use — while the file layout itself is already fixed. Every remaining resource copies
whatever this PR does, so review it as a template, not as one page.

It also carries `last_copy_created_at`, a calendar date wearing an instant's name (§3.6), which
makes it a useful first exercise of `dates.ts` against real data.

Exit: every recurrent-expense view works under `/app/recurrent-expenses`, including the
archived listing and archive/unarchive, with `422` validation shown inline.

### Phase 3 — Expenses

The hard one, and deliberately early so its surprises land before the rest depend on the
conventions: search, filters, sort, pagination, quick-add form, stats, budgets, tags. Chart.js
arrives here rather than in Phase 4, since `expenses/stats` is the only `chart` use left after
§0.

This is also the phase that retires `tz_offset` **on the API side** (§3.6): named ranges resolve
client-side to explicit `[start, end)` UTC-midnight bounds. The template path keeps computing
from `tz_offset` — it is this phase's oracle — and quick-add keeps taking a client zone offset,
now sent explicitly. `computeDateRange` and `parseTZOffset` are not touched until Phase 7.

Exit: `/app/expenses*` matches the template version feature for feature — **verified by running
both versions side by side** against the same data at `TZ=UTC+13` and `TZ=UTC-8`, comparing the
listing, the show page, the stats buckets and the budgets page. While the templates still
exist, they are a working oracle; after Phase 7 they are not.

### Phase 4 — Dashboard

All that §0 leaves here: the expense summary and its date picker (`dashboardDate`). Recurrent
expenses moved to Phase 2 and the other three resources are gone, so this is a small phase —
which is fine, it is not worth folding into Phase 3 given the dashboard is its own view with
its own range logic.

`handle_dashboard.go:50-52` computes its range from `tz_offset`. Apply §3.6's rules and re-run
the two-zone check here specifically; do not assume Phase 3 settled it.

### Phase 5 — Account, delete-data, exports

Destructive posts and the file download. Confirmation flows must stay at least as deliberate as
today's — a client-side `confirm()` is not equivalent to a form post. Export links stay plain
anchors.

### Phase 6 — Auth views

Login, register, logout. Last because they are the session boundaries: `RenewToken` and
`Destroy` (`handle_auth.go:50/90/105/110`) invalidate every piece of client state the SPA is
holding, and the auth rate limit has to move to the API routes as **one** shared middleware value
(the invariant in `CLAUDE.md`). A full page load after a successful login or logout is the
simplest way to reset that state — not because the CSRF token expires, which it does not (§3.2).

### Phase 7 — Flip the switch

Move the SPA from `/app/*` to `/`, add the catch-all serving the shell for any non-API,
non-static path, delete every per-view template and `TemplateName` constant, delete the HTML
handlers, drop `@hotwired/turbo` and `@hotwired/stimulus`, delete `web/static/js/controllers/`.

This is also where the date helpers the templates kept alive finally retire (§3.6): the four
template call sites go with their handlers, `computeDateRange` drops to bounds validation, and
`parseTZOffset` is deleted once the API quick-add's explicit offset is the only zone input left.
Deleting the templates removes the oracle, so do not combine this with any behavior change.

Exit: `web/views/` holds the shell and nothing else.

### Phase 8 — Cleanup, documentation, and dropping the retired tables

Rewrite `web/README.md` (its template-partial namespace and Turbo/Stimulus sections become
wrong), update `CLAUDE.md`'s route map, UI/Assets section and the render-helper invariant,
update `docs/architecture.md`'s request flow. Remove dead Go helpers (`render.go`'s page
helpers, `tmplData`, the template loader in `internal/serve/template.go`).

Then the second half of §0's removal — **the only irreversible step in this plan**:

- **Export first.** `foods`, `macro_entries`, `macro_goals` and `mood_entries` hold history no
  export path covers. Dump them to a file, confirm with the owner that it is kept somewhere,
  and only then write the migration. A `Down` section is not a backup: it recreates empty
  tables.
- One migration dropping those four tables and the `taggings` rows with
  `taggable_type = 'mood_entry'`, incrementing `PRAGMA user_version`, with a `Down` that
  recreates the empty tables and says in a comment that the data does not come back.
- Delete `internal/repo/{food,macro_entry,macro_goal,mood_entry}.go`, their column constants
  and their tests **in the same change** — otherwise `TestColumnConstantsMatchSchema` runs a
  column constant against a table that no longer exists.
- The deploy ordering in `CLAUDE.md` says migrations apply while the previous binary is still
  serving, so a migration that removes something the running code reads errors until the
  restart. This one is safe precisely because Phase 0B removed every caller a release earlier:
  the running binary still contains those repo methods and never invokes them. That ordering is
  the reason for the two-step split, not an accident of it.

---

## 6. Risks, honestly

- **Duration.** 29 views, 12 controllers, 35 routes after §0 — about a third less than this
  plan was originally sized for, but still months of evenings rather than a weekend. The phase
  split exists so that stopping after any phase leaves a working app; Phase 4 is where it is
  safe to pause indefinitely.
- **Doubling the handler layer.** Between Phase 0 and Phase 7, HTML handlers and API handlers
  both exist. That is the cost of not doing a big-bang rewrite, and the reason for the freeze
  rule: two implementations of a *stationary* target stay in sync; two implementations of a
  moving target do not.
- **What is genuinely lost.** No-JS operation, progressive enhancement, and the guarantee that
  a rendered page is a complete page. For a single-user app this is acceptable — but it is a
  real loss, not a nothing.
- **What is genuinely gained.** Client-held form state (no more re-render-to-preserve-input),
  real interactivity without controller-per-behavior plumbing, and a JSON API that a future
  mobile client or script could use.
- **The honest alternative.** Islands — Svelte components mounted inside the existing
  templates, Turbo left in place — gets most of the interactivity benefit for a fraction of
  this plan, and Phases 0–4 of this plan produce it anyway. If the SPA payoff stops feeling
  worth it around Phase 4, stopping there is a coherent end state, not a failure.

## 7. Decisions made (do not revisit)

1. **All phase work merges into `spa-base`, not `main`** (§5). `main` sees the migration once,
   at the end.
2. **SPA is staged under `/app/*`** until Phase 7, then moves to `/`. Every phase stays
   revertable by deleting a route.
3. **Router is hand-rolled and path-based.** No router dependency.
4. **`vitest` + `@testing-library/svelte` land in Phase 0**, configured with a non-UTC `TZ` (see
   §3.6) so date bugs fail on the first run rather than in production.
5. **Tailwind is deferred** to a separate post-migration project. Components reference the
   existing `layout.css` class names unchanged.
6. **No component library during the migration** (§4.2) — it breaks the freeze rule and the
   Phase 3 oracle, and every styled library (Flowbite, shadcn-svelte, DaisyUI) is decision 5
   in disguise. If one is adopted afterwards it should be headless (Melt UI, Bits UI), added a
   component at a time, and checked against `/csp-report` first.
7. **Sources live in `web/app/`, outside the served `web/static/` tree** (§3.9), with
   `routes/<resource>/` mirroring `web/views/<resource>/` one file per action. Only build output
   stays under `web/static/js/build/`. Settled in Phase 0, not discovered in Phase 2.
8. **Auth stays cookie-session.** No JWT, no token in `localStorage`.
9. **`tz_offset` is retired from the API in Phase 3**, replaced by client-computed explicit
   bounds. The server-side helpers (`computeDateRange`, `parseTZOffset`) stay until Phase 7
   because the templates are Phase 3's oracle, and quick-add keeps a client zone offset for its
   relative dates regardless — see §3.6.
10. **Macros, foods and moods are dropped from the app** (§0, decided 2026-08-23). They are not
    ported, not stubbed and not revisited. Expenses and what hangs off them — recurrent
    expenses, budgets, tags, categories, dashboard, exports, delete-data, auth — are kept.
11. **The removal is split in two: code in Phase 0B, tables in Phase 8.** Code goes early
    because deferring it means porting views that get deleted; tables go last because dropping
    them destroys history with no export path, and because the running binary must lose its
    callers a release before the schema loses the tables.
12. **The Phase 2 pilot resource is recurrent expenses**, replacing Foods. Tags therefore land
    in Phase 2 rather than Phase 3.

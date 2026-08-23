# Plan — Migrate NINETE to a Svelte SPA

Status: approved 2026-08-22. Phase 0 not started.
Audience: the maintainer reading it once, and an agent picking up any single phase later
without the conversation that produced it. Section 7 records the decisions already made — do
not re-litigate them.

**Starting point.** `main` has no Svelte tooling at all. A throwaway spike on
`experiment/svelte-api` (commit `84b9d4a`) proved the pieces work together — Svelte 5 under
`bun-plugin-svelte`, mounted from a Stimulus controller, reading a Go JSON endpoint — and the
facts this plan states about the bundler were verified there. Treat that branch as a reference
to read, not a branch to merge: Phase 0 re-lands only the build tooling, deliberately, without
the experiment route.

**Before writing any view code, read §3.6 (Dates).** It is the one section where a mistake is
invisible in review, invisible under the default `TZ=UTC` test run, and wrong for the user.
Expenses are the app's core feature and every one of them hangs off a calendar date.

**Freeze rule while this runs:** no new features and no behavior changes to existing views.
Every phase is a like-for-like port. If a view looks wrong after porting, the bug is in the
port, not in the design — that is the only reason this migration is checkable at all.

---

## 1. Goal and end state

Today: Go `html/template` renders every page, Turbo intercepts navigation and form posts,
Stimulus adds per-page interactivity. 47 template files, 19 Stimulus controllers, 67 routes.

End state:

- Go serves one HTML shell plus a JSON API. No per-view templates.
- Svelte 5 owns rendering, routing and client state.
- Turbo and Stimulus are gone, along with `@hotwired/*` from `package.json`.
- Session cookie auth, CSRF, and the CSP survive unchanged in shape — see §3.

Non-goals: server-side rendering of components, SvelteKit (Go serves the HTML), offline
support, multi-user features, any visual redesign.

---

## 2. Inventory — everything that must be migrated

Nothing below is optional. A phase is done when its rows are ticked.

### 2.1 Views (47 files, `web/views/`)

| Resource | Files | Notes |
| --- | --- | --- |
| `layout.html` | 1 | Becomes the shell. Last thing to change, §Phase 7 |
| `common/` | `_csrf`, `_footer`, `_form_buttons`, `_form_error`, `_header`, `_pagination` | Become shared components |
| `login`, `register` | 2 | Guest routes. Special: CSRF/session rotation, §3.2 |
| `dashboard` | 1 | Reads expense + macro stores |
| `expenses` | `index`, `new`, `edit`, `show`, `stats`, `budgets`, `_form`, `_quick_form` | Largest resource. Search, filters, sort, pagination |
| `recurrent_expenses` | `index`, `archived`, `new`, `edit`, `show`, `_form` | |
| `macros` | `index`, `new`, `edit`, `show`, `goals`, `stats`, `_form` | Charts |
| `foods` | `index`, `new`, `edit`, `show`, `_form` | Feeds macro prefill via `?from_food=&amount=` |
| `mood_entries` | `index`, `new`, `edit`, `show`, `stats`, `_form` | Charts |
| `account`, `delete_data`, `exports` | 3 | Destructive posts; file download |
| `error`, `not_found` | 2 | Become client-side error states + a server fallback |


### 2.2 Stimulus controllers (18, `web/static/js/controllers/`)

Each becomes either component-local state or a shared util. None survive as controllers.

`amount`, `chart`, `dashboardDate`, `date`, `dateHelp`, `filter`, `localDate`, `macroCalc`,
`macroDate`, `macroSelect`, `macroTrend`, `moodChart`, `nav`, `quickExpense`, `searchPanel`,
`sort`, `submitOnChange`, `theme`.

Two need decisions rather than a port:
- `theme` — writes `localStorage`, and `layout.html:9-22` has an inline anti-FOUC script that
  must stay server-rendered even in the SPA.
- `chart`/`macroTrend`/`moodChart` — Chart.js instances; in Svelte these become components
  with `$effect` lifecycle. Chart.js stays, it is not the problem being solved here.

### 2.3 Server-side behavior that has no client equivalent yet

| Behavior | Where | What replaces it |
| --- | --- | --- |
| Auth redirect `303 → /login` | `internal/serve/middleware.go:41` | `401` JSON on `/api/*`, §3.1 |
| CSRF token in every form | `common/_csrf.html`, `nosurf` at `middleware.go:79` | `X-CSRF-Token` header, §3.2 |
| Error re-render preserving input | `render.go:renderErr` | Client keeps form state; API returns `422` + field errors |
| `tz_offset` query injection | `web/static/js/index.ts:11-17` (a Turbo event) | Fetch wrapper, §3.6 |
| Pagination / sort / filter via query string | `handlers/shared.go`, `expense_search.go` | Client router owns the query string, API reads it unchanged |
| Icons rendered on `turbo:load` | `index.ts` + `icons.ts` | Per-component icon rendering |
| Loading spinner | Turbo's `.turbo-progress-bar`, restyled in `layout.css` | Own component, §3.7 |
| Export download | `/account/exports/expenses.json`, `data-turbo="false"` | Plain anchor, unchanged |
| 32 `http.Redirect` calls after POST | `internal/handlers/*.go` | API returns the created/updated resource; client navigates |

### 2.4 Styles

`web/static/css/layout.css`, 1450 lines, hand-written, one file. Decision in §4.

### 2.5 Tests

Handler tests currently assert status codes and rendered pages (`handle_*_test.go`, ~14 files).
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

Two traps:
- **Login and logout rotate the session**, and the CSRF token with it. A SPA that never
  reloads keeps sending a dead token. Simplest correct answer: after a successful login or
  logout, do a **full page load** (`window.location.assign(...)`), not a client-side route
  change. One reload per session boundary is a fine price.
- `GET /api/*` must stay side-effect free, since nosurf only guards unsafe methods.

### 3.3 Session

Unchanged. Worth stating explicitly because it is the piece people break "while they are in
there": `SameSite=Lax` is fine — every API call is same-origin — and `Secure` stays tied to
`IsProduction()`. The 5-second request timeout (`WithTimeout`, `routes.go`) applies to API
routes too; long exports already stream.

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

- API returns `422` with `{"error": "...", "fields": {"amount": "must be positive"}}`.
- `500` stays a generic message; do not leak `err.Error()` for unexpected failures.
- Keep the message text identical to today's, so the port is verifiable by eye.

### 3.6 Dates — the part that must not break

**Read this section before touching anything in Expenses, Macros or Moods.** Dates are the
one area where a port can look completely correct, pass every test, and still be wrong by one
day for the user — silently, and only sometimes. Expenses are the app's core feature and every
expense is filed under a calendar date that decides which billing window it lands in. An
off-by-one here does not look like a bug, it looks like the user misremembering.

#### The two kinds of value, which must never be confused

| Kind | Examples | Stored as | Means |
| --- | --- | --- | --- |
| **Calendar date** | `expenses.date`, macro day, mood `logged_at` day | epoch seconds at **UTC midnight** of that date | "the 21st", a label on a calendar. Has no time and no zone |
| **Instant** | `created_at`, `updated_at`, export `exported_at` | epoch seconds of a real moment | a point in time, displayed in the viewer's zone |

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
Moods — `handle_mood_entries.go:344` makes the end exclusive with `AddDate(0, 0, 1)`.

**Half-open `[start, end)` is the convention everywhere. Do not introduce an inclusive end.**

#### Rules for the Svelte port

1. **Format calendar dates with UTC getters only.** `new Date(epoch * 1000).getDate()` returns
   the previous day for any user west of UTC. Every calendar-date render must use `getUTCDate`,
   `getUTCMonth`, `getUTCFullYear` — the same split `localDateController` already makes. Port
   that file's two formatter functions verbatim into one shared module and use it everywhere;
   do not hand-roll formatting per component.
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

#### Retiring `tz_offset` (do this in Phase 3, not before)

`tz_offset` exists only so the *server* can decide which month "this month" means for the
client. In the SPA the client already knows its zone, so it should resolve the named range
itself and send explicit bounds:

```
GET /api/expenses?start=1754006400&end=1756684800     ← client computed, UTC midnight, half-open
```

instead of `?date_range=this_month&tz_offset=-120`. The named ranges (`this_month`,
`last_month`, `six_months`, `this_year`, `this_week`, `next_month`) stay as client-side labels
that resolve to bounds. This deletes `parseTZOffset`, deletes the "did I remember to attach
tz_offset to this fetch" bug class, and stops the server from guessing what "now" means for
someone else.

Server-side, `computeDateRange` becomes bounds validation instead of bounds computation. Keep
the boundary rule it encodes: `start` inclusive, `end` exclusive, both UTC midnight.
`/expenses/budgets` reads its month buckets off the same bounds, so its month/months mode
selection moves client-side with the labels.

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

- **Minification.** Not enabled today. The bundle is already 788 KB unminified and Svelte plus
  a router will not shrink it.
- **A content hash in the filename**, or `/static/*`'s `max-age=300` (`routes.go`) keeps
  serving a stale `index.js` after every deploy for five minutes. Today this only bites during
  development; in an SPA the whole app is that file.

---

## 4. Tailwind — decide separately, and later

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

---

## 5. Phases

Each phase ends with a green `make test`, `make lint-fix`, and a working app. No phase leaves
`main` unshippable. Every phase is one PR.

### Phase 0 — Groundwork (no view changes)

Branch: `spa/phase-0-groundwork`. Nothing a user can see changes in this phase. Its job is to
make every later phase mechanical.

**0.1 Build tooling**
- `bun add svelte bun-plugin-svelte` and `bun add -d vitest @testing-library/svelte`.
- `web/build.ts`: `Bun.build()` with `SveltePlugin`, entry `web/static/js/index.ts`, outdir
  `web/static/js/build`. Point `build-static-js` in the `Makefile` at `bun run web/build.ts`.
- Turn on minification, and emit a content-hashed filename. `/static/*` serves
  `Cache-Control: max-age=300` with no hash today (`internal/serve/routes.go`), so without a
  hash every deploy serves a stale bundle for five minutes — survivable now, not survivable
  once the bundle *is* the app.
- Reference implementation: `web/build.ts` on `experiment/svelte-api`.

**0.2 API route group**
- `/api/*` under its own middleware stack in `internal/serve/routes.go`: session load, body
  limit, timeout, CSRF — but **not** `setTmplData`, and **not** the HTML `AuthMiddleware`.
- API auth middleware: `401` + JSON body, never a `Location` header (§3.1).
- The auth rate limit currently guarding `POST /login` and `POST /register` must guard their
  API equivalents when those land in Phase 6, still sharing **one** middleware value — see the
  invariant in `CLAUDE.md`.

**0.3 JSON handler plumbing**
- Naming: `internal/handlers/handle_api_<resource>.go`, keeping the `handle_` convention.
- One shared JSON write helper and one error mapper: `422` with `{"error", "fields"}` for
  validation, `404`, `500` with a generic message (never `err.Error()` for unexpected
  failures), all in `internal/handlers/` alongside `render.go`.
- Sentinel errors go in the existing `errs.go`, not inline.

**0.4 Client fetch wrapper**
- `web/static/js/api.ts`: `X-CSRF-Token` from `<meta name="csrf-token">`, `401` →
  `window.location.assign("/login")`, JSON error envelope parsing, typed helpers.
- The shell gains the `<meta>` tag in Phase 1; until then read it from `layout.html`, which
  already has `cspNonce` and can carry `csrfToken` the same way.

**0.5 Dates module (see §3.6 before writing it)**
- `web/static/js/dates.ts`: the two formatters from `localDateController.ts` ported verbatim —
  UTC getters for calendar dates, local getters for instants — plus `YYYY-MM-DD` ⇄ epoch
  helpers for the API boundary.
- Tests for both, including the month-boundary and DST cases named in §3.6.

**0.6 Test setup**
- `vitest` + `@testing-library/svelte`, with `TZ=Pacific/Auckland` in the config and a second
  CI run at `TZ=America/Los_Angeles`. Under CI's default `TZ=UTC` every date bug in §3.6
  passes silently, which is the entire reason this is Phase 0 work and not Phase 5 work.
- A `make test-js` target, wired into CI next to `make test`.

**0.7 Documentation**
- `web/README.md`: how the Svelte build works, where components live, the `.svelte` files are
  not covered by prettier/eslint caveat.
- `CLAUDE.md`: the `/api/*` group in the route map.

Exit criteria:
- `make build-static-js`, `make test`, `make test-js` and `make lint-fix` all green.
- A signed-out request to any `/api/*` route returns `401` JSON with no `Location`.
- A signed-in POST to an `/api/*` route without `X-CSRF-Token` is rejected.
- `dates.ts` tests pass under both configured zones.
- The app looks and behaves exactly as it does today.

### Phase 1 — Shell and router, coexisting with templates

Serve the SPA under `/app/*` while every existing route keeps working. Shell template (with the
`<meta name="csrf-token">` tag from §3.2), hand-rolled path-based router, layout chrome
(`_header`, `_footer`, nav) as components, loading indicator, theme handling (the anti-FOUC
inline script in `layout.html:9-22` stays server-rendered and nonce'd), icon rendering.

Exit: `/app` renders the real chrome around a placeholder route, a hard refresh on a nested
path like `/app/foods/1/edit` still resolves, and every template route is untouched.

### Phase 2 — First real resource: Foods

Smallest full CRUD (`index`, `new`, `edit`, `show`, `_form`) with no charts and no pagination
complexity. This phase is where the conventions get set: component layout, form handling,
validation display, list/detail data flow.

Exit: every Foods view works under `/app/foods`, with `422` validation shown inline.

### Phase 3 — Expenses

The hard one, and deliberately early so its surprises land before nine more resources depend on
the conventions: search, filters, sort, pagination, quick-add form, stats, budgets, tags.

This is also the phase that retires `tz_offset` (§3.6): named ranges resolve client-side to
explicit `[start, end)` UTC-midnight bounds, and `computeDateRange` becomes validation instead
of computation.

Exit: `/app/expenses*` matches the template version feature for feature — **verified by running
both versions side by side** against the same data at `TZ=UTC+13` and `TZ=UTC-8`, comparing the
listing, the show page, the stats buckets and the budgets page. While the templates still
exist, they are a working oracle; after Phase 7 they are not.

### Phase 4 — Remaining resources

Recurrent expenses, macros (+ goals, + stats charts), mood entries (+ stats charts), dashboard.
Chart.js components land here. Parallelizable; each resource is its own PR.

Each of these carries its own dated fields — macro day (`handle_macros.go:349`), mood
`logged_at` with its exclusive end (`handle_mood_entries.go:344`), recurrent-expense next-run
dates. Apply §3.6's rules per resource and re-run the two-zone check; do not assume Phase 3
settled it for them.

### Phase 5 — Account, delete-data, exports

Destructive posts and the file download. Confirmation flows must stay at least as deliberate as
today's — a client-side `confirm()` is not equivalent to a form post. Export links stay plain
anchors.

### Phase 6 — Auth views

Login, register, logout. Last because of the CSRF/session rotation trap (§3.2): these are the
only routes where a full page load is the correct behavior.

### Phase 7 — Flip the switch

Move the SPA from `/app/*` to `/`, add the catch-all serving the shell for any non-API,
non-static path, delete every per-view template and `TemplateName` constant, delete the HTML
handlers, drop `@hotwired/turbo` and `@hotwired/stimulus`, delete `web/static/js/controllers/`.

Exit: `web/views/` holds the shell and nothing else.

### Phase 8 — Cleanup and documentation

Rewrite `web/README.md` (its template-partial namespace and Turbo/Stimulus sections become
wrong), update `CLAUDE.md`'s route map, UI/Assets section and the render-helper invariant,
update `docs/architecture.md`'s request flow. Remove dead Go helpers (`render.go`'s page
helpers, `tmplData`, the template loader in `internal/serve/template.go`).

---

## 6. Risks, honestly

- **Duration.** 47 views, 19 controllers, 67 routes. This is months of evenings, not a weekend.
  The phase split exists so that stopping after any phase leaves a working app — Phase 4 is
  where it is safe to pause indefinitely.
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

1. **SPA is staged under `/app/*`** until Phase 7, then moves to `/`. Every phase stays
   revertable by deleting a route.
2. **Router is hand-rolled and path-based.** No router dependency.
3. **`vitest` + `@testing-library/svelte` land in Phase 0**, configured with a non-UTC `TZ` (see
   §3.6) so date bugs fail on the first run rather than in production.
4. **Tailwind is deferred** to a separate post-migration project. Components reference the
   existing `layout.css` class names unchanged.
5. **Auth stays cookie-session.** No JWT, no token in `localStorage`.
6. **`tz_offset` is retired in Phase 3**, replaced by client-computed explicit bounds.

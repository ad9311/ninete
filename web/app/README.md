# `web/app/` — Svelte sources

Sources for the Svelte SPA. Nothing here is served: `setUpFileServer` mounts
`web/static/` verbatim, so sources placed there would be publicly fetchable.
Only build output (`web/static/js/build/`) is served.

Layout and naming rules: `docs/spa-migration.md` §3.9.

- `index.ts` — entry point (Phase 1). Mounts `App.svelte` onto `#app`, the
  mount point `web/views/app/index.html` renders.
- `App.svelte` — the shell's chrome (`Header`, `Footer`, `Spinner`) and the
  router's `$state`-held current path (Phase 1).
- `router.ts` — the hand-rolled path router: `BASE_PATH` (`/app` until Phase 7
  moves it to `/`), the `routes` match table, `matchRoute`/`toRoutePath` (pure),
  and the two DOM listeners (`onPopState`, `onLinkClick`) App.svelte wires up in
  an effect. See `docs/spa-migration.md` §3.7.
- `lib/` — plain `.ts` modules, no components
  - `api.ts` — the `/api/*` fetch wrapper (Phase 0.4). Every request goes
    through it: it attaches `X-CSRF-Token` from the shell's
    `<meta name="csrf-token">` (nosurf's cookie is `HttpOnly`, so JS cannot read
    the token anywhere else), sends a `401` to `/login`, and turns the JSON
    error envelope into an `APIRequestError` carrying `status` and `fields`.
  - `dates.ts` — the three formatters from `localDateController` plus the
    `YYYY-MM-DD` ⇄ epoch helpers (Phase 0.5). `formatDateUTC` is for calendar
    dates and `formatDate`/`formatDateTime` for instants; the two kinds are both
    epoch seconds in an `int64`, so nothing but that split keeps them apart.
    Read `docs/spa-migration.md` §3.6 before touching it.
  - `icons.ts` — moved from `web/static/js/` (Phase 1). Only the Stimulus entry
    point imports it: it is the `createIcons()` DOM scan, which has no
    lifecycle event to hang on in the SPA. The Svelte side imports the icon
    nodes it needs straight from `lucide` and renders them through
    `components/Icon.svelte` (Phase 2 — see below), so this file goes away with
    the last template.
  - `categories.ts` — `fetchCategories()`, wrapping `GET /api/categories`
    (Phase 2). Categories are a shared lookup table (CLAUDE.md), not a
    resource of their own, so this is the whole of it: an id and a name.
  - `currency.ts` — money helpers matching `internal/serve/template_func.go`'s
    `currency` and `amountController.ts`'s cents conversion (Phase 2):
    `formatCurrency` for display, `centsToInputValue`/`inputValueToCents` for
    a form field. Amounts are unsigned cents end to end, never a float.
    `formatCurrency` also covers `signedCurrency`'s job (a budget's negative
    "left" amount) — `Intl`'s currency formatting already prints a leading
    `-` for a negative value, so there is no second formatter (Phase 3).
  - `tags.ts` — `parseTagsInput`/`joinTagNames` for the semicolon-separated tag
    field the templates already use (Phase 2). Normalization (lowercase, trim,
    dedupe) stays server-side in `logic.ParseTagNames`; this only has to get
    the same strings there and back.
  - `dateRanges.ts` — the client-side twin of `computeDateRange`
    (`internal/handlers/expense_shared.go`), ported one for one down to the
    month arithmetic (Phase 3, §3.6 "Retiring `tz_offset` on the API side").
    `computeDateRange(key)` resolves a named range (`this_month`, `six_months`,
    ...) to explicit UTC-midnight `[start, end)` epoch-second bounds using the
    browser's own local calendar — the same role `tz_offset` played
    server-side — and the API only ever receives those bounds, never the key.
    `DATE_RANGE_OPTIONS`/`BUDGET_DATE_RANGE_OPTIONS` are the two option tables
    (`dateRangeLabels`/`budgetDateRanges`) a select needs; the budget table
    also carries each range's month vs. months mode, since the API can no
    longer derive it from a key it never receives.
- `components/` — shared, resource-agnostic components only. Phase 1 adds
  `Header.svelte` (theme switch, session-aware nav dropdown, logout form —
  ports `themeController`/`navController`), `Footer.svelte` (reads the shell's
  `<meta name="version">`) and `Spinner.svelte` (the loading backdrop, §3.7 —
  App.svelte subscribes it to `lib/pending.ts`, which `lib/api.ts` drives; no
  route sets it by hand). `ThemeSwitch` stays inlined in `Header.svelte` rather than its own
  file: it has exactly one caller, and rule 2 below reserves this directory for
  things more than one resource uses. Phase 2 adds `Icon.svelte`: a single
  lucide icon built with `createElement` and swapped into the DOM for the
  placeholder element via an action, replacing `data-lucide` +
  `createIcons()`'s scan (§2.3 of docs/spa-migration.md, "Per-component icon
  rendering") — DOM APIs, not `{@html}`, so §3.4 rule 3's ban never enters it.
  Phase 3 adds two more, both used by more than one resource already:
  `LocalDate.svelte` ports `localDateController.ts`'s two display modes — a
  calendar date with UTC getters, or an instant with local getters and a
  `formatDateTime` title tooltip (§3.6) — and `DateHelp.svelte` ports
  `dateHelpController.ts`'s tap-triggered popover (quick-add's date-format
  help, the expense search panel's date-bounds help), closing on outside
  click or Escape.
- `routes/<resource>/` — mirrors `web/views/<resource>/`, one file per action.
  Phase 1 adds only `Home.svelte`, the placeholder mounted at `/` until Phase 2
  replaces the match table's single entry with a real resource.
  `routes/recurrent_expenses/` (Phase 2, the pilot resource — §7 decision 12)
  is the first full example: `Index.svelte`/`Archived.svelte` are thin
  wrappers around a shared `List.svelte` (one table, parameterized by
  `archived`, since §3.9 rule 1 keeps the two actions as separate files but
  the markup is identical); `New.svelte`/`Edit.svelte` wrap a shared
  `Form.svelte`; `Show.svelte` and `Form.svelte` are their own files;
  `types.ts` mirrors `handle_api_recurrent_expenses.go`'s JSON shape exactly,
  snake_case included (§3.5), so nothing maps between the wire format and what
  a component reads. A resource's query string (filters, sort, pagination)
  reaches its route as a `search` prop — `App.svelte` tracks it separately
  from the matched path, since a page/sort/filter change must not remount the
  routed component the way a real path change does.
  `routes/expenses/` (Phase 3) is the largest example: `List.svelte` adds a
  search panel (description, tag, explicit `date_from`/`date_to`, a
  billed/created toggle) on top of the category and date-range filters
  Phase 2's List introduced, since the named date range now resolves
  client-side via `lib/dateRanges.ts` instead of riding along as
  `date_range`+`tz_offset`; `Form.svelte` adds a calendar-date field
  (`lib/dates.ts`'s `calendarDateToUnix`/`todayCalendarDate`); `New.svelte`
  toggles between it and `QuickAddForm.svelte`, which posts to
  `/expenses/quick` with an explicit `tz_offset` (§3.6's "Consumer 2" — quick
  add keeps a client zone even though the named ranges retire theirs) and
  shows a category picker on the first `category_id` field error, mirroring
  the template's re-render-to-ask-for-a-category flow without a page
  reload; `Stats.svelte` and `Budgets.svelte` are their own routes, the
  latter sending the `mode` (`month`/`months`) explicitly since the API
  no longer receives the range key it used to derive that from.
- `toolchain/` — not part of the app. `Probe.svelte` and its test are a canary
  for the test setup itself (Phase 0.6): they fail when vitest can no longer
  compile a component, when it resolves Svelte's server build instead of the
  client one, or when the jsdom opt-in stops working.

The directory is bundled from Phase 1 onward (`web/build.ts` adds `index.ts` as
a second entry point, alongside the Stimulus one); `lib/dates.ts` was
unit-tested without a bundle since Phase 0.5.

## Tests

Each test sits beside what it tests and takes its name — `dates.ts` /
`dates.test.ts`, `Probe.svelte` / `Probe.test.ts` (§3.9 rule 5). Run them with
`make test-js`, which runs the suite twice
— `Pacific/Auckland`, then `America/Los_Angeles`. Both signs of UTC offset are
needed: a calendar date formatted with local getters still reads correctly east
of UTC and only breaks west of it. `bun run test:js` runs the default zone
alone, and `TEST_TZ` overrides it. `TEST_TZ` rather than `TZ` so the developer's
own zone cannot silently replace it.

The default environment is Node, not jsdom, because `lib/` holds no components
and a module there that reaches for `document` should fail its own test rather
than pass because a DOM happened to be present. A component test opts in with
`// @vitest-environment jsdom` on the first line of the file.

`make lint-fix` covers `.svelte`: `prettier-plugin-svelte` formats the whole
file, and eslint runs `svelte-eslint-parser` with the TS parser nested inside
it, so both `@typescript-eslint` and the 37 `eslint-plugin-svelte` rules apply.
`processor: "svelte/svelte"` is set as well; it is not what makes the rules
fire, it is what makes `<!-- eslint-disable-next-line svelte/... -->` inside
markup be honoured. Rune modules (`*.svelte.js`, `*.svelte.ts`) get the same
treatment through their own config block.

Type checking is separate from linting and takes two tools, because `tsc` cannot
parse a component at all. `bun run typecheck:ts` (`tsc --noEmit`) covers `.ts`;
`bun run typecheck:svelte` (`svelte-check`) covers `.svelte`. Both run in
`make lint-fix` and in the `typecheck` CI job, and both were checked against a
deliberate type error rather than a clean exit: an error planted in `Probe.svelte`
is reported by `svelte-check` and passes `tsc` silently, which is the whole reason
for the second tool. `tsconfig.json` lists `web/app/**/*.svelte` in `include` —
`tsc` ignores it (not an extension it compiles) and `svelte-check` reads it, so
the component file set is written down rather than inferred.

The API boundary is what this really guards. A response crosses Go → JSON →
TypeScript with no compiler on either side of the wire, so a renamed field in a
Go struct arrives as `undefined` in a component and nothing else in the chain
notices.

One gap: stylelint reads `web/static/css/**/*.css` only and has no Svelte
processor, so a `<style>` block inside a component is formatted but not
stylelinted. A handful of `eslint-plugin-svelte` rules cover part of that
ground (`svelte/no-dupe-style-properties`,
`svelte/no-unknown-style-directive-property`). No component has a `<style>`
block yet — §5 decision 5 keeps them on `layout.css` class names — so wire
stylelint up if that changes rather than before.

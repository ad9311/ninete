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
  - `icons.ts` — moved from `web/static/js/` (Phase 1). Both entry points
    import it; the Svelte side reaches it through `components/Icon.svelte`
    rather than `createIcons()`'s DOM scan (Phase 2 — see below).
  - `categories.ts` — `fetchCategories()`, wrapping `GET /api/categories`
    (Phase 2). Categories are a shared lookup table (CLAUDE.md), not a
    resource of their own, so this is the whole of it: an id and a name.
  - `currency.ts` — money helpers matching `internal/serve/template_func.go`'s
    `currency` and `amountController.ts`'s cents conversion (Phase 2):
    `formatCurrency` for display, `centsToInputValue`/`inputValueToCents` for
    a form field. Amounts are unsigned cents end to end, never a float.
  - `tags.ts` — `parseTagsInput`/`joinTagNames` for the semicolon-separated tag
    field the templates already use (Phase 2). Normalization (lowercase, trim,
    dedupe) stays server-side in `logic.ParseTagNames`; this only has to get
    the same strings there and back.
- `components/` — shared, resource-agnostic components only. Phase 1 adds
  `Header.svelte` (theme switch, session-aware nav dropdown, logout form —
  ports `themeController`/`navController`), `Footer.svelte` (reads the shell's
  `<meta name="version">`) and `Spinner.svelte` (the router-level pending flag,
  §3.7). `ThemeSwitch` stays inlined in `Header.svelte` rather than its own
  file: it has exactly one caller, and rule 2 below reserves this directory for
  things more than one resource uses. Phase 2 adds `Icon.svelte`: a single
  lucide icon built with `createElement` and swapped into the DOM for the
  placeholder element via an action, replacing `data-lucide` +
  `createIcons()`'s scan (§2.3 of docs/spa-migration.md, "Per-component icon
  rendering") — DOM APIs, not `{@html}`, so §3.4 rule 3's ban never enters it.
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

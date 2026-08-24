# `web/app/` — Svelte sources

Sources for the Svelte SPA. Nothing here is served: `setUpFileServer` mounts
`web/static/` verbatim, so sources placed there would be publicly fetchable.
Only build output (`web/static/js/build/`) is served.

Layout and naming rules: `docs/spa-migration.md` §3.9.

- `index.ts` — entry point (added in Phase 1)
- `App.svelte`, `router.ts` — shell and router (Phase 1)
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
  - `icons.ts` — moved from `web/static/js/` (Phase 1)
- `components/` — shared, resource-agnostic components only
- `routes/<resource>/` — mirrors `web/views/<resource>/`, one file per action
- `toolchain/` — not part of the app. `Probe.svelte` and its test are a canary
  for the test setup itself (Phase 0.6): they fail when vitest can no longer
  compile a component, when it resolves Svelte's server build instead of the
  client one, or when the jsdom opt-in stops working. Nothing else imports them

The directory is not bundled until Phase 1; `lib/dates.ts` is unit-tested
without a bundle from Phase 0.5.

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

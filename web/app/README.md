# `web/app/` — Svelte sources

Sources for the Svelte SPA. Nothing here is served: `setUpFileServer` mounts
`web/static/` verbatim, so sources placed there would be publicly fetchable.
Only build output (`web/static/js/build/`) is served.

Layout and naming rules: `docs/spa-migration.md` §3.9.

- `index.ts` — entry point (added in Phase 1)
- `App.svelte`, `router.ts` — shell and router (Phase 1)
- `lib/` — plain `.ts` modules, no components (`api.ts`, `dates.ts`, `icons.ts`)
- `components/` — shared, resource-agnostic components only
- `routes/<resource>/` — mirrors `web/views/<resource>/`, one file per action

The directory is not bundled until Phase 1; `lib/dates.ts` is unit-tested
without a bundle from Phase 0.5.

# `web/` — Shell Template, Static Assets and Svelte Sources

Reference for the frontend half of NINETE. `CLAUDE.md` at the repo root covers conventions and the invariants that must not be broken, and `docs/architecture.md` covers layering and the Go-side request flow; this file covers what you need before editing anything under `web/`.

Three directories:

- `web/views/` — exactly one Go `html/template` file: the SPA shell.
- `web/static/` — CSS, the built JS bundle, and images, served from `/static/*`.
- `web/app/` — Svelte sources for the SPA. **Never served.** Build output only ever lands in
  `web/static/js/build/`.

There are no rendered pages: the SPA is served from `/`, and `web/views/` holds the shell alone.
`docs/spa-migration.md` records how it got that way. `web/app/README.md` is the reference for the
Svelte sources themselves — layout, naming, what goes in `lib/` vs. `components/` vs. `routes/`.

---

## `web/views` — The shell template

`web/views/app/index.html` is the only file left under `web/views`. It carries its own `<html>`
document — there is no shared chrome and no partials any more, unlike the templates it replaced.
It still opens with `{{ define "layout" }}`, and must: `parseTemplates` executes every view by
that name, so removing the define renders an empty page and returns a 500 at request time, not at
startup. The define is a template *name*, not the shared layout the SPA removed.

It is still parsed and executed through the same machinery every other template used to go
through (`internal/serve/template.go`'s `LoadTemplates`/`parseTemplates`,
`internal/handlers/render.go`'s `render`/`renderPage`/`tmplData`), because it still needs
per-request values injected server-side that a static file can't carry:

| Key | Meaning |
| --- | --- |
| `cspNonce` | Nonce for the inline anti-FOUC `<script>` |
| `csrfToken` | Read once by `web/app/lib/api.ts` from the `<meta name="csrf-token">` tag it renders into — nosurf's own cookie is `HttpOnly`, so JS has no other way to see it |
| `version` | Build stamp, rendered into a `<meta name="version">` tag |
| `appBundle` | The content-hashed `/static/js/build/app-<hash>.js` path, read from the asset manifest (`internal/serve/manifest.go`) rather than hardcoded |

`GetApp` (`internal/handlers/handle_app.go`) is the only render call left in the codebase. It
backs the catch-all at `/`, answering every non-API, non-static path with the same document so a
hard refresh on a nested route like `/expenses/12/edit` resolves — the client router
(`web/app/router.ts`) owns the rest of the path. See the "Render helpers need setTmplData"
invariant in `CLAUDE.md` for why `GetApp` has to sit inside the app middleware group rather than
the root router.

### Content Security Policy

The inline anti-FOUC `<script>` in the shell **must** carry `nonce="{{ .cspNonce }}"`. Without it
the browser blocks the tag and posts a violation report to `/csp-report`. See
`docs/spa-migration.md` §3.4 before adding any other inline script or style — Svelte 5 compiles to
plain JS and needs no CSP relaxation, and none should be added.

### Editing loop

The shell is parsed once at boot. In development, `render` re-parses it at most once every 2
seconds, so an edit appears on the next load without recompiling the binary.

### Adding a view

There is one view today and the SPA is not expected to grow a second. If one is ever added, two
constraints decide whether it is reachable, and neither fails at build time:

- **It must sit exactly one directory below `web/views`.** The glob in `template.go` uses `**`,
  which Go's `filepath.Glob` treats as a single path segment, not a recursive match. A view placed
  directly in `web/views`, or nested two levels deep, is never parsed —
  `web/views/app/index.html` is at the depth that works.
- **It needs a matching `handlers.TemplateName` constant.** Nothing checks the two agree at
  compile time; a mismatch logs `missing template` and returns a 500 at request time.

---

## `web/static` — CSS, built JS, images

Served from `/static/*`, which is mounted on the root router **outside** the app middleware
chain. Serving an asset must never load a session or query the database — see the "Static assets
stay off the app chain" invariant in `CLAUDE.md` before changing how these are served.

Responses carry `Cache-Control: public, max-age=300`. The JS bundle's filename is content-hashed
(see "The build" below) and so is safe from a stale cache regardless, but `layout.css` and the
images are not — the short window is what keeps a deploy from serving a stale stylesheet, so do
not raise it without hashing those too.

- `web/static/css/layout.css` — one hand-written stylesheet for the whole app. Design tokens are
  custom properties on `:root`. Theming swaps a `theme-light`/`theme-dark` class on `<html>`; the
  shell's inline nonce'd script sets it before first paint, and `web/app`'s theme handling
  persists the choice in `localStorage`. Linted with `stylelint` via `make lint-fix`.
- `web/static/js/build/` — the generated bundle and its manifest. Git-ignored. Never edited by
  hand — see "The build" below.
- `web/static/img/` — currently just `favicon.ico`, referenced by the shell.

---

## `web/app` — Svelte sources

Layout and naming rules live in `docs/spa-migration.md` §3.9 and are documented in
`web/app/README.md`; this section covers how the code gets from here into the browser.

### Why sources sit outside `web/static/`

`setUpFileServer` mounts the whole of `web/static/` verbatim
(`http.FileServer(http.Dir("./web/static/"))`), so **everything under it is publicly fetchable**.
Nothing secret lives in a component, so this is tidiness rather than a vulnerability, but there is
no reason to ship sources once a build step exists. Only build output belongs under
`web/static/`.

### The build

`web/build.ts` drives `Bun.build()` with `bun-plugin-svelte`, since Svelte components need a
plugin the `bun build` CLI has no flag for.

```
web/app/**                            sources, the single entry point (web/app/index.ts)
       ↓  bun run web/build.ts   (make build-static-js)
web/static/js/build/app-<hash>.js     minified bundle, git-ignored, served from /static/*
web/static/js/build/manifest.json     {"app": "app-<hash>.js"}
```

- **No component carries a `<style>` block, and the build assumes it.** Styling lives in
  `web/static/css/layout.css` (§5 decision 5, `web/app/README.md`), which the shell links
  directly. A `<style>` block would make `bun-plugin-svelte` emit a sibling
  `app-<hash>.css` that nothing links — and that `web/build.ts` then prunes, since its `keep`
  set holds only the manifest's values. Adding component styles therefore means three edits at
  once: record the CSS output in the manifest, link it from the shell, and wire `stylelint` up
  to reach it.

- **The bundle is git-ignored.** Run `make build-static-js` after editing any `.ts` or `.svelte`;
  `make dev` does it as part of its build, and the test suite depends on it because
  `internal/serve` asserts `/static/*` serves the bundle.
- **A build keeps the previous generation.** `web/build.ts` prunes the output directory *after*
  writing, keeping the filenames the outgoing `manifest.json` named as well as the new ones.
  `scripts/deploy.sh` builds the JS into the live checkout while the previous binary is still
  serving pages that name the old hashes, and only restarts the service at the end — deleting
  those files up front would answer every request in that window with a 404 for the bundle.
- **Filenames are content-hashed.** `web/build.ts` writes `manifest.json` beside the bundle;
  `internal/serve/manifest.go` reads it at startup (`LoadAssetManifest`, called next to
  `LoadTemplates`, and again from the development template-reload hook so a rebuild's new hash is
  picked up without a restart) and `setTmplData` puts the resolved `/static/*` path into the
  template map (`appBundle`) for the shell to read. The shell never hardcodes a filename, and
  `routes_test.go` reads the manifest rather than asserting a literal path. A missing manifest is
  not always a boot error: `LoadAssetManifest` treats it the way `parseTemplates` treats an empty
  views glob — a package testing far enough from the repo root to have no `web/` tree at all
  (`internal/logic`, `internal/repo`) gets an empty manifest rather than a hard failure, since
  those tests never render a page or ask for a bundle path. A manifest that exists but fails to
  parse is still an error.

### Tests and lint

`make test-js`, not `make test` — the Go suite does not run the frontend tests. `make lint-fix`
covers `.svelte` as well as `.ts`.

Both have details that matter and are documented once, in **`web/app/README.md`**: why the tests
run in two time zones, what lint does and does not reach, and where test files live. Read that
rather than assuming from here.

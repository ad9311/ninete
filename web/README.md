# `web/` — Shell Template, Static Assets and Svelte Sources

Reference for the frontend half of NINETE. `CLAUDE.md` at the repo root covers conventions and the invariants that must not be broken, and `docs/architecture.md` covers layering and the Go-side request flow; this file covers what you need before editing anything under `web/`.

Three directories:

- `web/views/` — exactly one Go `html/template` file: the SPA shell.
- `web/static/` — build output (the JS bundle and the stylesheet) and images, served from
  `/static/*`.
- `web/app/` — Svelte and CSS sources for the SPA. **Never served.** Build output only ever lands
  in `web/static/js/build/` and `web/static/css/build/`.

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
| `appStylesheet` | The content-hashed `/static/css/build/css-<hash>.css` path, from the same manifest |

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

## `web/static` — build output and images

Served from `/static/*`, which is mounted on the root router **outside** the app middleware
chain. Serving an asset must never load a session or query the database — see the "Static assets
stay off the app chain" invariant in `CLAUDE.md` before changing how these are served.

Responses carry `Cache-Control: public, max-age=300`. The bundle and the stylesheet both carry a
content hash (see "The build" below) and so are safe from a stale cache regardless; the images
are not, and the short window is what keeps a deploy from serving a stale one, so do not raise it
without hashing those too.

- `web/static/css/build/` — the generated stylesheet. Git-ignored, never edited by hand. **The
  source is `web/app/app.css`**, and it is the only stylesheet in the repository — see "Styling"
  below.
- `web/static/js/build/` — the generated bundle. Git-ignored. Never edited by hand — see "The
  build" below.
- `web/static/manifest.json` — the asset manifest both builds write, mapping an entry name to the
  public path of the hashed file. Git-ignored; read by `internal/serve/manifest.go` at startup.
- `web/static/img/` — currently just `favicon.ico`, referenced by the shell.

---

## Styling

**One stylesheet, `web/app/app.css`, and it is a Tailwind v4 source file.** There is no
`tailwind.config.js` and there must not be one: v4 is CSS-first, so the tokens (`@theme`), the
dark variant (`@custom-variant`) and the component layer all live in that file. It replaced
`web/static/css/layout.css` — 1373 hand-written lines, most of them single-use layout rules —
in September 2026; `docs/spa-migration.md` §4.1 records why the change waited until the SPA
migration had finished.

Three things about it decide how a component should be written:

- **Colours are roles, not values, and the theme owns them.** `:root` and `html.theme-dark` each
  define the full set of `--ui-*` variables, and `@theme` maps them onto Tailwind colours. So
  `bg-surface`, `text-muted` and `border-line` follow the theme on their own. **A `dark:` utility
  in a component means a role is missing** — add the role to `app.css` rather than pairing
  colours at the call site. Theming still works the way it always did: a `theme-light`/
  `theme-dark` class on `<html>`, written by the shell's inline nonce'd script before first paint
  and persisted to `localStorage` by `components/Header.svelte`.
- **The base layer styles form controls and document defaults, and utilities beat it.** An
  `<input>` gets the app's control styling with no class; the header's compact selects and the
  search panel's hidden checkbox override it with plain utilities, because Tailwind's utility
  layer sits after the base layer. There is nothing to opt out of.
- **The component layer is short on purpose.** `.btn*`, `.chip*`, `.data-table`, `.page-link*`,
  `.toggle-switch`, `.budget-bar` and `.route-progress-bar` are there because more than one route
  uses them *and* utilities cannot express them in markup without drifting — a vendor
  pseudo-element, a `::before` ring, a `:checked ~` sibling. Everything else is utilities in the
  component that needs it. `web/app/README.md` carries the rule for adding to either layer.

`stylelint` reads `web/app/**/*.css` (that is, this one file) via `make lint-fix`, with
Tailwind's at-rules allowlisted in `stylelint.config.mjs`. `prettier-plugin-tailwindcss` sorts
class attributes, so utility order in markup is canonical and not worth arguing about in review.

The `tailwindcss` package in `devDependencies` is there for that Prettier plugin, which needs to
resolve the stylesheet to sort against. **The build does not use it**: `bun-plugin-tailwind` (a
runtime dependency, beside `bun-plugin-svelte`) carries its own copy of Tailwind and the native
`oxide` binaries. Both are pinned to exact versions rather than a caret range, because the copy
inside the plugin is what compiles and the copy in `devDependencies` is what decides class order
— a range on either lets the two drift onto different Tailwind versions silently. Bump them
together, and check `bun-plugin-tailwind`'s embedded version (the banner comment at the top of
the built CSS) still matches.

---

## `web/app` — Svelte and CSS sources

Layout and naming rules live in `docs/spa-migration.md` §3.9 and are documented in
`web/app/README.md`; this section covers how the code gets from here into the browser.

### Why sources sit outside `web/static/`

`setUpFileServer` mounts the whole of `web/static/` verbatim
(`http.FileServer(http.Dir("./web/static/"))`), so **everything under it is publicly fetchable**.
Nothing secret lives in a component, so this is tidiness rather than a vulnerability, but there is
no reason to ship sources once a build step exists. Only build output belongs under
`web/static/`.

### The build

`web/build.ts` drives `Bun.build()`, with `bun-plugin-svelte` for the components and `bun-plugin-tailwind`
for the stylesheet — two entry points, one per language, because neither plugin has a `bun build`
CLI flag.

```
web/app/index.ts + web/app/**         SPA sources
web/app/app.css                       the stylesheet source
       ↓  bun run web/build.ts   (make build-static)
web/static/js/build/app-<hash>.js     minified bundle, git-ignored, served from /static/*
web/static/css/build/css-<hash>.css   compiled stylesheet, git-ignored, served from /static/*
web/static/manifest.json              {"app": "/static/js/build/app-<hash>.js",
                                       "css": "/static/css/build/css-<hash>.css"}
```

- **The manifest holds whole public paths, not filenames.** The two entries build into different
  directories and `web/build.ts` is the only place that decides which; Go looks the path up and
  hands it to the shell rather than rebuilding it from a prefix it would have to keep in sync.

- **No component carries a `<style>` block, and the build assumes it.** Styling is utilities in
  the markup plus `web/app/app.css`, which the shell links. A `<style>` block would make
  `bun-plugin-svelte` emit a second, differently-named CSS file that nothing links — and that
  `web/build.ts` then prunes, since its live set holds only the manifest's values. It would also
  escape `stylelint`, which has no Svelte processor. Put the rule in `app.css`'s component layer,
  or express it as utilities.

- **The build output is git-ignored.** Run `make build-static` after editing any `.ts`,
  `.svelte` or `.css`;
  `make dev` does it as part of its build, and the test suite depends on it because
  `internal/serve` asserts `/static/*` serves both.
- **A build keeps the previous generation.** `web/build.ts` prunes the output directory *after*
  writing, keeping the filenames the outgoing `manifest.json` named as well as the new ones.
  `scripts/deploy.sh` builds into the live checkout while the previous binary is still serving
  pages that name the old hashes, and only restarts the service at the end — deleting those files
  up front would answer every request in that window with a 404 for the bundle and an unstyled
  page. Both output directories are pruned, each against its own generation.
- **Filenames are content-hashed.** `web/build.ts` writes `manifest.json` under `web/static/`;
  `internal/serve/manifest.go` reads it at startup (`LoadAssetManifest`, called next to
  `LoadTemplates`, and again from the development template-reload hook so a rebuild's new hash is
  picked up without a restart) and `setTmplData` puts the resolved `/static/*` paths into the
  template map (`appBundle`, `appStylesheet`) for the shell to read. The shell never hardcodes a
  filename, and `routes_test.go` reads the manifest rather than asserting a literal path. A missing manifest is
  not always a boot error: `LoadAssetManifest` treats it the way `parseTemplates` treats an empty
  views glob — a package testing far enough from the repo root to have no `web/` tree at all
  (`internal/logic`, `internal/repo`) gets an empty manifest rather than a hard failure, since
  those tests never render a page or ask for a bundle path. A manifest that exists but fails to
  parse is still an error.

### Tests and lint

`make test-js`, not `make test` — the Go suite does not run the frontend tests. `make lint-fix`
covers `.svelte` and `.css` as well as `.ts`.

Both have details that matter and are documented once, in **`web/app/README.md`**: why the tests
run in two time zones, what lint does and does not reach, and where test files live. Read that
rather than assuming from here.

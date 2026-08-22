# `web/` — Templates and Static Assets

Reference for the frontend half of NINETE. `CLAUDE.md` at the repo root covers conventions and the invariants that must not be broken, and `docs/architecture.md` covers layering and the Go-side request flow; this file covers what you need before editing anything under `web/`.

Two directories:

- `web/views/` — Go `html/template` files rendered server-side.
- `web/static/` — CSS, the TypeScript bundle, and images, served from `/static/*`.

---

## `web/views` — Templates

Parsed by `internal/serve/template.go`. Route definitions live in `internal/serve/routes.go` and remain the source of truth for what URL reaches which handler.

### Layout

`web/views/layout.html` defines `layout` and is the only file containing `<html>`. It renders the `header` partial, a `{{ block "main" . }}`, and the `footer` partial.

Every page view is the same two lines of scaffolding around its content:

```
{{ template "layout" . }}
{{ define "main" }}
  …page content…
{{ end }}
```

### Naming and registration

Path pattern is `web/views/<resource>/<action>.html`.

The template key is `<directory>/<basename>` (`viewKey` in `template.go`), which is exactly the value of the matching `handlers.TemplateName` constant:

```
web/views/foods/edit.html  ⇒  "foods/edit"  ⇒  handlers.FoodsEdit
```

**Adding a view means two edits**: the file itself, and a `TemplateName` constant in `internal/handlers/constants.go`. Nothing checks these agree at compile time — a handler rendering a name with no template logs `missing template` and returns a 500.

**The resource directory is not optional.** Views are globbed with `./web/views/**/*.html`, and Go's `filepath.Glob` treats `**` as a single path segment rather than a recursive match — so the pattern means exactly one directory below `web/views`. A template at `web/views/foo.html`, or nested two levels deep at `web/views/a/b/foo.html`, is silently never parsed: no error at boot, just a 500 the first time a handler asks for it. `layout.html` sits at the top level only because `parseTemplates` globs it separately.

### Partials

Files named `_*.html` **in a resource directory under `web/views`** are parsed into the shared base template. Partials are globbed with `./web/views/**/_*.html`, so the one-level rule above applies to them too: a partial at `web/views/_foo.html` matches neither the partials glob nor the views glob, and every page using it fails at render.

Being parsed into one shared base means **every partial's `define` name lives in one global namespace**, and a duplicate name silently overwrites another page's partial.

Names currently in use:

| Name | File | Purpose |
| --- | --- | --- |
| `header` | `common/_header.html` | Site header, rendered by the layout |
| `footer` | `common/_footer.html` | Site footer carrying the build stamp, rendered by the layout |
| `csrf` | `common/_csrf.html` | Hidden CSRF field for forms |
| `form_error` | `common/_form_error.html` | Renders `.error` |
| `submit_button`, `delete_button` | `common/_form_buttons.html` | Shared form buttons |
| `pagination` | `common/_pagination.html` | Pager controls |
| `expense_form`, `food_form`, … | `<resource>/_form.html` | Per-resource form bodies |

Cross-resource partials go in `web/views/common/`; resource-specific ones sit next to the views that use them.

### Data contract

Handlers render a `map[string]any` built by `h.tmplData(r)`. It always carries:

| Key | Meaning |
| --- | --- |
| `csrfToken` | Token for the `csrf` partial |
| `cspNonce` | Nonce for inline `<script>` / `<style>` |
| `error` | Empty string unless an error path set it |
| `isUserSignedIn` | Auth state |
| `currentUser` | `*logic.User`, nil for guests |
| `version` | Build stamp (`prog.Version`), rendered by the `footer` partial and the `version` meta tag |

Handlers add their own keys on top. A listing page typically adds its rows, `categories`, `pagination` and `basePath`.

`renderErr` sets `error` and re-renders the same page rather than redirecting, so forms keep the user's input — which is why form templates read their values back from the data map instead of relying on the browser.

### Content Security Policy

Any inline `<script>` or `<style>` **must** carry `nonce="{{ .cspNonce }}"`. Without it the browser blocks the tag and posts a violation report to `/csp-report`.

Assets from other origins (CDN scripts, Google Fonts, remote images) are blocked by the policy in `internal/serve/middleware.go`. Anything new has to be vendored into `web/static/`.

### Template functions

Registered in `internal/serve/template_func.go`:

| Function | Purpose |
| --- | --- |
| `currency` | `uint64` cents ⇒ `$1,234.56`. Money is stored in cents — never format it by hand |
| `signedCurrency` | Same, for an `int64` that can go negative. Stored amounts are unsigned, so this is only for derived figures such as a budget's remaining amount |
| `truncateFloat` | Trims a float for display |
| `timeStamp` | Unix seconds ⇒ `YYYY-MM-DD` |
| `sumAmount`, `sumTotal` | Totals over a slice of rows, by `Amount` / `Total` field |
| `sortURL` | Column-header link that flips sort order and preserves current filters |
| `pageURL`, `pageRange` | Pagination links, and the window of page numbers to show |
| `filterURL` | Link that changes one filter key and keeps the rest |
| `dateRangeOptions`, `perPageChoices` | Option lists for the range and page-size selects |
| `add`, `sub`, `titleize` | Small helpers |

Filter, sort and pagination state travels in `handlers.PaginationData`, and the URL helpers above rebuild query strings from it. A new filter therefore needs a field on that struct — do not hand-write query strings in templates, or the other links will drop the new parameter.

### Editing loop

Templates are parsed once at boot. In development, `render` re-parses them at most once every 2 seconds, so a template edit appears on the next page load without recompiling the binary. Navigation is Turbo-driven, so refresh the browser to be sure you are seeing a fresh render.

A template syntax error surfaces as a 500 with `ERROR EXECUTING TEMPLATE` in the response and the real reason in the server log.

---

## `web/static` — CSS, JS, images

Served from `/static/*`, which is mounted on the root router **outside** the app middleware chain. Serving an asset must never load a session or query the database — see the "Static assets stay off the app chain" invariant in `CLAUDE.md` before changing how these are served.

Responses carry `Cache-Control: public, max-age=300`. The window is short because the filenames are not content-hashed; a longer one would strand a stale bundle in the browser after a deploy.

### CSS — `web/static/css/layout.css`

One stylesheet for the whole app. Design tokens are custom properties on `:root`.

Theming swaps a `theme-light` / `theme-dark` class on `<html>`. An inline, nonce'd script in the layout sets it before first paint (so there is no flash), and `themeController` toggles it afterwards, persisting the choice in `localStorage`.

Linted with `stylelint` via `make lint-fix`.

### JS — `web/static/js/`

TypeScript, bundled by Bun. Dependencies: `@hotwired/turbo` (navigation), `@hotwired/stimulus` (behavior), `chart.js` (stats charts), `lucide` (icons).

```
index.ts              entrypoint: starts Turbo + Stimulus, registers controllers
controllers/*.ts      one Stimulus controller per file
icons.ts              lucide initialization
global.d.ts           window.Stimulus typing
build/index.js        generated bundle — git-ignored
```

Things that are easy to get wrong:

- **A new controller is inert until registered in `index.ts`**, which maps a kebab-case identifier used in markup (`data-controller="quick-expense"`) to a camelCase file in `controllers/`.
- **`index.ts` appends `tz_offset` to every Turbo fetch request.** That is how the server learns the browser's timezone; `parseTZOffset` reads it for all date-range math. A request that does not go through Turbo has no offset and falls back to UTC.
- **Icons initialize on both `turbo:load` and `turbo:render`.** The second listener is required: form re-renders, including non-2xx error responses, do not fire `turbo:load`, and `<i data-lucide>` elements would stay unconverted and invisible.
- **The loading spinner is Turbo's progress-bar element restyled, not an overlay of ours.** Turbo creates `.turbo-progress-bar`, shows it once a visit or form submission has been in flight for `Turbo.config.drive.progressBarDelay` (lowered from Turbo's 500 ms default to 250 ms in `index.ts`), and removes it when the navigation ends; `layout.css` turns that element into a full-viewport backdrop with a centred spinner drawn as its `::before`. Because the timing stays inside Turbo's own visit lifecycle, cached-snapshot previews, hover prefetches and aborted visits are all handled, and the element being created per show means the spin animation starts from 0 every time. Do not rebuild this as a Stimulus controller driving your own overlay: Turbo replaces `<body>` on every render, so an element-scoped controller loses its pending timers mid-navigation, and a cached revisit renders its preview before the delay is up. That was tried and reverted. Anything that opts out of Turbo (`data-turbo="false"`, such as the export download) gets no spinner; per-button `data-turbo-submits-with` text still applies on top.
- **The bundle is generated and git-ignored.** Run `make build-static-js` after editing any `.ts`; `make dev` does it as part of its build.

Linted with `eslint`, formatted with `prettier` (the `prettier-plugin-go-template` plugin also formats `.html` templates), both via `make lint-fix`.

### Images — `web/static/img/`

Currently just `favicon.ico`, referenced by the layout.

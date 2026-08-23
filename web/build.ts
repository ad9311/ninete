// Bundles the frontend entrypoints. This is a script rather than a plain
// `bun build` command line because Svelte components need a bundler plugin,
// and the CLI has no flag for one.
//
// Entry points during the SPA migration (docs/spa-migration.md §3.9):
// Phase 1 adds `web/app/index.ts` beside the Stimulus entry below, and Phase 7
// removes the Stimulus one. Sources live outside `web/static/`, which is served
// verbatim; only the build output belongs under it.
import { SveltePlugin } from "bun-plugin-svelte";

const result = await Bun.build({
  entrypoints: ["web/static/js/index.ts"],
  outdir: "web/static/js/build",
  target: "browser",
  minify: true,
  plugins: [SveltePlugin({ development: false })],
});

if (!result.success) {
  for (const log of result.logs) {
    console.error(log);
  }
  process.exit(1);
}

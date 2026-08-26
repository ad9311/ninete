// Bundles the frontend entrypoints. This is a script rather than a plain
// `bun build` command line because Svelte components need a bundler plugin,
// and the CLI has no flag for one.
//
// Entry points during the SPA migration (docs/spa-migration.md §3.9): the
// Stimulus entry stays until Phase 7 removes it; Phase 1 adds the Svelte one
// beside it. Sources live outside `web/static/`, which is served verbatim;
// only the build output belongs under it.
//
// Filenames carry a content hash (Phase 1, deferred from 0.1): each entry
// builds separately so its output name is unambiguous, and the two filenames
// land in manifest.json for Go to read at startup and hand to the templates
// through setTmplData. Without the hash, /static/*'s max-age=300 leaves a
// deploy serving a stale bundle for up to five minutes.
import { rm } from "node:fs/promises";
import { basename } from "node:path";
import { SveltePlugin } from "bun-plugin-svelte";

const outdir = "web/static/js/build";

// Each run's filenames carry a fresh content hash, so a stale one from a
// previous build would otherwise sit alongside the new one, unreferenced but
// undeleted, forever.
await rm(outdir, { recursive: true, force: true });

interface Entry {
  /** Manifest key, and the filename prefix before the content hash. */
  name: string;
  path: string;
}

const entries: Entry[] = [
  { name: "index", path: "web/static/js/index.ts" },
  { name: "app", path: "web/app/index.ts" },
];

const manifest: Record<string, string> = {};

for (const entry of entries) {
  const result = await Bun.build({
    entrypoints: [entry.path],
    outdir,
    target: "browser",
    minify: true,
    naming: `${entry.name}-[hash].[ext]`,
    plugins: [SveltePlugin({ development: false })],
  });

  if (!result.success) {
    for (const log of result.logs) {
      console.error(log);
    }
    process.exit(1);
  }

  const entryPoint = result.outputs.find(
    (output) => output.kind === "entry-point",
  );
  if (!entryPoint) {
    console.error(`${entry.path}: build produced no entry-point output`);
    process.exit(1);
  }

  manifest[entry.name] = basename(entryPoint.path);
}

await Bun.write(`${outdir}/manifest.json`, JSON.stringify(manifest, null, 2));

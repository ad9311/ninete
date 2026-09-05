// Bundles the frontend. This is a script rather than a plain `bun build`
// command line because both entry points need a plugin the CLI has no flag
// for: Svelte components for the JS bundle, Tailwind for the stylesheet.
//
// Two entry points, one per language: `web/app/index.ts` (the SPA, §3.9 of
// docs/spa-migration.md) and `web/app/app.css` (the whole stylesheet). Sources
// live outside `web/static/`, which is served verbatim; only build output
// belongs under it.
//
// Filenames carry a content hash: each entry builds separately so its output
// name is unambiguous, and the paths land in manifest.json for Go to read at
// startup and hand to the shell through setTmplData. Without the hash,
// /static/*'s max-age=300 leaves a deploy serving a stale bundle — or a stale
// stylesheet — for up to five minutes.
import { readdir, rm } from "node:fs/promises";
import { basename } from "node:path";
import { SveltePlugin } from "bun-plugin-svelte";
import tailwind from "bun-plugin-tailwind";

const manifestPath = "web/static/manifest.json";

interface Entry {
  /** Manifest key, and the filename prefix before the content hash. */
  name: string;
  path: string;
  /** Build output directory, and the only place this entry's files may land. */
  outdir: string;
  plugins: Bun.BunPlugin[];
}

const entries: Entry[] = [
  {
    name: "app",
    path: "web/app/index.ts",
    outdir: "web/static/js/build",
    plugins: [SveltePlugin({ development: false })],
  },
  {
    name: "css",
    path: "web/app/app.css",
    outdir: "web/static/css/build",
    plugins: [tailwind],
  },
];

// The generation this build replaces. It is read *before* anything is written
// and kept alive afterwards on purpose: scripts/deploy.sh builds into the live
// checkout while the previous binary is still serving, and that binary holds
// the previous manifest in memory until systemd restarts it at the end of the
// deploy. Deleting the files it still names would answer every page load in
// that window with a 404 for the bundle. One generation of slack covers it,
// and the build after next reclaims the space.
async function previousManifest(): Promise<Record<string, string>> {
  const file = Bun.file(manifestPath);
  if (!(await file.exists())) {
    return {};
  }

  try {
    return (await file.json()) as Record<string, string>;
  } catch {
    // An unreadable manifest only costs us the pruning hint.
    return {};
  }
}

const previous = await previousManifest();
const manifest: Record<string, string> = {};

for (const entry of entries) {
  const result = await Bun.build({
    entrypoints: [entry.path],
    outdir: entry.outdir,
    target: "browser",
    minify: true,
    naming: `${entry.name}-[hash].[ext]`,
    plugins: entry.plugins,
  });

  if (!result.success) {
    for (const log of result.logs) {
      console.error(log);
    }
    process.exit(1);
  }

  // A JS entry point reports itself as one; a CSS entry has no JS to hang that
  // label on and comes back as a plain asset. Both produce exactly one file
  // here, so falling back to the sole output covers the CSS case without
  // guessing at extensions.
  const output =
    result.outputs.find((o) => o.kind === "entry-point") ??
    (result.outputs.length === 1 ? result.outputs[0] : undefined);
  if (!output) {
    console.error(`${entry.path}: build produced no single output to name`);
    process.exit(1);
  }

  // The manifest holds the public path, not the bare filename: it is the only
  // place that knows which directory an entry built into, and Go should not
  // have to reconstruct it (internal/serve/manifest.go).
  manifest[entry.name] =
    `/${entry.outdir.replace(/^web\//, "")}/${basename(output.path)}`;
}

await Bun.write(manifestPath, JSON.stringify(manifest, null, 2));

// Everything older than the two generations above is unreferenced by any
// running or restartable binary, so it can go. Pruned per output directory,
// since the two entries no longer share one.
const live = new Set([...Object.values(manifest), ...Object.values(previous)]);

for (const outdir of new Set(entries.map((entry) => entry.outdir))) {
  const publicDir = `/${outdir.replace(/^web\//, "")}`;

  for (const name of await readdir(outdir)) {
    if (live.has(`${publicDir}/${name}`)) {
      continue;
    }

    await rm(`${outdir}/${name}`, { recursive: true, force: true });
  }
}

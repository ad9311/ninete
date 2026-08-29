import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defaultClientConditions } from "vite";
import { defineConfig } from "vitest/config";

export default defineConfig({
  // Compiles `.svelte` files for the component tests. Bundling for the browser
  // is bun's job (`web/build.ts`); this plugin exists only so vitest can mount a
  // component, and the two never produce the same artifact. `configFile: false`
  // keeps it that way: a `svelte.config.js` added later would change what the
  // tests compile without changing what ships, since bun-plugin-svelte does not
  // read it.
  plugins: [svelte({ configFile: false })],
  // Svelte 5 ships separate server and client entry points and picks between
  // them by export condition. Without the browser condition vitest resolves the
  // server one, which renders to a string and has no `mount`, so every component
  // test fails on a missing export rather than on anything it asserts.
  //
  // Spread rather than `["browser"]`: `resolve.conditions` replaces Vite's
  // defaults, so naming only "browser" would drop `module` and
  // `development|production` and let a dependency resolve differently here than
  // it does anywhere else. `defaultClientConditions` already contains "browser".
  //
  // Global rather than scoped to the component tests: everything in web/app/
  // ships to a browser (`web/build.ts` sets `target: "browser"`), so client
  // resolution is the honest default for all of it. It is unrelated to which
  // globals exist — that is `environment`, below, and a node-environment test
  // still has no `document` regardless of what conditions resolved.
  resolve: { conditions: [...defaultClientConditions] },
  test: {
    include: ["web/app/**/*.test.ts"],
    // Node, not jsdom, is the default on purpose: §3.9 rule 3 says `lib/` holds
    // no components, and a lib module that reaches for `document` should fail
    // its own test rather than pass because a DOM happened to be present. The
    // component tests opt in with `// @vitest-environment jsdom` at the top of
    // the file.
    environment: "node",
    // Every date bug §3.6 of docs/spa-migration.md describes passes silently
    // under UTC, which is what CI and most machines run, so a zone is pinned
    // here rather than inherited from the machine.
    //
    // Both signs are needed, and the negative one does the heavy lifting: a
    // calendar date is stored at UTC midnight, which in Auckland (UTC+12/+13)
    // is still the same day locally, so a formatter wrongly using local getters
    // reads correctly there and wrong in Los Angeles (UTC-8). Measured against
    // this suite: the west-of-UTC run catches 12 of the cases the east-of-UTC
    // run catches 1 of. Auckland stays the default because it observes DST on
    // the opposite half of the year.
    //
    // TEST_TZ, not TZ: an inherited TZ would silently replace this with
    // whatever zone the developer's machine is in, which is usually close
    // enough to UTC to catch nothing.
    env: { TZ: process.env.TEST_TZ ?? "Pacific/Auckland" },
  },
});

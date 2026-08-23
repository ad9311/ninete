import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    include: ["web/app/**/*.test.ts"],
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

import tsParser from "@typescript-eslint/parser";
import tsPlugin from "@typescript-eslint/eslint-plugin";
import eslintConfigPrettier from "eslint-config-prettier";
import sveltePlugin from "eslint-plugin-svelte";
import svelteParser from "svelte-eslint-parser";

// Asserted rather than assumed: the shape of this export is exactly what went
// wrong before, and a plugin upgrade that moves the rules elsewhere should fail
// the lint run loudly instead of turning every svelte rule off again.
const svelteRecommendedRules = sveltePlugin.configs.recommended.at(-1)?.rules;
if (
  !svelteRecommendedRules ||
  Object.keys(svelteRecommendedRules).length === 0
) {
  throw new Error(
    "eslint-plugin-svelte: no rules found in configs.recommended.at(-1)",
  );
}

export default [
  {
    // Every TS source under web/: the Stimulus entry, the build script, and
    // the Svelte sources in web/app/. Scoping this to web/static/js/ let
    // web/build.ts through unlinted.
    files: ["web/**/*.{ts,tsx}"],
    languageOptions: {
      parser: tsParser,
      ecmaVersion: "latest",
      sourceType: "module",
    },
    plugins: {
      "@typescript-eslint": tsPlugin,
    },
    rules: {
      ...tsPlugin.configs.recommended.rules,
    },
  },
  {
    files: ["web/**/*.svelte"],
    languageOptions: {
      // svelte-eslint-parser handles the markup and hands the contents of
      // <script lang="ts"> to the TS parser. Without the nested parser the
      // type annotations in a script block are syntax errors.
      parser: svelteParser,
      parserOptions: { parser: tsParser },
      ecmaVersion: "latest",
      sourceType: "module",
    },
    // Extracts each script block for the rules below, and is also what makes
    // `<!-- eslint-disable-next-line -->` work inside markup. Without it the
    // svelte rules load but see nothing to report.
    processor: "svelte/svelte",
    plugins: {
      "@typescript-eslint": tsPlugin,
      svelte: sveltePlugin,
    },
    rules: {
      ...tsPlugin.configs.recommended.rules,
      // eslint-plugin-svelte v3 exports flat-config *arrays*, so
      // `configs.recommended` has no `.rules` — spreading it reads as
      // `...undefined`, which is a legal no-op that silently disables all 37
      // rules while leaving the TS ones firing, so nothing looks wrong. Take
      // the rules from the last entry of the array, which is the one holding
      // them (`svelte:recommended:rules`).
      ...svelteRecommendedRules,
    },
  },
  // Last, so it wins: turns off every rule prettier already decides, in both
  // blocks above.
  eslintConfigPrettier,
];

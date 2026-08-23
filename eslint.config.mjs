import tsParser from "@typescript-eslint/parser";
import tsPlugin from "@typescript-eslint/eslint-plugin";
import eslintConfigPrettier from "eslint-config-prettier";
import sveltePlugin from "eslint-plugin-svelte";
import svelteParser from "svelte-eslint-parser";

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
    plugins: {
      "@typescript-eslint": tsPlugin,
      svelte: sveltePlugin,
    },
    rules: {
      ...tsPlugin.configs.recommended.rules,
      ...sveltePlugin.configs.recommended.rules,
    },
  },
  // Last, so it wins: turns off every rule prettier already decides, in both
  // blocks above.
  eslintConfigPrettier,
];

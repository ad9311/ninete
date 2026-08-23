import tsParser from "@typescript-eslint/parser";
import tsPlugin from "@typescript-eslint/eslint-plugin";
import eslintConfigPrettier from "eslint-config-prettier";

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
  eslintConfigPrettier,
];

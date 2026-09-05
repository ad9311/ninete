// The only stylesheet is web/app/app.css, a Tailwind v4 source file. Tailwind's
// own at-rules and the utility names inside @apply are not CSS the standard
// config knows about, so the three rules that would reject them are relaxed
// here rather than silenced inline in a dozen places.
export default {
  extends: ["stylelint-config-standard"],
  rules: {
    // @theme, @custom-variant, @source, @apply, @utility: Tailwind v4's own
    // vocabulary. Listed explicitly so a genuine typo in an at-rule is still
    // caught.
    "at-rule-no-unknown": [
      true,
      {
        ignoreAtRules: [
          "apply",
          "custom-variant",
          "plugin",
          "source",
          "theme",
          "utility",
          "variant",
        ],
      },
    ],
    // `@import "tailwindcss"` is the documented spelling; the standard config
    // wants url().
    "import-notation": null,
    // Tailwind's layers are declared by the framework, not by this file.
    "no-invalid-position-at-import-rule": null,
  },
};

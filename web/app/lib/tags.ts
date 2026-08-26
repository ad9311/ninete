// Splits/joins the semicolon-separated tag input the templates already use
// (see recurrent_expenses/_form.html). Normalization (lowercase, trim,
// dedupe) stays server-side in logic.ParseTagNames — this only has to get the
// same strings there and back, not reimplement the rules.

export function parseTagsInput(raw: string): string[] {
  return raw
    .split(";")
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0);
}

export function joinTagNames(tags: string[]): string {
  return tags.join("; ");
}

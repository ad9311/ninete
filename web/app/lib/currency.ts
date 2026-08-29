// Money helpers matching internal/serve/template_func.go's `currency` and
// web/static/js/controllers/amountController.ts's cents conversion. Amounts
// are stored and transmitted as unsigned cents (an integer), never a float —
// see CLAUDE.md's note on why money is never formatted by hand.

/** Cents (as the API sends them) formatted like the Go `currency` helper: "$1,234.56". */
export function formatCurrency(cents: number): string {
  return (cents / 100).toLocaleString("en-US", {
    style: "currency",
    currency: "USD",
  });
}

/** Cents to the string a `<input type="number" step="0.01">` should show: "1234.56". */
export function centsToInputValue(cents: number): string {
  return (cents / 100).toFixed(2);
}

/**
 * The reverse of centsToInputValue, for what the user actually typed.
 * Returns null for anything that is not a usable amount rather than
 * coercing it to 0, so the caller can tell "empty" apart from "zero".
 *
 * Takes `string | number` because Svelte's `bind:value` on
 * `<input type="number">` binds the live DOM value as a number, not the
 * string the field was seeded with — a caller reading that binding back
 * needs this to accept either.
 */
export function inputValueToCents(value: string | number): number | null {
  const normalized = String(value).replaceAll(",", "").trim();
  if (!normalized) {
    return null;
  }

  const parsed = Number(normalized);
  if (!Number.isFinite(parsed)) {
    return null;
  }

  return Math.round(parsed * 100);
}

// The two paginated lists (expenses, recurrent expenses) share this. Both used
// to carry their own copy of the per-page choices and the window arithmetic,
// which is how a `per_page` the API accepts could reach one list and not the
// other.

/** The `per_page` values the UI offers — the same list the API accepts
 *  (`apiListQuery` in internal/handlers/query.go). The parse helpers below
 *  fall back to the first entry only for *rendering* the footer and links; the
 *  request itself carries the raw value (see `rawParam`), so a hand-edited
 *  `?per_page=37` gets the API's 422 rather than quietly becoming 15. */
export const PER_PAGE_CHOICES = [15, 25, 50, 100];

export const DEFAULT_PER_PAGE = PER_PAGE_CHOICES[0];

/** Reads `per_page` from a query string, falling back to the default for
 *  anything not offered above. */
export function parsePerPage(params: URLSearchParams): number {
  const value = Number(params.get("per_page"));

  return PER_PAGE_CHOICES.includes(value) ? value : DEFAULT_PER_PAGE;
}

/** Reads `page`, clamped to 1 for a missing, zero, negative or unparseable
 *  value. For rendering only, like `parsePerPage`. */
export function parsePage(params: URLSearchParams): number {
  return Math.max(Number(params.get("page") ?? "1") || 1, 1);
}

/**
 * A query-string value exactly as the URL carries it, or undefined when it is
 * absent or empty, so `lib/api.ts` drops it. The lists send `page`,
 * `per_page` and `category_id` through this rather than through the parse
 * helpers: the API validates every query value and answers 422 for one it does
 * not accept, and a client that sanitized first would hide exactly the kind of
 * mismatch that check exists to surface.
 */
export function rawParam(
  params: URLSearchParams,
  key: string,
): string | undefined {
  return params.get(key) || undefined;
}

/**
 * The window of page numbers to render: at most five, centred on the current
 * page and clamped to both ends, so page 1 of 20 shows 1-5 and page 20 shows
 * 16-20 rather than a window running off either edge.
 */
export function pageRange(totalPages: number, currentPage: number): number[] {
  if (totalPages <= 0) return [];

  let start = Math.max(currentPage - 2, 1);
  const end = Math.min(start + 4, totalPages);
  start = Math.max(end - 4, 1);

  const pages: number[] = [];
  for (let page = start; page <= end; page++) pages.push(page);

  return pages;
}

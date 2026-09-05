// The two paginated lists (expenses, recurrent expenses) share this. Both used
// to carry their own copy of the per-page choices and the window arithmetic,
// which is how a `per_page` the API accepts could reach one list and not the
// other.

/** The `per_page` values the UI offers. A query string asking for anything
 *  else falls back to the first entry, so this doubles as the allowlist. */
export const PER_PAGE_CHOICES = [15, 25, 50, 100];

export const DEFAULT_PER_PAGE = PER_PAGE_CHOICES[0];

/** Reads `per_page` from a query string, falling back to the default for
 *  anything not offered above. */
export function parsePerPage(params: URLSearchParams): number {
  const value = Number(params.get("per_page"));

  return PER_PAGE_CHOICES.includes(value) ? value : DEFAULT_PER_PAGE;
}

/** Reads `page`, clamped to 1 for a missing, zero, negative or unparseable
 *  value. */
export function parsePage(params: URLSearchParams): number {
  return Math.max(Number(params.get("page") ?? "1") || 1, 1);
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

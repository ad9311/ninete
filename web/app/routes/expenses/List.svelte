<script lang="ts">
  // Ports expenses/index.html: search panel, category/date-range filters,
  // sortable columns, pagination and tags — the most involved listing in the
  // app. §3.6 of docs/spa-migration.md governs the date-range half: named
  // ranges (the date_range select) resolve to explicit [start, end) bounds
  // client-side via lib/dateRanges.ts, while the explicit search bounds
  // (date_from/date_to) stay YYYY-MM-DD in the URL, because that is what the
  // form fields hold and what a shared link should carry — but they are
  // resolved to epoch bounds here before the fetch, exactly as named ranges
  // are. They target created_at, an *instant*, so the day they name only
  // becomes a window once the viewer's zone is applied, and the browser is the
  // only party that knows it. The billed date is a month now, so a
  // day-precision bound on it meant nothing; date_range filters that.
  import { untrack } from "svelte";
  import { AlignLeft, CalendarRange, ChevronDown, Search, Tag } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import LocalDate from "../../components/LocalDate.svelte";
  import PaginationFooter from "../../components/PaginationFooter.svelte";
  import SortHeader from "../../components/SortHeader.svelte";
  import { APIRequestError, get } from "../../lib/api";
  import { type Category, fetchCategories } from "../../lib/categories";
  import { formatCurrency } from "../../lib/currency";
  import { computeDateRange, DATE_RANGE_OPTIONS } from "../../lib/dateRanges";
  import { localDayEnd, localDayStart } from "../../lib/dates";
  import { parsePage, parsePerPage } from "../../lib/pagination";
  import { BASE_PATH, navigate } from "../../router";
  import type { Expense, ExpenseListResponse, Pagination } from "./types";

  interface Props {
    search?: string;
  }

  let { search = "" }: Props = $props();

  let categories = $state<Category[]>([]);
  let rows = $state<Expense[]>([]);
  let pagination = $state<Pagination | null>(null);
  let error = $state("");

  const params = $derived(new URLSearchParams(search));
  const categoryId = $derived(Number(params.get("category_id") ?? "0"));
  const sortField = $derived(params.get("sort_field") ?? "created_at");
  const sortOrder = $derived(params.get("sort_order") ?? "DESC");
  const page = $derived(parsePage(params));
  const perPage = $derived(parsePerPage(params));

  const query = $derived(params.get("q") ?? "");
  const tag = $derived(params.get("tag") ?? "");
  const dateFrom = $derived(params.get("date_from") ?? "");
  const dateTo = $derived(params.get("date_to") ?? "");
  const hasDateBounds = $derived(dateFrom !== "" || dateTo !== "");
  // Resolved here rather than in the fetch effect so a lopsided or malformed
  // pair in a hand-edited URL surfaces as one error message instead of a failed
  // request.
  //
  // Both bounds or neither. Completing the missing side from the one that was
  // given is what this used to do, and it turned a From-only search into a
  // one-day search while the panel still showed Range — the listing came back
  // near-empty with nothing on screen saying why. Day mode is how a one-day
  // search is asked for.
  const createdBounds = $derived.by(() => {
    if (!hasDateBounds) return null;
    if (dateFrom === "" || dateTo === "") return "half" as const;

    let start: number;
    let end: number;
    try {
      start = localDayStart(dateFrom);
      end = localDayEnd(dateTo);
    } catch {
      return "invalid" as const;
    }

    // Caught here rather than left to the API, whose message names start and
    // end — the epoch bounds computed just above, which are not the From and To
    // fields the user filled in.
    if (start > end) return "inverted" as const;

    return { start, end };
  });
  const CREATED_BOUNDS_ERRORS: Record<"half" | "invalid" | "inverted", string> =
    {
      half: "Fill in both dates, or switch to Day to search one day.",
      invalid: "Dates must use the YYYY-MM-DD format.",
      inverted: "The From date must be on or before the To date.",
    };
  const hasTextSearch = $derived(query !== "" || tag !== "");
  const searchActive = $derived(hasDateBounds || hasTextSearch);
  // Only the explicit bounds clear the preset range. They filter created_at
  // while the preset filters the billed date, so the two cannot combine into a
  // meaningful window and the API drops the preset whenever they are present
  // (expenseSearch.apply, expense_search.go) — the select has to say so.
  //
  // A text search does not: it keeps whatever range is selected, this_month by
  // default. The API's matching implicit widening never fires from here,
  // because this chain always resolves the named range to bounds itself and
  // that is what sets explicitRange server-side (GetAPIExpenses).
  const clearsPresetRange = $derived(hasDateBounds);
  const dateRangeValue = $derived(
    clearsPresetRange ? "all_time" : params.get("date_range") || "this_month",
  );
  const totalAmount = $derived(rows.reduce((sum, row) => sum + row.amount, 0));

  // The search inputs are locally editable and only resynced from the URL
  // when it actually changes (a real navigation) — never on every keystroke.
  let searchInput = $state("");
  let tagInput = $state("");
  let dateFromInput = $state("");
  let dateToInput = $state("");
  // The mode is not its own query parameter: a one-day search is exactly
  // date_from === date_to, so the URL already says it. Deriving it keeps the
  // two dates the only source of truth, and a shared link reopens in the mode
  // it was searched in.
  //
  // Day is the opening mode, which is the initial value below rather than
  // anything derived: an unbounded URL says nothing about the mode, and the
  // resync must leave a deliberate Range choice alone. It re-runs on every
  // navigation — including one that only changed `q` or `tag` — so rewriting
  // the mode unconditionally would snap the panel back to Day the moment a
  // text search was submitted from Range, taking the To field with it.
  //
  // This decides the *mode* alone: with both fields empty no bound is sent
  // either way, and a half-filled pair is still refused rather than quietly
  // narrowed (see createdBounds above).
  let singleDay = $state(true);
  type DateMode = "range" | "day";
  const dateMode = $derived<DateMode>(singleDay ? "day" : "range");

  $effect(() => {
    searchInput = query;
    tagInput = tag;
    dateFromInput = dateFrom;
    dateToInput = dateTo;
    if (hasDateBounds) singleDay = dateFrom === dateTo;
  });

  // The To field is unmounted in Day mode but its value keeps following From,
  // so switching back to Range shows the day just searched rather than a stale
  // bound — the useful starting point for widening it.
  $effect(() => {
    if (singleDay) dateToInput = dateFromInput;
  });

  // Switching to Day with only the To field filled would otherwise mirror an
  // empty From over it and throw the date away. Fold it back into From first;
  // the effect above then keeps the two in step.
  function selectDateMode(mode: DateMode): void {
    if (
      mode === "day" &&
      dateFromInput.trim() === "" &&
      dateToInput.trim() !== ""
    ) {
      dateFromInput = dateToInput;
    }
    singleDay = mode === "day";
  }

  const SEARCH_PANEL_KEY = "search-panel-open";

  function panelStorageKey(): string {
    return `${SEARCH_PANEL_KEY}:${window.location.pathname}`;
  }

  function readInitialPanelOpen(): boolean {
    try {
      const stored = sessionStorage.getItem(panelStorageKey());
      if (stored !== null) return stored === "true";
    } catch {
      // Storage can be unavailable; fall through to the URL-derived default.
    }

    return searchActive;
  }

  let panelOpen = $state(untrack(readInitialPanelOpen));

  $effect(() => {
    try {
      sessionStorage.setItem(panelStorageKey(), String(panelOpen));
    } catch {
      // Best-effort, same as the recurrent expense form's localStorage use.
    }
  });

  $effect(() => {
    let cancelled = false;

    fetchCategories()
      .then((result) => {
        if (!cancelled) categories = result;
      })
      .catch(() => {
        // The filter select is just empty until a retry.
      });

    return () => {
      cancelled = true;
    };
  });

  $effect(() => {
    let cancelled = false;
    // The preset range's bounds, on the billed date. The search's own bounds
    // (createdBounds) filter created_at and are a separate pair for that
    // reason; the API drops this one whenever they are present.
    const rangeBounds = computeDateRange(dateRangeValue);

    if (typeof createdBounds === "string") {
      rows = [];
      pagination = null;
      error = CREATED_BOUNDS_ERRORS[createdBounds];

      return;
    }

    get<ExpenseListResponse>("/expenses", {
      params: {
        q: query || undefined,
        tag: tag || undefined,
        created_start: createdBounds?.start,
        created_end: createdBounds?.end,
        category_id: categoryId > 0 ? categoryId : undefined,
        sort_field: sortField,
        sort_order: sortOrder,
        page,
        per_page: perPage,
        start: rangeBounds?.start,
        end: rangeBounds?.end,
      },
    })
      .then((result) => {
        if (cancelled) return;
        rows = result.data;
        pagination = result.pagination;
        error = "";
      })
      .catch((err) => {
        if (cancelled) return;
        error =
          err instanceof APIRequestError
            ? err.message
            : "Something went wrong.";
      });

    return () => {
      cancelled = true;
    };
  });

  function buildHref(
    overrides: Record<string, string | number | undefined>,
  ): string {
    const next = new URLSearchParams(search);
    for (const [key, value] of Object.entries(overrides)) {
      if (value === undefined || value === "") {
        next.delete(key);
      } else {
        next.set(key, String(value));
      }
    }

    return `${BASE_PATH}/expenses?${next.toString()}`;
  }

  function sortHref(field: string): string {
    const order = sortField === field && sortOrder === "ASC" ? "DESC" : "ASC";

    return buildHref({ sort_field: field, sort_order: order, page: 1 });
  }

  function onCategoryChange(event: Event): void {
    const value = (event.currentTarget as HTMLSelectElement).value;
    navigate(
      buildHref({ category_id: value ? Number(value) : undefined, page: 1 }),
    );
  }

  function onDateRangeChange(event: Event): void {
    const value = (event.currentTarget as HTMLSelectElement).value;
    navigate(buildHref({ date_range: value, page: 1 }));
  }

  function onPerPageChange(value: number): void {
    navigate(buildHref({ per_page: value, page: 1 }));
  }

  function submitSearch(event: SubmitEvent): void {
    event.preventDefault();
    navigate(
      buildHref({
        q: searchInput.trim() || undefined,
        tag: tagInput.trim() || undefined,
        date_from: dateFromInput.trim() || undefined,
        // Read from From rather than the mirrored input: the effect that keeps
        // them equal has not necessarily flushed when this runs.
        date_to:
          (singleDay ? dateFromInput.trim() : dateToInput.trim()) || undefined,
        page: 1,
      }),
    );
  }

  function clearSearch(): void {
    const overrides: Record<string, string | number | undefined> = {
      q: undefined,
      tag: undefined,
      date_from: undefined,
      date_to: undefined,
      page: 1,
    };
    // Explicit bounds force the range select to all_time; dropping it here too
    // keeps Clear from leaving an unbounded, unfiltered listing behind. The
    // URL's own value, not dateRangeValue: that derived reads "all_time"
    // whenever bounds are set, so it would also throw away a range the user
    // picked by hand (six_months + explicit bounds would snap to this_month).
    if (hasDateBounds && params.get("date_range") === "all_time") {
      overrides.date_range = undefined;
    }

    navigate(buildHref(overrides));
  }

  // The search row's fields share a shape; the two strings below narrow it.
  // Written once rather than repeated across four labels, which is where the
  // widths drifted apart before.
  //
  // The shared string carries no `flex`/`flex-basis` of its own: the date
  // labels append `dateFieldClass` to it, and Tailwind resolves two utilities
  // setting the same property by their order in the *generated stylesheet*,
  // not by their order in the attribute. A `basis-48` left in here would
  // therefore beat the date fields' `basis-40` rather than losing to it.
  const searchFieldClass = "text-muted inline-flex min-w-0 items-center gap-2";
  // The description and tag fields, which grow into the spare width.
  const textFieldClass = "flex-1 basis-48 max-md:basis-auto";
  // Fixed rather than growing: if the date inputs absorbed the spare width
  // there would be no free space for justify-end to push the cluster right. On
  // a narrow screen they share the row instead. `shrink` is spelled out
  // because `flex-none` sets the whole shorthand, flex-shrink: 0 included.
  //
  // The basis is wider than a bare text field would need because these are
  // native date inputs: the picker icon and the browser's own YYYY-MM-DD
  // rendering have a larger intrinsic minimum, and at 10rem they sat cramped.
  // The extra width comes out of the cluster's leading whitespace, which is
  // what narrows the gap to the tag field. Desktop only — `max-md:basis-0`
  // already hands the fields the full row once it stacks.
  const dateFieldClass =
    "flex-none shrink basis-48 max-md:flex-1 max-md:basis-0 max-md:gap-1";

  // The two bounds modes, as a segmented control rather than a checkbox: the
  // choice changes what the fields below *mean*, and naming both halves says
  // so where a lone "Single day" tickbox left the unticked state unnamed. Day
  // then renders one field instead of a live one beside a disabled mirror.
  const DATE_MODES: [DateMode, string][] = [
    ["range", "Range"],
    ["day", "Day"],
  ];
  // The visible half of each segment. The radio itself is `sr-only`, so the
  // span carries both the look and the selected state — and it has to be the
  // span rather than the label, because `peer-*` matches a *sibling* of the
  // input and the label is its parent.
  const segmentClass =
    "flex cursor-pointer items-center justify-center px-3 py-1 text-sm " +
    "text-muted select-none peer-checked:bg-primary peer-checked:text-on-primary " +
    // Inset: an outline paints outside the border box, the span is flush
    // against the fieldset's padding box on every edge, and the fieldset
    // clips with `overflow-hidden` — so a normal ring is painted straight
    // into the clipped region and never appears.
    "peer-focus-visible:outline-2 peer-focus-visible:-outline-offset-2 " +
    "peer-focus-visible:outline-primary " +
    "hover:text-primary peer-checked:hover:text-on-primary";

  const sortableColumns: [string, string][] = [
    ["category_id", "Category"],
    ["description", "Description"],
    ["amount", "Amount"],
    ["date", "Billed month"],
    ["created_at", "Created"],
  ];
</script>

<details class="group mb-3" bind:open={panelOpen}>
  <summary
    class="inline-flex w-fit cursor-pointer items-center gap-2 py-1 text-sm text-muted hover:text-primary"
  >
    <Icon icon={Search} class="h-4 w-4 shrink-0" />
    <span>Search</span>
    <Icon
      icon={ChevronDown}
      class="h-4 w-4 transition-transform group-open:rotate-180"
    />
  </summary>
  <form
    class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-2 max-md:flex-col max-md:items-stretch"
    role="search"
    aria-label="Search expenses"
    onsubmit={submitSearch}
  >
    <label class="{searchFieldClass} {textFieldClass}">
      <span class="sr-only">Description</span>
      <Icon icon={AlignLeft} class="h-4 w-4 shrink-0" />
      <input
        type="search"
        class="min-w-0"
        bind:value={searchInput}
        placeholder="Description"
        maxlength="50"
      />
    </label>
    <label class="{searchFieldClass} {textFieldClass}">
      <span class="sr-only">Tag</span>
      <Icon icon={Tag} class="h-4 w-4 shrink-0" />
      <input
        type="search"
        class="min-w-0"
        bind:value={tagInput}
        placeholder="Tag"
        maxlength="50"
      />
    </label>
    <!-- Grows to take the leftover width but packs its contents to the right,
      so the free space collects between the tag input and the date cluster.
      That, plus a tighter internal gap than the row's, keeps the mode and the
      bounds it governs reading as one group. -->
    <div
      class="flex min-w-0 flex-1 basis-120 flex-wrap items-center justify-end gap-2 max-md:basis-auto"
    >
      <!-- The mode comes before the fields it governs: it decides how many
        there are, so reading it first is the order the panel is used in. -->
      <!-- The wrapper is what takes the full width once the panel stacks, so
        the pill gets a line of its own and sits at the right edge while
        keeping its natural width. Stretching the pill itself instead left two
        very wide segments. On desktop it is an inert flex item around a
        flex-none child, so the row is unchanged. -->
      <div class="flex flex-none items-center max-md:w-full max-md:justify-end">
        <fieldset
          class="flex items-center overflow-hidden rounded-xs border border-line"
        >
          <legend class="sr-only">Created date bounds mode</legend>
          {#each DATE_MODES as [value, label], index (value)}
            <label class="inline-flex">
              <input
                type="radio"
                class="peer sr-only"
                name="date-mode"
                {value}
                checked={dateMode === value}
                onchange={() => selectDateMode(value)}
              />
              <!-- The span, not the label, carries the look: `peer-*` matches
                a sibling of the input, and the label is its parent. -->
              <span
                class="{segmentClass} {index === 0
                  ? 'border-r border-line'
                  : ''}"
              >
                {label}
              </span>
            </label>
          {/each}
        </fieldset>
      </div>
      <label class="{searchFieldClass} {dateFieldClass}">
        <span class="sr-only"
          >{singleDay ? "Created day" : "Created from date"}</span
        >
        <span class="text-sm" aria-hidden="true"
          >{singleDay ? "On" : "From"}</span
        >
        <!-- A date input's value is already YYYY-MM-DD, which is exactly what
          the URL carries and what localDayStart/localDayEnd parse, so the
          browser's own picker and validation replace the pattern and length
          checks this field used to spell out by hand. -->
        <input type="date" class="min-w-0" bind:value={dateFromInput} />
      </label>
      {#if !singleDay}
        <label class="{searchFieldClass} {dateFieldClass}">
          <span class="sr-only">Created to date</span>
          <span class="text-sm" aria-hidden="true">To</span>
          <input type="date" class="min-w-0" bind:value={dateToInput} />
        </label>
      {/if}
    </div>
    <!-- Its own line on desktop: `md:w-full` makes the wrapping row break
      before it, so the Clear button appearing beside Search can never push the
      fields above out of their single line. The stacked layout already gives
      every child the full width, so the pair is left as it was there. -->
    <div
      class="ml-auto flex gap-2 max-md:mt-3 max-md:ml-0 md:w-full md:justify-end"
    >
      <button type="submit" class="btn btn-primary min-w-20 max-md:flex-1">
        Search
      </button>
      {#if searchActive}
        <button
          type="button"
          class="btn btn-neutral min-w-20 max-md:flex-1"
          onclick={clearSearch}
        >
          Clear
        </button>
      {/if}
    </div>
  </form>
</details>

<div class="mb-3 flex flex-wrap justify-end gap-3">
  <label class="inline-flex items-center gap-2 text-muted">
    <span class="sr-only">Category</span>
    <Icon icon={Tag} class="h-4 w-4 shrink-0" />
    <select class="w-56" value={categoryId || ""} onchange={onCategoryChange}>
      <option value="">All categories</option>
      {#each categories as category (category.id)}
        <option value={category.id}>{category.name}</option>
      {/each}
    </select>
  </label>
  <label class="inline-flex items-center gap-2 text-muted">
    <span class="sr-only">Date range</span>
    <Icon icon={CalendarRange} class="h-4 w-4 shrink-0" />
    <select class="w-56" value={dateRangeValue} onchange={onDateRangeChange}>
      <option value="all_time">All time</option>
      {#each DATE_RANGE_OPTIONS as option (option.value)}
        <option value={option.value}>{option.label}</option>
      {/each}
    </select>
  </label>
</div>

<!-- Outside the search panel, which is collapsible: a load failure rendered
  inside it would be invisible whenever the panel is closed, leaving an empty
  table and no explanation. This carries the bounds-validation messages too —
  they are about the fields above, but they also mean the listing did not
  load. -->
{#if error}
  <p class="text-danger">{error}</p>
{/if}

<div class="overflow-x-auto">
  <table class="data-table">
    <thead>
      <tr>
        {#each sortableColumns as [field, label] (field)}
          <SortHeader
            {label}
            href={sortHref(field)}
            active={sortField === field}
            order={sortOrder}
          />
        {/each}
        <th>Tags</th>
      </tr>
    </thead>
    <tbody>
      {#each rows as row (row.id)}
        <tr>
          <td>{row.category_name}</td>
          <td>
            <a href={`${BASE_PATH}/expenses/${row.id}`}>{row.description}</a>
          </td>
          <td class="amount font-semibold text-fg">
            {formatCurrency(row.amount)}
          </td>
          <td><LocalDate value={row.date} month /></td>
          <td><LocalDate value={row.created_at} datetime /></td>
          <td>
            {#if row.tags.length > 0}
              <div class="flex flex-wrap gap-2">
                {#each row.tags as t (t)}
                  <span class="chip chip-tag">{t}</span>
                {/each}
              </div>
            {:else}
              <span class="chip">No tags</span>
            {/if}
          </td>
        </tr>
      {/each}
    </tbody>
    <tfoot>
      <tr>
        <th colspan="6">
          Total expenses
          <span class="amount font-semibold text-fg"
            >{formatCurrency(totalAmount)}</span
          >
        </th>
      </tr>
    </tfoot>
  </table>
</div>

<PaginationFooter
  {pagination}
  {page}
  {perPage}
  hrefFor={(target) => buildHref({ page: target })}
  {onPerPageChange}
/>

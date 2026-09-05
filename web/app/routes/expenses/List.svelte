<script lang="ts">
  // Ports expenses/index.html: search panel, category/date-range filters,
  // sortable columns, pagination and tags — the most involved listing in the
  // app. §3.6 of docs/spa-migration.md governs the date-range half: named
  // ranges (the date_range select) resolve to explicit [start, end) bounds
  // client-side via lib/dateRanges.ts, while the explicit search bounds
  // (date_from/date_to) are plain YYYY-MM-DD strings the API parses itself,
  // unchanged from the template path.
  import { untrack } from "svelte";
  import { AlignLeft, CalendarRange, ChevronDown, Search, Tag } from "lucide";
  import DateHelp from "../../components/DateHelp.svelte";
  import Icon from "../../components/Icon.svelte";
  import LocalDate from "../../components/LocalDate.svelte";
  import PaginationFooter from "../../components/PaginationFooter.svelte";
  import SortHeader from "../../components/SortHeader.svelte";
  import { APIRequestError, get } from "../../lib/api";
  import { type Category, fetchCategories } from "../../lib/categories";
  import { formatCurrency } from "../../lib/currency";
  import { computeDateRange, DATE_RANGE_OPTIONS } from "../../lib/dateRanges";
  import { parsePage, parsePerPage } from "../../lib/pagination";
  import { BASE_PATH, navigate } from "../../router";
  import type { Expense, ExpenseListResponse, Pagination } from "./types";

  interface Props {
    search?: string;
  }

  let { search = "" }: Props = $props();

  const CREATED_DATE_FIELD = "created_at";

  let categories = $state<Category[]>([]);
  let rows = $state<Expense[]>([]);
  let pagination = $state<Pagination | null>(null);
  let error = $state("");

  const params = $derived(new URLSearchParams(search));
  const categoryId = $derived(Number(params.get("category_id") ?? "0"));
  const sortField = $derived(params.get("sort_field") ?? "date");
  const sortOrder = $derived(params.get("sort_order") ?? "DESC");
  const page = $derived(parsePage(params));
  const perPage = $derived(parsePerPage(params));

  const query = $derived(params.get("q") ?? "");
  const tag = $derived(params.get("tag") ?? "");
  const dateFrom = $derived(params.get("date_from") ?? "");
  const dateTo = $derived(params.get("date_to") ?? "");
  const dateField = $derived(
    params.get("date_field") === CREATED_DATE_FIELD
      ? CREATED_DATE_FIELD
      : "date",
  );
  const hasDateBounds = $derived(dateFrom !== "" || dateTo !== "");
  const hasTextSearch = $derived(query !== "" || tag !== "");
  const searchActive = $derived(hasDateBounds || hasTextSearch);
  const explicitRange = $derived(params.has("date_range"));
  // A text search with no explicit date_range widens to all time so matches
  // older than the default range are not silently hidden — mirrors
  // expenseSearch.clearsPresetRange() (expense_search.go).
  const clearsPresetRange = $derived(
    hasDateBounds || (hasTextSearch && !explicitRange),
  );
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
  let dateFieldChecked = $state(false);

  $effect(() => {
    searchInput = query;
    tagInput = tag;
    dateFromInput = dateFrom;
    dateToInput = dateTo;
    dateFieldChecked = dateField === CREATED_DATE_FIELD;
  });

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
    const bounds = computeDateRange(dateRangeValue);

    get<ExpenseListResponse>("/expenses", {
      params: {
        q: query || undefined,
        tag: tag || undefined,
        date_from: dateFrom || undefined,
        date_to: dateTo || undefined,
        date_field:
          dateField === CREATED_DATE_FIELD ? CREATED_DATE_FIELD : undefined,
        category_id: categoryId > 0 ? categoryId : undefined,
        sort_field: sortField,
        sort_order: sortOrder,
        page,
        per_page: perPage,
        start: bounds?.start,
        end: bounds?.end,
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
        date_to: dateToInput.trim() || undefined,
        date_field: dateFieldChecked ? CREATED_DATE_FIELD : undefined,
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
      date_field: undefined,
      page: 1,
    };
    // An active search forces the range select to all_time; dropping it here
    // too keeps Clear from leaving an unbounded, unfiltered listing behind.
    // The URL's own value, not dateRangeValue: that derived reads "all_time"
    // for *any* active search, so it would also throw away a range the user
    // picked by hand (six_months + explicit bounds would snap to this_month).
    if (searchActive && params.get("date_range") === "all_time") {
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
  const dateFieldClass =
    "flex-none shrink basis-40 max-md:flex-1 max-md:basis-0 max-md:gap-1";

  const sortableColumns: [string, string][] = [
    ["category_id", "Category"],
    ["description", "Description"],
    ["amount", "Amount"],
    ["date", "Billed"],
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
      so the free space collects between the tag input and the toggle. That,
      plus a tighter internal gap than the row's, makes the toggle read as part
      of the date cluster rather than as a trailer on the field before it. -->
    <div
      class="flex min-w-0 flex-1 basis-[30rem] flex-wrap items-center justify-end gap-2 max-md:basis-auto"
    >
      <!-- The label's children are flat on purpose: `peer-checked:` reaches a
        following *sibling*, so the two words cannot be nested in a wrapper. -->
      <label
        class="inline-flex flex-none cursor-pointer items-center gap-2 text-muted select-none max-md:mt-3 max-md:grow max-md:basis-full"
        title="Apply the date bounds to the billed date or the created date"
      >
        <span class="sr-only">
          Apply date bounds to the created date instead of the billed date
        </span>
        <input
          type="checkbox"
          class="peer absolute h-px w-px opacity-0"
          bind:checked={dateFieldChecked}
        />
        <span class="toggle-switch" aria-hidden="true"></span>
        <span class="min-w-16 text-sm peer-checked:hidden" aria-hidden="true">
          Billed
        </span>
        <span
          class="hidden min-w-16 text-sm peer-checked:inline"
          aria-hidden="true"
        >
          Created
        </span>
      </label>
      <label class="{searchFieldClass} {dateFieldClass}">
        <span class="sr-only">From date</span>
        <span class="text-sm" aria-hidden="true">From</span>
        <!-- The regex has to be an expression, not a quoted attribute: Svelte
          reads {4} inside a plain attribute value as an interpolation and the
          template's `\d{4}-\d{2}-\d{2}` would ship as `\d4-\d2-\d2`, which no
          real date matches, so the field could never pass validation. -->
        <input
          type="text"
          class="min-w-0"
          bind:value={dateFromInput}
          placeholder="YYYY-MM-DD"
          inputmode="numeric"
          pattern={"\\d{4}-\\d{2}-\\d{2}"}
          title="Use the YYYY-MM-DD format, e.g. 2026-07-12"
          maxlength="10"
        />
      </label>
      <label class="{searchFieldClass} {dateFieldClass}">
        <span class="sr-only">To date</span>
        <span class="text-sm" aria-hidden="true">To</span>
        <!-- Expression form, same reason as the From field above. -->
        <input
          type="text"
          class="min-w-0"
          bind:value={dateToInput}
          placeholder="YYYY-MM-DD"
          inputmode="numeric"
          pattern={"\\d{4}-\\d{2}-\\d{2}"}
          title="Use the YYYY-MM-DD format, e.g. 2026-07-12"
          maxlength="10"
        />
      </label>
      <DateHelp
        label="Show accepted date format"
        title="Dates must be:"
        panelClass="left-auto right-0 max-md:right-auto max-md:left-0"
      >
        <ul>
          <li><code>YYYY-MM-DD</code> (e.g. <code>2026-07-12</code>)</li>
          <li>Both bounds are inclusive</li>
          <li>Leave empty to use the date range filter</li>
          <li>Bounds apply to the billed or created date, per the selector</li>
        </ul>
      </DateHelp>
    </div>
    {#if error}
      <p class="text-danger">{error}</p>
    {/if}
    <div class="ml-auto flex gap-2 max-md:mt-3 max-md:ml-0">
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
        <th>Actions</th>
      </tr>
    </thead>
    <tbody>
      {#each rows as row (row.id)}
        <tr>
          <td>{row.category_name}</td>
          <td>{row.description}</td>
          <td class="font-semibold text-fg">{formatCurrency(row.amount)}</td>
          <td><LocalDate value={row.date} /></td>
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
          <td><a href={`${BASE_PATH}/expenses/${row.id}`}>Visit</a></td>
        </tr>
      {/each}
    </tbody>
    <tfoot>
      <tr>
        <th colspan="7">
          Total expenses
          <span class="font-semibold text-fg"
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

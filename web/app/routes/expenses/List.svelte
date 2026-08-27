<script lang="ts">
  // Ports expenses/index.html: search panel, category/date-range filters,
  // sortable columns, pagination and tags — the most involved listing in the
  // app. §3.6 of docs/spa-migration.md governs the date-range half: named
  // ranges (the date_range select) resolve to explicit [start, end) bounds
  // client-side via lib/dateRanges.ts, while the explicit search bounds
  // (date_from/date_to) are plain YYYY-MM-DD strings the API parses itself,
  // unchanged from the template path.
  import { untrack } from "svelte";
  import {
    AlignLeft,
    CalendarRange,
    ChevronDown,
    Rows3,
    Search,
    Tag,
  } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import DateHelp from "../../components/DateHelp.svelte";
  import { APIRequestError, get } from "../../lib/api";
  import { type Category, fetchCategories } from "../../lib/categories";
  import { formatCurrency } from "../../lib/currency";
  import { computeDateRange, DATE_RANGE_OPTIONS } from "../../lib/dateRanges";
  import LocalDate from "../../components/LocalDate.svelte";
  import { BASE_PATH, navigate } from "../../router";
  import type { Expense, ExpenseListResponse, Pagination } from "./types";

  interface Props {
    search?: string;
  }

  let { search = "" }: Props = $props();

  const PER_PAGE_CHOICES = [15, 25, 50, 100];
  const CREATED_DATE_FIELD = "created_at";

  let categories = $state<Category[]>([]);
  let rows = $state<Expense[]>([]);
  let pagination = $state<Pagination | null>(null);
  let error = $state("");

  const params = $derived(new URLSearchParams(search));
  const categoryId = $derived(Number(params.get("category_id") ?? "0"));
  const sortField = $derived(params.get("sort_field") ?? "date");
  const sortOrder = $derived(params.get("sort_order") ?? "DESC");
  const page = $derived(Math.max(Number(params.get("page") ?? "1") || 1, 1));
  const perPage = $derived(
    PER_PAGE_CHOICES.includes(Number(params.get("per_page")))
      ? Number(params.get("per_page"))
      : 15,
  );

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

  function pageRange(totalPages: number, currentPage: number): number[] {
    if (totalPages <= 0) return [];

    let start = Math.max(currentPage - 2, 1);
    const end = Math.min(start + 4, totalPages);
    start = Math.max(end - 4, 1);

    const pages: number[] = [];
    for (let i = start; i <= end; i++) pages.push(i);

    return pages;
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

  function onPerPageChange(event: Event): void {
    const value = Number((event.currentTarget as HTMLSelectElement).value);
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
    if (searchActive && dateRangeValue === "all_time") {
      overrides.date_range = undefined;
    }

    navigate(buildHref(overrides));
  }

  const sortableColumns: [string, string][] = [
    ["category_id", "Category"],
    ["description", "Description"],
    ["amount", "Amount"],
    ["date", "Billed"],
    ["created_at", "Created"],
  ];
</script>

<details class="search-panel" bind:open={panelOpen}>
  <summary class="search-summary">
    <Icon icon={Search} class="filter-icon" />
    <span>Search</span>
    <Icon icon={ChevronDown} class="search-caret" />
  </summary>
  <form
    class="search-bar"
    role="search"
    aria-label="Search expenses"
    onsubmit={submitSearch}
  >
    <label class="search-field">
      <span class="sr-only">Description</span>
      <Icon icon={AlignLeft} class="filter-icon" />
      <input
        type="search"
        bind:value={searchInput}
        placeholder="Description"
        maxlength="50"
      />
    </label>
    <label class="search-field">
      <span class="sr-only">Tag</span>
      <Icon icon={Tag} class="filter-icon" />
      <input
        type="search"
        bind:value={tagInput}
        placeholder="Tag"
        maxlength="50"
      />
    </label>
    <div class="search-date-group">
      <label
        class="date-toggle"
        title="Apply the date bounds to the billed date or the created date"
      >
        <span class="sr-only"
          >Apply date bounds to the created date instead of the billed date</span
        >
        <input type="checkbox" bind:checked={dateFieldChecked} />
        <span class="switch" aria-hidden="true"></span>
        <span class="date-toggle-text" aria-hidden="true">
          <span class="when-off">Billed</span>
          <span class="when-on">Created</span>
        </span>
      </label>
      <label class="search-field search-field-date">
        <span class="sr-only">From date</span>
        <span class="search-date-label" aria-hidden="true">From</span>
        <input
          type="text"
          bind:value={dateFromInput}
          placeholder="YYYY-MM-DD"
          inputmode="numeric"
          pattern="\d{4}-\d{2}-\d{2}"
          title="Use the YYYY-MM-DD format, e.g. 2026-07-12"
          maxlength="10"
        />
      </label>
      <label class="search-field search-field-date">
        <span class="sr-only">To date</span>
        <span class="search-date-label" aria-hidden="true">To</span>
        <input
          type="text"
          bind:value={dateToInput}
          placeholder="YYYY-MM-DD"
          inputmode="numeric"
          pattern="\d{4}-\d{2}-\d{2}"
          title="Use the YYYY-MM-DD format, e.g. 2026-07-12"
          maxlength="10"
        />
      </label>
      <DateHelp
        label="Show accepted date format"
        panelClass="date-help-panel-end"
      >
        <p class="date-help-title">Dates must be:</p>
        <ul>
          <li><code>YYYY-MM-DD</code> (e.g. <code>2026-07-12</code>)</li>
          <li>Both bounds are inclusive</li>
          <li>Leave empty to use the date range filter</li>
          <li>Bounds apply to the billed or created date, per the selector</li>
        </ul>
      </DateHelp>
    </div>
    {#if error}
      <p class="form-error-text">{error}</p>
    {/if}
    <div class="search-actions">
      <button type="submit" class="btn-primary search-button">Search</button>
      {#if searchActive}
        <button
          type="button"
          class="btn-neutral search-button"
          onclick={clearSearch}
        >
          Clear
        </button>
      {/if}
    </div>
  </form>
</details>

<div class="filters">
  <label>
    <span class="sr-only">Category</span>
    <Icon icon={Tag} class="filter-icon" />
    <select value={categoryId || ""} onchange={onCategoryChange}>
      <option value="">All categories</option>
      {#each categories as category (category.id)}
        <option value={category.id}>{category.name}</option>
      {/each}
    </select>
  </label>
  <label>
    <span class="sr-only">Date range</span>
    <Icon icon={CalendarRange} class="filter-icon" />
    <select value={dateRangeValue} onchange={onDateRangeChange}>
      <option value="all_time">All time</option>
      {#each DATE_RANGE_OPTIONS as option (option.value)}
        <option value={option.value}>{option.label}</option>
      {/each}
    </select>
  </label>
</div>

<div class="table-scroll">
  <table class="data-table">
    <thead>
      <tr>
        {#each sortableColumns as [field, label] (field)}
          <th>
            <a href={sortHref(field)} class="sort-link">
              {label}
              {#if sortField === field}
                <span class="sort-indicator"
                  >{sortOrder === "ASC" ? "▲" : "▼"}</span
                >
              {/if}
            </a>
          </th>
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
          <td class="amount-value">{formatCurrency(row.amount)}</td>
          <td><LocalDate value={row.date} /></td>
          <td><LocalDate value={row.created_at} datetime /></td>
          <td>
            {#if row.tags.length > 0}
              <div class="chip-list">
                {#each row.tags as t (t)}
                  <span class="chip chip-tag">{t}</span>
                {/each}
              </div>
            {:else}
              <span class="chip chip-empty">No tags</span>
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
          <span class="amount-value">{formatCurrency(totalAmount)}</span>
        </th>
      </tr>
    </tfoot>
  </table>
</div>

<div class="pagination-footer">
  <label class="per-page">
    <span class="sr-only">Rows per page</span>
    <Icon icon={Rows3} class="filter-icon" />
    <select value={perPage} onchange={onPerPageChange}>
      {#each PER_PAGE_CHOICES as choice (choice)}
        <option value={choice}>{choice} per page</option>
      {/each}
    </select>
  </label>

  {#if pagination && pagination.total_pages > 1}
    <nav class="pagination" aria-label="Pagination">
      {#if pagination.has_prev}
        <a href={buildHref({ page: page - 1 })} class="pagination-link">Prev</a>
      {:else}
        <span class="pagination-link pagination-disabled">Prev</span>
      {/if}

      {#each pageRange(pagination.total_pages, page) as p (p)}
        {#if p === page}
          <span class="pagination-link pagination-current">{p}</span>
        {:else}
          <a href={buildHref({ page: p })} class="pagination-link">{p}</a>
        {/if}
      {/each}

      {#if pagination.has_next}
        <a href={buildHref({ page: page + 1 })} class="pagination-link">Next</a>
      {:else}
        <span class="pagination-link pagination-disabled">Next</span>
      {/if}
    </nav>
  {/if}
</div>

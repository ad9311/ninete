<script lang="ts">
  // Shared by Index.svelte and Archived.svelte (§3.9 rule 1 keeps them as
  // separate route files; the table itself is identical apart from which
  // half of the split GetAPIRecurrentExpenses's ?archived answers).
  import { Rows3, Tag } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import { APIRequestError, get } from "../../lib/api";
  import { type Category, fetchCategories } from "../../lib/categories";
  import { formatCurrency } from "../../lib/currency";
  import { BASE_PATH, navigate } from "../../router";
  import type {
    Pagination,
    RecurrentExpense,
    RecurrentExpenseListResponse,
  } from "./types";

  interface Props {
    archived: boolean;
    basePath: string;
    search?: string;
  }

  let { archived, basePath, search = "" }: Props = $props();

  const PER_PAGE_CHOICES = [15, 25, 50, 100];

  let categories = $state<Category[]>([]);
  let rows = $state<RecurrentExpense[]>([]);
  let pagination = $state<Pagination | null>(null);
  let error = $state("");

  const params = $derived(new URLSearchParams(search));
  const categoryId = $derived(Number(params.get("category_id") ?? "0"));
  const sortField = $derived(params.get("sort_field") ?? "created_at");
  const sortOrder = $derived(params.get("sort_order") ?? "DESC");
  const page = $derived(Math.max(Number(params.get("page") ?? "1") || 1, 1));
  const perPage = $derived(
    PER_PAGE_CHOICES.includes(Number(params.get("per_page")))
      ? Number(params.get("per_page"))
      : 15,
  );
  const totalAmount = $derived(rows.reduce((sum, row) => sum + row.amount, 0));

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

    get<RecurrentExpenseListResponse>("/recurrent-expenses", {
      params: {
        archived,
        sort_field: sortField,
        sort_order: sortOrder,
        page,
        per_page: perPage,
        category_id: categoryId > 0 ? categoryId : undefined,
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

  function buildHref(overrides: {
    sort_field?: string;
    sort_order?: string;
    per_page?: number;
    page?: number;
    category_id?: number;
  }): string {
    const next = new URLSearchParams();
    next.set("sort_field", overrides.sort_field ?? sortField);
    next.set("sort_order", overrides.sort_order ?? sortOrder);
    next.set("per_page", String(overrides.per_page ?? perPage));
    next.set("page", String(overrides.page ?? page));

    const cat =
      overrides.category_id !== undefined ? overrides.category_id : categoryId;
    if (cat > 0) {
      next.set("category_id", String(cat));
    }

    return `${BASE_PATH}${basePath}?${next.toString()}`;
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
    navigate(buildHref({ category_id: value ? Number(value) : 0, page: 1 }));
  }

  function onPerPageChange(event: Event): void {
    const value = Number((event.currentTarget as HTMLSelectElement).value);
    navigate(buildHref({ per_page: value, page: 1 }));
  }

  const sortableColumns: [string, string][] = [
    ["category_id", "Category"],
    ["description", "Description"],
    ["amount", "Amount"],
    ["period", "Period (months)"],
  ];
</script>

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
</div>

{#if error}
  <p class="form-error-text">{error}</p>
{/if}

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
        <th>Runs</th>
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
          <td>{row.period}</td>
          <td>
            {#if row.occurrence_limit}
              {row.occurrence_count} of {row.occurrence_limit}
            {:else}
              <span class="chip chip-empty">Unlimited</span>
            {/if}
          </td>
          <td>
            {#if row.tags.length > 0}
              <div class="chip-list">
                {#each row.tags as tag (tag)}
                  <span class="chip chip-tag">{tag}</span>
                {/each}
              </div>
            {:else}
              <span class="chip chip-empty">No tags</span>
            {/if}
          </td>
          <td>
            <a href={`${BASE_PATH}/recurrent-expenses/${row.id}`}>Visit</a>
          </td>
        </tr>
      {/each}
    </tbody>
    <tfoot>
      <tr>
        <th colspan="7">
          Total {archived ? "archived " : ""}recurrent expenses
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

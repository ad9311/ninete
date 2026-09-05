<script lang="ts">
  // Shared by Index.svelte and Archived.svelte (§3.9 rule 1 keeps them as
  // separate route files; the table itself is identical apart from which
  // half of the split GetAPIRecurrentExpenses's ?archived answers).
  import { Tag } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import PaginationFooter from "../../components/PaginationFooter.svelte";
  import SortHeader from "../../components/SortHeader.svelte";
  import { APIRequestError, get } from "../../lib/api";
  import { type Category, fetchCategories } from "../../lib/categories";
  import { formatCurrency } from "../../lib/currency";
  import { parsePage, parsePerPage } from "../../lib/pagination";
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

  let categories = $state<Category[]>([]);
  let rows = $state<RecurrentExpense[]>([]);
  let pagination = $state<Pagination | null>(null);
  let error = $state("");

  const params = $derived(new URLSearchParams(search));
  const categoryId = $derived(Number(params.get("category_id") ?? "0"));
  const sortField = $derived(params.get("sort_field") ?? "created_at");
  const sortOrder = $derived(params.get("sort_order") ?? "DESC");
  const page = $derived(parsePage(params));
  const perPage = $derived(parsePerPage(params));
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

  function onCategoryChange(event: Event): void {
    const value = (event.currentTarget as HTMLSelectElement).value;
    navigate(buildHref({ category_id: value ? Number(value) : 0, page: 1 }));
  }

  function onPerPageChange(value: number): void {
    navigate(buildHref({ per_page: value, page: 1 }));
  }

  const sortableColumns: [string, string][] = [
    ["category_id", "Category"],
    ["description", "Description"],
    ["amount", "Amount"],
    ["period", "Period (months)"],
  ];
</script>

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
</div>

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
          <td class="font-semibold text-fg">{formatCurrency(row.amount)}</td>
          <td>{row.period}</td>
          <td>
            {#if row.occurrence_limit}
              {row.occurrence_count} of {row.occurrence_limit}
            {:else}
              <span class="chip">Unlimited</span>
            {/if}
          </td>
          <td>
            {#if row.tags.length > 0}
              <div class="flex flex-wrap gap-2">
                {#each row.tags as tag (tag)}
                  <span class="chip chip-tag">{tag}</span>
                {/each}
              </div>
            {:else}
              <span class="chip">No tags</span>
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

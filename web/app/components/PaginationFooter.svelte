<script lang="ts">
  // The footer under a paginated table: rows-per-page on the left, page links
  // on the right. Both list routes rendered their own copy of this, down to
  // the disabled Prev/Next spans, so the two could disagree about what a
  // pagination control looks like.
  //
  // Named for the footer rather than for pagination itself because each
  // resource already exports a `Pagination` type mirroring its Go struct, and
  // a route needs to import both.
  import { Rows3 } from "lucide";
  import { PER_PAGE_CHOICES, pageRange } from "../lib/pagination";
  import Icon from "./Icon.svelte";

  /**
   * The parts of a resource's pagination payload this needs. Structural on
   * purpose: each resource declares its own `Pagination` mirroring the Go
   * struct (§3.5), and this component must not depend on one of them.
   */
  interface PageInfo {
    total_pages: number;
    has_prev: boolean;
    has_next: boolean;
  }

  interface Props {
    /** Null until the first response lands; the page links stay hidden. */
    pagination: PageInfo | null;
    page: number;
    perPage: number;
    /** Builds the URL for a page number, keeping the caller's other filters. */
    hrefFor: (page: number) => string;
    onPerPageChange: (perPage: number) => void;
  }

  let { pagination, page, perPage, hrefFor, onPerPageChange }: Props = $props();

  function handlePerPage(event: Event): void {
    onPerPageChange(Number((event.currentTarget as HTMLSelectElement).value));
  }
</script>

<div class="mt-3 flex flex-wrap items-center justify-between gap-3">
  <label class="inline-flex items-center gap-2 text-muted">
    <span class="sr-only">Rows per page</span>
    <Icon icon={Rows3} class="h-4 w-4 shrink-0" />
    <select class="w-36" value={perPage} onchange={handlePerPage}>
      {#each PER_PAGE_CHOICES as choice (choice)}
        <option value={choice}>{choice} per page</option>
      {/each}
    </select>
  </label>

  {#if pagination && pagination.total_pages > 1}
    <nav class="flex flex-wrap gap-2" aria-label="Pagination">
      {#if pagination.has_prev}
        <a href={hrefFor(page - 1)} class="page-link">Prev</a>
      {:else}
        <span class="page-link page-link-disabled">Prev</span>
      {/if}

      {#each pageRange(pagination.total_pages, page) as p (p)}
        {#if p === page}
          <span class="page-link page-link-current" aria-current="page"
            >{p}</span
          >
        {:else}
          <a href={hrefFor(p)} class="page-link">{p}</a>
        {/if}
      {/each}

      {#if pagination.has_next}
        <a href={hrefFor(page + 1)} class="page-link">Next</a>
      {:else}
        <span class="page-link page-link-disabled">Next</span>
      {/if}
    </nav>
  {/if}
</div>

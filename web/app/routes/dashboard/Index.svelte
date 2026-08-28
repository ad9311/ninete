<script lang="ts">
  // Ports dashboard/index.html. The macro half of the card grid went with
  // Phase 0B (§0) — this is the expense summary alone, so no date picker: the
  // range is always this month vs. last month, computed client-side exactly
  // like Stats.svelte and List.svelte do (§3.6 of docs/spa-migration.md).
  import { SquareArrowOutUpRight } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import { APIRequestError, get } from "../../lib/api";
  import { formatCurrency } from "../../lib/currency";
  import { computeDateRange } from "../../lib/dateRanges";
  import { BASE_PATH } from "../../router";
  import type { DashboardResponse, DashboardSummary } from "./types";

  let summary = $state<DashboardSummary | null>(null);
  let error = $state("");

  $effect(() => {
    let cancelled = false;
    const thisBounds = computeDateRange("this_month");
    const lastBounds = computeDateRange("last_month");

    get<DashboardResponse>("/dashboard", {
      params: {
        this_start: thisBounds?.start,
        this_end: thisBounds?.end,
        last_start: lastBounds?.start,
        last_end: lastBounds?.end,
      },
    })
      .then((result) => {
        if (cancelled) return;
        summary = result.data;
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
</script>

{#if error}
  <p class="form-error-text">{error}</p>
{:else}
  <div class="card-grid">
    <section class="card" aria-labelledby="month-spending-card-title">
      <header class="card-header">
        <h2 id="month-spending-card-title" class="card-title">
          This month's spending
        </h2>
        <div class="card-actions">
          <a
            href={`${BASE_PATH}/expenses`}
            class="card-action-link"
            aria-label="View expenses"
            title="View expenses"
          >
            <Icon icon={SquareArrowOutUpRight} class="card-action-icon" />
          </a>
        </div>
      </header>
      {#if summary}
        <span class="card-value amount-value"
          >{formatCurrency(summary.this_month_total)}</span
        >
        <span class="card-delta">
          {#if summary.month_change_sign}
            {summary.month_change_sign}{summary.month_change_pct}% vs last month
            ({formatCurrency(summary.last_month_total)})
          {:else}
            No data for last month
          {/if}
        </span>
      {/if}
    </section>
    <section class="card" aria-labelledby="top-categories-card-title">
      <header class="card-header">
        <h2 id="top-categories-card-title" class="card-title">
          Top categories this month
        </h2>
      </header>
      {#if summary}
        {#if summary.top_categories.length > 0}
          <ul class="summary-list">
            {#each summary.top_categories as category (category.name)}
              <li class="summary-list-item">
                <span>{category.name}</span>
                <span class="amount-value"
                  >{formatCurrency(category.total)}</span
                >
              </li>
            {/each}
          </ul>
        {:else}
          <p class="card-empty">No expenses this month</p>
        {/if}
      {/if}
    </section>
  </div>
{/if}

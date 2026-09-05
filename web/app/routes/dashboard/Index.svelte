<script lang="ts">
  // Ports dashboard/index.html. The macro half of the card grid went with
  // Phase 0B (§0) — this is the expense summary alone, so no date picker: the
  // range is always this month vs. last month, computed client-side exactly
  // like Stats.svelte and List.svelte do (§3.6 of docs/spa-migration.md).
  import { SquareArrowOutUpRight } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
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
  <p class="text-danger">{error}</p>
{:else}
  <div class="grid grid-cols-[repeat(auto-fill,minmax(18rem,1fr))] gap-4">
    <Card
      title="This month's spending"
      level={2}
      actionsLabel="Spending actions"
    >
      {#snippet actions()}
        <CardAction
          icon={SquareArrowOutUpRight}
          label="View expenses"
          href={`${BASE_PATH}/expenses`}
        />
      {/snippet}
      {#if summary}
        <span class="text-2xl font-semibold text-fg">
          {formatCurrency(summary.this_month_total)}
        </span>
        <span class="text-sm text-muted">
          {#if summary.month_change_sign}
            {summary.month_change_sign}{summary.month_change_pct}% vs last month
            ({formatCurrency(summary.last_month_total)})
          {:else}
            No data for last month
          {/if}
        </span>
      {/if}
    </Card>
    <Card title="Top categories this month" level={2}>
      {#if summary}
        {#if summary.top_categories.length > 0}
          <ul class="grid gap-2">
            {#each summary.top_categories as category (category.name)}
              <li class="flex justify-between text-sm">
                <span>{category.name}</span>
                <span class="font-semibold text-fg">
                  {formatCurrency(category.total)}
                </span>
              </li>
            {/each}
          </ul>
        {:else}
          <p class="text-sm text-muted">No expenses this month</p>
        {/if}
      {/if}
    </Card>
  </div>
{/if}

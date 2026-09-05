<script lang="ts">
  // Ports expenses/stats.html. Chart.js registration and options mirror
  // chartController.ts exactly (§2.3 of docs/spa-migration.md, "Chart.js
  // arrives here rather than Phase 4" per docs/spa-migration.md's Phase 3
  // section) — same palette, same horizontal bar, same currency tooltip.
  import {
    BarController,
    BarElement,
    CategoryScale,
    Chart,
    LinearScale,
    Tooltip,
  } from "chart.js";
  import { Target, Wallet } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
  import SortHeader from "../../components/SortHeader.svelte";
  import { APIRequestError, get } from "../../lib/api";
  import { formatCurrency } from "../../lib/currency";
  import { computeDateRange, DATE_RANGE_OPTIONS } from "../../lib/dateRanges";
  import { BASE_PATH, navigate } from "../../router";
  import type { ExpenseStatRow, ExpenseStatsResponse } from "./types";

  Chart.register(
    BarController,
    BarElement,
    CategoryScale,
    LinearScale,
    Tooltip,
  );

  const PALETTE = [
    "#2d6eb0",
    "#e07b39",
    "#3aab6d",
    "#b03060",
    "#8a5cc4",
    "#c4a030",
    "#3399aa",
    "#a04040",
    "#5a8a30",
    "#7060b0",
  ];

  function categoryColor(index: number): string {
    return PALETTE[index % PALETTE.length];
  }

  function buildChart(
    canvas: HTMLCanvasElement,
    rows: ExpenseStatRow[],
  ): Chart {
    return new Chart(canvas, {
      type: "bar",
      data: {
        labels: rows.map((r) => r.name),
        datasets: [
          {
            label: "Total",
            data: rows.map((r) => r.total),
            backgroundColor: rows.map((_, i) => categoryColor(i)),
          },
        ],
      },
      options: {
        indexAxis: "y",
        responsive: true,
        plugins: {
          tooltip: {
            callbacks: {
              label: (ctx) => ` ${formatCurrency(ctx.raw as number)}`,
            },
          },
        },
      },
    });
  }

  function mountChart(canvas: HTMLCanvasElement, rows: ExpenseStatRow[]) {
    let chart = buildChart(canvas, rows);

    return {
      update(nextRows: ExpenseStatRow[]): void {
        chart.destroy();
        chart = buildChart(canvas, nextRows);
      },
      destroy(): void {
        chart.destroy();
      },
    };
  }

  interface Props {
    search?: string;
  }

  let { search = "" }: Props = $props();

  const params = $derived(new URLSearchParams(search));
  const sortField = $derived(params.get("sort_field") ?? "total");
  const sortOrder = $derived(params.get("sort_order") ?? "DESC");
  const dateRangeValue = $derived(params.get("date_range") || "this_month");

  let rows = $state<ExpenseStatRow[]>([]);
  let error = $state("");

  const totalAmount = $derived(rows.reduce((sum, row) => sum + row.total, 0));

  $effect(() => {
    let cancelled = false;
    const bounds = computeDateRange(dateRangeValue);

    get<ExpenseStatsResponse>("/expenses/stats", {
      params: {
        sort_field: sortField,
        sort_order: sortOrder,
        start: bounds?.start,
        end: bounds?.end,
      },
    })
      .then((result) => {
        if (cancelled) return;
        rows = result.data;
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

    return `${BASE_PATH}/expenses/stats?${next.toString()}`;
  }

  function sortHref(field: string): string {
    const order = sortField === field && sortOrder === "ASC" ? "DESC" : "ASC";

    return buildHref({ sort_field: field, sort_order: order });
  }

  function onDateRangeChange(event: Event): void {
    const value = (event.currentTarget as HTMLSelectElement).value;
    navigate(buildHref({ date_range: value }));
  }

  const sortableColumns: [string, string][] = [
    ["category", "Category"],
    ["total", "Total"],
  ];
</script>

<Card title="Expense stats" actionsLabel="Expense navigation">
  {#snippet actions()}
    <CardAction icon={Wallet} label="Expenses" href={`${BASE_PATH}/expenses`} />
    <CardAction
      icon={Target}
      label="Budgets"
      href={`${BASE_PATH}/expenses/budgets`}
    />
  {/snippet}
  <div class="mb-3 flex flex-wrap justify-end gap-3">
    <label class="inline-flex items-center gap-2 text-muted">
      <span class="sr-only">Date range</span>
      <select class="w-56" value={dateRangeValue} onchange={onDateRangeChange}>
        <option value="all_time">All time</option>
        {#each DATE_RANGE_OPTIONS as option (option.value)}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>
    </label>
  </div>
  {#if error}
    <p class="text-danger">{error}</p>
  {/if}
  <div class="max-w-full py-4">
    <canvas use:mountChart={rows}></canvas>
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
        </tr>
      </thead>
      <tbody>
        {#each rows as row (row.name)}
          <tr>
            <td>{row.name}</td>
            <td class="font-semibold text-fg">{formatCurrency(row.total)}</td>
          </tr>
        {/each}
      </tbody>
      <tfoot>
        <tr>
          <th colspan="2">
            Total expenses
            <span class="font-semibold text-fg">
              {formatCurrency(totalAmount)}
            </span>
          </th>
        </tr>
      </tfoot>
    </table>
  </div>
</Card>

<script lang="ts">
  // Ports expenses/budgets.html. mode (month vs months) is sent by the client
  // rather than derived from a range key server-side — see
  // GetAPIExpenseBudgets's comment in internal/handlers/handle_api_expense_budgets.go
  // for why the API needs it spelled out once tz_offset+date_range retire.
  import { Pencil, Target, Wallet } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
  import Icon from "../../components/Icon.svelte";
  import { APIRequestError, get, put } from "../../lib/api";
  import {
    centsToInputValue,
    formatCurrency,
    inputValueToCents,
  } from "../../lib/currency";
  import {
    BUDGET_DATE_RANGE_OPTIONS,
    computeDateRange,
  } from "../../lib/dateRanges";
  import { BASE_PATH, navigate } from "../../router";
  import type {
    BudgetEditRow,
    BudgetRow,
    ExpenseBudgetsResponse,
  } from "./types";

  interface Props {
    search?: string;
  }

  let { search = "" }: Props = $props();

  const params = $derived(new URLSearchParams(search));
  const dateRangeValue = $derived(
    params.get("date_range") || BUDGET_DATE_RANGE_OPTIONS[0].value,
  );
  const rangeOption = $derived(
    BUDGET_DATE_RANGE_OPTIONS.find((o) => o.value === dateRangeValue) ??
      BUDGET_DATE_RANGE_OPTIONS[0],
  );

  let rows = $state<BudgetRow[]>([]);
  let editRows = $state<BudgetEditRow[]>([]);
  let mode = $state<"month" | "months">("month");
  let loadError = $state("");
  let saveError = $state("");
  let saving = $state(false);
  let refreshToken = $state(0);

  // The multi-month rows render the same three-part header whether or not the
  // category has a budget — one is a <summary>, the other a <div>. The pointer
  // cursor is not part of it: the <div> expands nothing, so it would promise a
  // click that does not exist.
  const budgetRowClass = "flex items-center gap-3";

  // Locally editable amounts for the edit form, keyed by category id.
  let amountInputs = $state<Record<number, string | number>>({});

  const totalAmount = $derived(rows.reduce((sum, row) => sum + row.total, 0));

  $effect(() => {
    let cancelled = false;
    void refreshToken;
    const bounds =
      computeDateRange(dateRangeValue) ??
      computeDateRange(BUDGET_DATE_RANGE_OPTIONS[0].value);

    get<ExpenseBudgetsResponse>("/expenses/budgets", {
      params: {
        start: bounds?.start,
        end: bounds?.end,
        mode: rangeOption.mode,
      },
    })
      .then((result) => {
        if (cancelled) return;
        rows = result.rows;
        mode = result.mode;
        editRows = result.edit_rows;
        loadError = "";

        const seeded: Record<number, string | number> = {};
        for (const row of result.edit_rows) {
          seeded[row.category_id] =
            row.amount > 0 ? Number(centsToInputValue(row.amount)) : "";
        }
        amountInputs = seeded;
      })
      .catch((err) => {
        if (cancelled) return;
        loadError =
          err instanceof APIRequestError
            ? err.message
            : "Something went wrong.";
      });

    return () => {
      cancelled = true;
    };
  });

  function onDateRangeChange(event: Event): void {
    const value = (event.currentTarget as HTMLSelectElement).value;
    navigate(`${BASE_PATH}/expenses/budgets?date_range=${value}`);
  }

  async function saveBudgets(event: SubmitEvent): Promise<void> {
    event.preventDefault();
    saving = true;
    saveError = "";

    const amounts: Record<string, number> = {};
    for (const row of editRows) {
      amounts[String(row.category_id)] =
        inputValueToCents(amountInputs[row.category_id] ?? "") ?? 0;
    }

    try {
      await put("/expenses/budgets", { amounts });
      refreshToken += 1;
    } catch (err) {
      saveError =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
    } finally {
      saving = false;
    }
  }
</script>

<Card title="Expense budgets" actionsLabel="Expense navigation">
  {#snippet actions()}
    <CardAction icon={Wallet} label="Expenses" href={`${BASE_PATH}/expenses`} />
    <CardAction
      icon={Target}
      label="Stats"
      href={`${BASE_PATH}/expenses/stats`}
    />
  {/snippet}
  <div class="mb-3 flex flex-wrap justify-end gap-3">
    <label class="inline-flex items-center gap-2 text-muted">
      <span class="sr-only">Date range</span>
      <select class="w-56" value={dateRangeValue} onchange={onDateRangeChange}>
        {#each BUDGET_DATE_RANGE_OPTIONS as option (option.value)}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>
    </label>
  </div>
  {#if loadError}
    <p class="text-danger">{loadError}</p>
  {/if}
  <div class="overflow-x-auto">
    <table class="data-table">
      <thead>
        <tr>
          <th>Category</th>
          <th>Budget</th>
          <th>Spent</th>
          {#if mode === "month"}
            <th>Left</th>
            <th>Progress</th>
          {/if}
        </tr>
      </thead>
      <tbody>
        {#each rows as row (row.category_name)}
          <!-- `is-over` is read by app.css to recolour the bar's vendor
            pseudo-element, which no utility can reach; `text-danger` handles
            the text half. -->
          <tr class:is-over={row.over} class:text-danger={row.over}>
            {#if mode === "month"}
              <td>{row.category_name}</td>
              <td class="font-semibold">
                {row.has_budget ? formatCurrency(row.budget) : "—"}
              </td>
              <td class="font-semibold">{formatCurrency(row.total)}</td>
              <td class="font-semibold">
                {row.has_budget ? formatCurrency(row.left) : "—"}
              </td>
              <td>
                {#if row.has_budget}
                  <div class="flex min-w-32 items-center gap-2">
                    <progress class="budget-bar" max="100" value={row.bar_pct}
                    ></progress>
                    <span class="text-sm whitespace-nowrap text-muted">
                      {row.pct}%
                    </span>
                  </div>
                {/if}
              </td>
            {:else}
              <td colspan="3">
                {#if row.has_budget}
                  <details>
                    <summary class="{budgetRowClass} cursor-pointer">
                      <span class="font-medium">{row.category_name}</span>
                      <span class="font-semibold">
                        {formatCurrency(row.total)}
                      </span>
                      <span class="text-sm text-muted">
                        {formatCurrency(row.budget)}/mo
                      </span>
                    </summary>
                    <p class="mt-2 text-sm text-muted">
                      {row.months_over} of {row.month_count} months over · avg {formatCurrency(
                        row.avg_per_month,
                      )}
                    </p>
                    <ul class="mt-2">
                      {#each row.months as month (month.month)}
                        <li
                          class="flex items-center gap-3 py-1"
                          class:is-over={month.over}
                          class:text-danger={month.over}
                        >
                          <span class="min-w-20 text-sm text-muted">
                            {month.month}
                          </span>
                          <span class="font-semibold">
                            {formatCurrency(month.total)}
                          </span>
                          <div class="flex min-w-32 items-center gap-2">
                            <progress
                              class="budget-bar"
                              max="100"
                              value={month.bar_pct}
                            ></progress>
                            <span class="text-sm whitespace-nowrap text-muted">
                              {month.pct}%
                            </span>
                          </div>
                        </li>
                      {/each}
                    </ul>
                  </details>
                {:else}
                  <div class={budgetRowClass}>
                    <span class="font-medium">{row.category_name}</span>
                    <span class="font-semibold">
                      {formatCurrency(row.total)}
                    </span>
                    <span class="text-sm text-muted">—</span>
                  </div>
                {/if}
              </td>
            {/if}
          </tr>
        {/each}
      </tbody>
      <tfoot>
        <tr>
          <th colspan={mode === "month" ? 5 : 3}>
            Total expenses
            <span class="font-semibold text-fg">
              {formatCurrency(totalAmount)}
            </span>
          </th>
        </tr>
      </tfoot>
    </table>
  </div>
  <details class="mt-4">
    <summary
      class="inline-flex w-fit cursor-pointer items-center gap-2 py-1 text-sm text-muted hover:text-primary"
    >
      <Icon icon={Pencil} class="h-4 w-4" />
      Edit budgets
    </summary>
    {#if saveError}
      <p class="text-danger">{saveError}</p>
    {/if}
    <form onsubmit={saveBudgets} class="grid max-w-form gap-3">
      <p class="my-2 text-sm text-muted">
        A blank amount clears that category's budget.
      </p>
      {#each editRows as row (row.category_id)}
        <label>
          {row.name}
          <input
            type="number"
            min="0"
            step="0.01"
            value={amountInputs[row.category_id] ?? ""}
            oninput={(event) => {
              amountInputs[row.category_id] = (
                event.currentTarget as HTMLInputElement
              ).value;
            }}
          />
        </label>
      {/each}
      <button
        type="submit"
        class="btn btn-primary mt-3 justify-self-end"
        disabled={saving}
      >
        {saving ? "Saving..." : "Save budgets"}
      </button>
    </form>
  </details>
</Card>

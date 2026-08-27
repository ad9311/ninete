<script lang="ts">
  // Ports expenses/budgets.html. mode (month vs months) is sent by the client
  // rather than derived from a range key server-side — see
  // GetAPIExpenseBudgets's comment in internal/handlers/handle_api_expense_budgets.go
  // for why the API needs it spelled out once tz_offset+date_range retire.
  import { Pencil, Target, Wallet } from "lucide";
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

<section class="card" aria-labelledby="expense-budgets-card-title">
  <header class="card-header">
    <h1 id="expense-budgets-card-title" class="card-title">Expense budgets</h1>
    <nav class="card-actions" aria-label="Expense navigation">
      <a
        href={`${BASE_PATH}/expenses`}
        class="card-action-link"
        aria-label="Expenses"
        title="Expenses"
      >
        <Icon icon={Wallet} class="card-action-icon" />
      </a>
      <a
        href={`${BASE_PATH}/expenses/stats`}
        class="card-action-link"
        aria-label="Stats"
        title="Stats"
      >
        <Icon icon={Target} class="card-action-icon" />
      </a>
    </nav>
  </header>
  <div class="filters">
    <label>
      <span class="sr-only">Date range</span>
      <select value={dateRangeValue} onchange={onDateRangeChange}>
        {#each BUDGET_DATE_RANGE_OPTIONS as option (option.value)}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>
    </label>
  </div>
  {#if loadError}
    <p class="form-error-text">{loadError}</p>
  {/if}
  <div class="table-scroll">
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
          {#if mode === "month"}
            <tr class:budget-row-over={row.over}>
              <td>{row.category_name}</td>
              <td class="amount-value"
                >{row.has_budget ? formatCurrency(row.budget) : "—"}</td
              >
              <td class="amount-value">{formatCurrency(row.total)}</td>
              <td class="amount-value"
                >{row.has_budget ? formatCurrency(row.left) : "—"}</td
              >
              <td>
                {#if row.has_budget}
                  <div class="budget-progress">
                    <progress max="100" value={row.bar_pct}></progress>
                    <span class="budget-percent">{row.pct}%</span>
                  </div>
                {/if}
              </td>
            </tr>
          {:else}
            <tr class:budget-row-over={row.over}>
              <td colspan="3">
                {#if row.has_budget}
                  <details class="budget-months">
                    <summary class="budget-summary">
                      <span class="budget-category">{row.category_name}</span>
                      <span class="amount-value"
                        >{formatCurrency(row.total)}</span
                      >
                      <span class="budget-per-month"
                        >{formatCurrency(row.budget)}/mo</span
                      >
                    </summary>
                    <p class="budget-months-note">
                      {row.months_over} of {row.month_count} months over · avg {formatCurrency(
                        row.avg_per_month,
                      )}
                    </p>
                    <ul class="budget-month-list">
                      {#each row.months as month (month.month)}
                        <li
                          class="budget-month"
                          class:budget-row-over={month.over}
                        >
                          <span class="budget-month-label">{month.month}</span>
                          <span class="amount-value"
                            >{formatCurrency(month.total)}</span
                          >
                          <div class="budget-progress">
                            <progress max="100" value={month.bar_pct}
                            ></progress>
                            <span class="budget-percent">{month.pct}%</span>
                          </div>
                        </li>
                      {/each}
                    </ul>
                  </details>
                {:else}
                  <div class="budget-summary">
                    <span class="budget-category">{row.category_name}</span>
                    <span class="amount-value">{formatCurrency(row.total)}</span
                    >
                    <span class="budget-per-month">—</span>
                  </div>
                {/if}
              </td>
            </tr>
          {/if}
        {/each}
      </tbody>
      <tfoot>
        <tr>
          <th colspan={mode === "month" ? 5 : 3}>
            Total expenses
            <span class="amount-value">{formatCurrency(totalAmount)}</span>
          </th>
        </tr>
      </tfoot>
    </table>
  </div>
  <details class="budget-edit">
    <summary class="search-summary">
      <Icon icon={Pencil} class="search-caret" />
      Edit budgets
    </summary>
    {#if saveError}
      <p class="form-error-text">{saveError}</p>
    {/if}
    <form onsubmit={saveBudgets}>
      <p class="budget-edit-hint">
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
      <button type="submit" class="btn-primary form-submit" disabled={saving}>
        {saving ? "Saving..." : "Save budgets"}
      </button>
    </form>
  </details>
</section>

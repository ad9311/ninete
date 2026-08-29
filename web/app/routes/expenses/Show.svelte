<script lang="ts">
  import { Repeat, SquarePen, Wallet } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import LocalDate from "../../components/LocalDate.svelte";
  import { APIRequestError, del, get } from "../../lib/api";
  import { formatCurrency } from "../../lib/currency";
  import { BASE_PATH, navigate } from "../../router";
  import type { Expense } from "./types";

  let { id }: { id: string } = $props();

  let expense = $state<Expense | null>(null);
  let loadError = $state("");

  $effect(() => {
    let cancelled = false;

    get<Expense>(`/expenses/${id}`)
      .then((result) => {
        if (cancelled) return;
        expense = result;
        // Cleared on success, same as Edit.svelte: without this a failed load
        // leaves its banner over the next expense the route resolves to.
        loadError = "";
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

  async function deleteExpense(): Promise<void> {
    if (!window.confirm("Delete this expense?")) return;

    try {
      await del(`/expenses/${id}`);
      navigate("/expenses");
    } catch (err) {
      loadError =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
    }
  }
</script>

<section class="card" aria-labelledby="expense-card-title">
  <header class="card-header">
    <h1 id="expense-card-title" class="card-title">Expense</h1>
    <nav class="card-actions" aria-label="Expense navigation">
      <a
        href={`${BASE_PATH}/expenses/${id}/edit`}
        class="card-action-link"
        aria-label="Edit expense"
        title="Edit expense"
      >
        <Icon icon={SquarePen} class="card-action-icon" />
      </a>
      <a
        href={`${BASE_PATH}/expenses`}
        class="card-action-link"
        aria-label="Expenses"
        title="Expenses"
      >
        <Icon icon={Wallet} class="card-action-icon" />
      </a>
      <a
        href={`${BASE_PATH}/recurrent-expenses`}
        class="card-action-link"
        aria-label="Recurrent expenses"
        title="Recurrent expenses"
      >
        <Icon icon={Repeat} class="card-action-icon" />
      </a>
    </nav>
  </header>
  {#if loadError}
    <p class="form-error-text">{loadError}</p>
  {/if}
  {#if expense}
    <table>
      <tbody>
        <tr>
          <th>Category</th>
          <td>{expense.category_name}</td>
        </tr>
        <tr>
          <th>Description</th>
          <td>{expense.description}</td>
        </tr>
        <tr>
          <th>Amount</th>
          <td class="amount-value">{formatCurrency(expense.amount)}</td>
        </tr>
        <tr>
          <th>Billed</th>
          <td><LocalDate value={expense.date} /></td>
        </tr>
        <tr>
          <th>Created</th>
          <td><LocalDate value={expense.created_at} datetime /></td>
        </tr>
        <tr>
          <th>Tags</th>
          <td>
            {#if expense.tags.length > 0}
              <div class="chip-list">
                {#each expense.tags as tag (tag)}
                  <span class="chip chip-tag">{tag}</span>
                {/each}
              </div>
            {:else}
              <span class="chip chip-empty">No tags</span>
            {/if}
          </td>
        </tr>
      </tbody>
    </table>
    <button
      type="button"
      class="btn-danger form-submit"
      onclick={deleteExpense}
    >
      Delete
    </button>
  {/if}
</section>

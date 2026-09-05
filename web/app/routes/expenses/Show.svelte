<script lang="ts">
  import { Repeat, SquarePen, Wallet } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
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

<Card title="Expense" actionsLabel="Expense navigation">
  {#snippet actions()}
    <CardAction
      icon={SquarePen}
      label="Edit expense"
      href={`${BASE_PATH}/expenses/${id}/edit`}
    />
    <CardAction icon={Wallet} label="Expenses" href={`${BASE_PATH}/expenses`} />
    <CardAction
      icon={Repeat}
      label="Recurrent expenses"
      href={`${BASE_PATH}/recurrent-expenses`}
    />
  {/snippet}
  {#if loadError}
    <p class="text-danger">{loadError}</p>
  {/if}
  {#if expense}
    <div class="overflow-x-auto">
      <table class="data-table">
        <tbody>
          <tr>
            <th>Category</th>
            <td>{expense.category_name}</td>
          </tr>
          <tr>
            <th>Description</th>
            <!-- `.data-table td:last-child` is nowrap for the list tables'
              action column; every cell here is a last child, and a 50-character
              description has to be allowed to wrap on a phone. -->
            <td class="whitespace-normal">{expense.description}</td>
          </tr>
          <tr>
            <th>Amount</th>
            <td class="font-semibold text-fg">
              {formatCurrency(expense.amount)}
            </td>
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
                <div class="flex flex-wrap gap-2">
                  {#each expense.tags as tag (tag)}
                    <span class="chip chip-tag">{tag}</span>
                  {/each}
                </div>
              {:else}
                <span class="chip">No tags</span>
              {/if}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <button
      type="button"
      class="btn btn-danger mt-3 justify-self-end"
      onclick={deleteExpense}
    >
      Delete
    </button>
  {/if}
</Card>

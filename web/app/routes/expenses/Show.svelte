<script lang="ts">
  import { Repeat, SquarePen, Wallet } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
  import { APIRequestError, del, get } from "../../lib/api";
  import { BASE_PATH, navigate } from "../../router";
  import Details from "./Details.svelte";
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
    <Details {expense} />
    <button
      type="button"
      class="btn btn-danger mt-3 justify-self-end"
      onclick={deleteExpense}
    >
      Delete
    </button>
  {/if}
</Card>

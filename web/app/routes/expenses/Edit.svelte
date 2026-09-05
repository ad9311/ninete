<script lang="ts">
  import { Eye, Repeat, Wallet } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
  import { APIRequestError, get, put } from "../../lib/api";
  import { BASE_PATH, navigate } from "../../router";
  import Form from "./Form.svelte";
  import type { Expense, ExpenseRequestBody } from "./types";

  let { id }: { id: string } = $props();

  let expense = $state<Expense | null>(null);
  let loadError = $state("");
  let submitError = $state("");
  let pending = $state(false);

  $effect(() => {
    let cancelled = false;

    get<Expense>(`/expenses/${id}`)
      .then((result) => {
        if (cancelled) return;
        expense = result;
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

  async function handleSubmit(body: ExpenseRequestBody): Promise<void> {
    pending = true;
    submitError = "";

    try {
      await put<Expense>(`/expenses/${id}`, body);
      navigate(`/expenses/${id}`);
    } catch (err) {
      submitError =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
    } finally {
      pending = false;
    }
  }
</script>

<Card title="Edit expense" actionsLabel="Expense navigation">
  {#snippet actions()}
    <CardAction
      icon={Eye}
      label="View expense"
      href={`${BASE_PATH}/expenses/${id}`}
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
  {:else if expense}
    <!-- Remounts on a param-only route change — see the identical comment in
      recurrent_expenses/Edit.svelte. -->
    {#key expense.id}
      <Form
        initial={expense}
        submitLabel="Submit"
        error={submitError}
        {pending}
        onSubmit={handleSubmit}
      />
    {/key}
  {/if}
</Card>

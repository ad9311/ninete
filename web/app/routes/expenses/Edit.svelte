<script lang="ts">
  import { Eye, Repeat, Wallet } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import { APIRequestError, get, put } from "../../lib/api";
  import { navigate } from "../../router";
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

<section class="card" aria-labelledby="edit-expense-card-title">
  <header class="card-header">
    <h1 id="edit-expense-card-title" class="card-title">Edit expense</h1>
    <nav class="card-actions" aria-label="Expense navigation">
      <a
        href={`/app/expenses/${id}`}
        class="card-action-link"
        aria-label="View expense"
        title="View expense"
      >
        <Icon icon={Eye} class="card-action-icon" />
      </a>
      <a
        href="/app/expenses"
        class="card-action-link"
        aria-label="Expenses"
        title="Expenses"
      >
        <Icon icon={Wallet} class="card-action-icon" />
      </a>
      <a
        href="/app/recurrent-expenses"
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
</section>

<script lang="ts">
  import { Eye, Repeat } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import { APIRequestError, get, patch } from "../../lib/api";
  import { navigate } from "../../router";
  import Form from "./Form.svelte";
  import type { RecurrentExpense, RecurrentExpenseRequestBody } from "./types";

  let { id }: { id: string } = $props();

  let recurrentExpense = $state<RecurrentExpense | null>(null);
  let loadError = $state("");
  let submitError = $state("");
  let pending = $state(false);

  $effect(() => {
    let cancelled = false;

    get<RecurrentExpense>(`/recurrent-expenses/${id}`)
      .then((result) => {
        if (!cancelled) recurrentExpense = result;
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

  async function handleSubmit(
    body: RecurrentExpenseRequestBody,
  ): Promise<void> {
    pending = true;
    submitError = "";

    try {
      await patch<RecurrentExpense>(`/recurrent-expenses/${id}`, body);
      navigate(`/recurrent-expenses/${id}`);
    } catch (err) {
      submitError =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
    } finally {
      pending = false;
    }
  }
</script>

<section class="card" aria-labelledby="edit-recurrent-expense-card-title">
  <header class="card-header">
    <h1 id="edit-recurrent-expense-card-title" class="card-title">
      Edit recurrent expense
    </h1>
    <nav class="card-actions" aria-label="Recurrent expense navigation">
      <a
        href={`/app/recurrent-expenses/${id}`}
        class="card-action-link"
        aria-label="View recurrent expense"
        title="View recurrent expense"
      >
        <Icon icon={Eye} class="card-action-icon" />
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
  {:else if recurrentExpense}
    <Form
      initial={recurrentExpense}
      submitLabel="Submit"
      error={submitError}
      {pending}
      onSubmit={handleSubmit}
    />
  {/if}
</section>

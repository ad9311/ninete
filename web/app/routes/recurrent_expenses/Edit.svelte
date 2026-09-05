<script lang="ts">
  import { Eye, Repeat } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
  import { APIRequestError, get, put } from "../../lib/api";
  import { BASE_PATH, navigate } from "../../router";
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
        if (cancelled) return;
        recurrentExpense = result;
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

  async function handleSubmit(
    body: RecurrentExpenseRequestBody,
  ): Promise<void> {
    pending = true;
    submitError = "";

    try {
      await put<RecurrentExpense>(`/recurrent-expenses/${id}`, body);
      navigate(`/recurrent-expenses/${id}`);
    } catch (err) {
      submitError =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
    } finally {
      pending = false;
    }
  }
</script>

<Card
  title="Edit recurrent expense"
  actionsLabel="Recurrent expense navigation"
>
  {#snippet actions()}
    <CardAction
      icon={Eye}
      label="View recurrent expense"
      href={`${BASE_PATH}/recurrent-expenses/${id}`}
    />
    <CardAction
      icon={Repeat}
      label="Recurrent expenses"
      href={`${BASE_PATH}/recurrent-expenses`}
    />
  {/snippet}
  {#if loadError}
    <p class="text-danger">{loadError}</p>
  {:else if recurrentExpense}
    <!-- Form seeds its fields from `initial` once and then diverges as the
      user types, so a new record arriving in the same component instance
      (App.svelte reuses one across a param-only route change) has to remount
      it — otherwise the fields would still hold the previous record while the
      submit would have written those values to this one. -->
    {#key recurrentExpense.id}
      <Form
        initial={recurrentExpense}
        submitLabel="Submit"
        error={submitError}
        {pending}
        onSubmit={handleSubmit}
      />
    {/key}
  {/if}
</Card>

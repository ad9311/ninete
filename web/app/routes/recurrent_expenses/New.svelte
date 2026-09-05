<script lang="ts">
  import { Repeat } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
  import { APIRequestError, post } from "../../lib/api";
  import { BASE_PATH, navigate } from "../../router";
  import Form from "./Form.svelte";
  import type { RecurrentExpense, RecurrentExpenseRequestBody } from "./types";

  let error = $state("");
  let pending = $state(false);

  async function handleSubmit(
    body: RecurrentExpenseRequestBody,
  ): Promise<void> {
    pending = true;
    error = "";

    try {
      await post<RecurrentExpense>("/recurrent-expenses", body);
      navigate("/recurrent-expenses");
    } catch (err) {
      error =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
    } finally {
      pending = false;
    }
  }
</script>

<Card title="New recurrent expense" actionsLabel="Recurrent expense navigation">
  {#snippet actions()}
    <CardAction
      icon={Repeat}
      label="Recurrent expenses"
      href={`${BASE_PATH}/recurrent-expenses`}
    />
  {/snippet}
  <Form submitLabel="Submit" {error} {pending} onSubmit={handleSubmit} />
</Card>

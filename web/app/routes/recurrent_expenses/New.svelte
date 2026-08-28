<script lang="ts">
  import { Repeat } from "lucide";
  import Icon from "../../components/Icon.svelte";
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

<section class="card" aria-labelledby="new-recurrent-expense-card-title">
  <header class="card-header">
    <h1 id="new-recurrent-expense-card-title" class="card-title">
      New recurrent expense
    </h1>
    <nav class="card-actions" aria-label="Recurrent expense navigation">
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
  <Form submitLabel="Submit" {error} {pending} onSubmit={handleSubmit} />
</section>

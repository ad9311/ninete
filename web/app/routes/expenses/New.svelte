<script lang="ts">
  // Ports expenses/new.html. quickExpenseController.ts's `data-quick-force`
  // has no equivalent here: it existed only to survive a full-page re-render
  // after a failed quick submission, and QuickAddForm never unmounts on its
  // own error path, so the remembered toggle (localStorage, same key) is the
  // whole of the port.
  import { Repeat, Wallet } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import { APIRequestError, post } from "../../lib/api";
  import { BASE_PATH, navigate } from "../../router";
  import Form from "./Form.svelte";
  import QuickAddForm from "./QuickAddForm.svelte";
  import type { Expense, ExpenseRequestBody } from "./types";

  const QUICK_MODE_KEY = "expense-quick-mode";

  function readQuickMode(): boolean {
    try {
      return localStorage.getItem(QUICK_MODE_KEY) === "1";
    } catch {
      return false;
    }
  }

  let quick = $state(readQuickMode());
  let error = $state("");
  let pending = $state(false);

  function toggleQuick(): void {
    quick = !quick;
    try {
      localStorage.setItem(QUICK_MODE_KEY, quick ? "1" : "0");
    } catch {
      // Persistence is best-effort, same as the recurrent expense form.
    }
  }

  async function handleSubmit(body: ExpenseRequestBody): Promise<void> {
    pending = true;
    error = "";

    try {
      await post<Expense>("/expenses", body);
      navigate("/expenses");
    } catch (err) {
      error =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
    } finally {
      pending = false;
    }
  }
</script>

<section class="card" aria-labelledby="new-expense-card-title">
  <header class="card-header">
    <h1 id="new-expense-card-title" class="card-title">New expense</h1>
    <nav class="card-actions" aria-label="Expense navigation">
      <label class="quick-toggle" title="Quick add">
        <input type="checkbox" checked={quick} onchange={toggleQuick} />
        Quick
      </label>
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
  {#if quick}
    <QuickAddForm />
  {:else}
    <Form submitLabel="Submit" {error} {pending} onSubmit={handleSubmit} />
  {/if}
</section>

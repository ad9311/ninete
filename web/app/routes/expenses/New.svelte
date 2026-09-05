<script lang="ts">
  // Ports expenses/new.html. quickExpenseController.ts's `data-quick-force`
  // has no equivalent here: it existed only to survive a full-page re-render
  // after a failed quick submission, and QuickAddForm never unmounts on its
  // own error path, so the remembered toggle (localStorage, same key) is the
  // whole of the port.
  import { Repeat, Wallet } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
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

<Card title="New expense" actionsLabel="Expense navigation">
  {#snippet actions()}
    <label
      class="inline-flex cursor-pointer items-center gap-2 text-sm font-medium text-fg select-none"
      title="Quick add"
    >
      <input type="checkbox" checked={quick} onchange={toggleQuick} />
      Quick
    </label>
    <CardAction icon={Wallet} label="Expenses" href={`${BASE_PATH}/expenses`} />
    <CardAction
      icon={Repeat}
      label="Recurrent expenses"
      href={`${BASE_PATH}/recurrent-expenses`}
    />
  {/snippet}
  {#if quick}
    <QuickAddForm />
  {:else}
    <Form submitLabel="Submit" {error} {pending} onSubmit={handleSubmit} />
  {/if}
</Card>

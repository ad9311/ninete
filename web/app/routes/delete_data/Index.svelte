<script lang="ts">
  // Each section gates a real DELETE through lib/api.ts behind confirm().
  // These are account-scale destructive actions, so the confirmation is
  // deliberate (see routes/recurrent_expenses/Show.svelte's note on this
  // pattern at resource scale).
  import { ArrowLeft } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import { APIRequestError, del, get } from "../../lib/api";
  import { BASE_PATH } from "../../router";
  import type { DeleteDataCounts, DeleteDataCountsResponse } from "./types";

  let counts = $state<DeleteDataCounts | null>(null);
  let error = $state("");
  let pending = $state(false);

  // No cancel guard here, unlike the mount-time fetch below: this only runs in
  // response to a resolved delete the user just triggered, matching how
  // Show.svelte's unarchive() re-fetches after a POST.
  async function reloadCounts(): Promise<void> {
    const result = await get<DeleteDataCountsResponse>("/delete-data");
    counts = result.data;
    error = "";
  }

  $effect(() => {
    let cancelled = false;

    get<DeleteDataCountsResponse>("/delete-data")
      .then((result) => {
        if (cancelled) return;
        counts = result.data;
        error = "";
      })
      .catch((err) => {
        if (cancelled) return;
        error =
          err instanceof APIRequestError
            ? err.message
            : "Something went wrong.";
      });

    return () => {
      cancelled = true;
    };
  });

  async function deleteSection(
    confirmMessage: string,
    path: string,
  ): Promise<void> {
    if (!confirm(confirmMessage)) return;

    pending = true;
    try {
      await del(path);
      await reloadCounts();
    } catch (err) {
      error =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
    } finally {
      pending = false;
    }
  }
</script>

<section class="card" aria-labelledby="account-delete-data-title">
  <header class="card-header">
    <h1 id="account-delete-data-title" class="card-title">Delete data</h1>
    <nav class="card-actions" aria-label="Delete data actions">
      <a
        href={`${BASE_PATH}/account`}
        class="card-action-link"
        aria-label="Back to account"
        title="Back to account"
      >
        <Icon icon={ArrowLeft} class="card-action-icon" />
      </a>
    </nav>
  </header>
  <p class="card-empty">
    Each action removes every record of that type for your account and cannot be
    undone.
  </p>
  {#if error}
    <p class="form-error-text">{error}</p>
  {/if}
</section>

{#if counts}
  <div class="card-grid">
    <section class="card" aria-labelledby="account-expenses-title">
      <header class="card-header">
        <h2 id="account-expenses-title" class="card-title">Expenses</h2>
      </header>
      <span class="card-delta">{counts.expenses} record(s)</span>
      <button
        type="button"
        class="btn-danger form-submit"
        disabled={pending}
        onclick={() =>
          deleteSection(
            "Delete ALL your expenses? This cannot be undone.",
            "/delete-data/expenses",
          )}
      >
        Delete
      </button>
    </section>

    <section class="card" aria-labelledby="account-recurrent-expenses-title">
      <header class="card-header">
        <h2 id="account-recurrent-expenses-title" class="card-title">
          Recurrent Expenses
        </h2>
      </header>
      <span class="card-delta">{counts.recurrent_expenses} record(s)</span>
      <button
        type="button"
        class="btn-danger form-submit"
        disabled={pending}
        onclick={() =>
          deleteSection(
            "Delete ALL your recurrent expenses? This cannot be undone.",
            "/delete-data/recurrent-expenses",
          )}
      >
        Delete
      </button>
    </section>

    <section class="card" aria-labelledby="account-expense-budgets-title">
      <header class="card-header">
        <h2 id="account-expense-budgets-title" class="card-title">
          Expense Budgets
        </h2>
      </header>
      <span class="card-delta">{counts.expense_budgets} record(s)</span>
      <button
        type="button"
        class="btn-danger form-submit"
        disabled={pending}
        onclick={() =>
          deleteSection(
            "Delete your expense budgets? This cannot be undone.",
            "/delete-data/expense-budgets",
          )}
      >
        Delete
      </button>
    </section>

    <section class="card" aria-labelledby="account-tags-title">
      <header class="card-header">
        <h2 id="account-tags-title" class="card-title">Tags</h2>
      </header>
      <span class="card-delta">{counts.tags} record(s)</span>
      <button
        type="button"
        class="btn-danger form-submit"
        disabled={pending}
        onclick={() =>
          deleteSection(
            "Delete ALL your tags? This cannot be undone.",
            "/delete-data/tags",
          )}
      >
        Delete
      </button>
    </section>
  </div>

  <section class="card" aria-labelledby="account-delete-all-title">
    <header class="card-header">
      <h2 id="account-delete-all-title" class="card-title">
        Delete everything
      </h2>
    </header>
    <p class="card-empty">
      This removes every record across all sections above in one action. This
      cannot be undone.
    </p>
    <button
      type="button"
      class="btn-danger form-submit"
      disabled={pending}
      onclick={() =>
        deleteSection(
          "Delete ALL of your data across every section? This CANNOT be undone.",
          "/delete-data",
        )}
    >
      Delete
    </button>
  </section>
{/if}

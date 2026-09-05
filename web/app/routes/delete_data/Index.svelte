<script lang="ts">
  // Each section gates a real DELETE through lib/api.ts behind confirm().
  // These are account-scale destructive actions, so the confirmation is
  // deliberate (see routes/recurrent_expenses/Show.svelte's note on this
  // pattern at resource scale).
  import { ArrowLeft } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
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

  // The four scoped sections differ only in wording, count and endpoint, so
  // they are a table rather than four near-identical blocks of markup — the
  // shape that let one card's confirm() text drift from its endpoint before.
  const SECTIONS: {
    title: string;
    path: string;
    confirm: string;
    count: (counts: DeleteDataCounts) => number;
  }[] = [
    {
      title: "Expenses",
      path: "/delete-data/expenses",
      confirm: "Delete ALL your expenses? This cannot be undone.",
      count: (c) => c.expenses,
    },
    {
      title: "Recurrent Expenses",
      path: "/delete-data/recurrent-expenses",
      confirm: "Delete ALL your recurrent expenses? This cannot be undone.",
      count: (c) => c.recurrent_expenses,
    },
    {
      title: "Expense Budgets",
      path: "/delete-data/expense-budgets",
      confirm: "Delete your expense budgets? This cannot be undone.",
      count: (c) => c.expense_budgets,
    },
    {
      title: "Tags",
      path: "/delete-data/tags",
      confirm: "Delete ALL your tags? This cannot be undone.",
      count: (c) => c.tags,
    },
  ];

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

<Card title="Delete data" actionsLabel="Delete data actions">
  {#snippet actions()}
    <CardAction
      icon={ArrowLeft}
      label="Back to account"
      href={`${BASE_PATH}/account`}
    />
  {/snippet}
  <p class="text-sm text-muted">
    Each action removes every record of that type for your account and cannot be
    undone.
  </p>
  {#if error}
    <p class="text-danger">{error}</p>
  {/if}
</Card>

{#if counts}
  {@const loaded = counts}
  <div class="grid grid-cols-[repeat(auto-fill,minmax(18rem,1fr))] gap-4">
    {#each SECTIONS as section (section.path)}
      <Card title={section.title} level={2}>
        <span class="text-sm text-muted">
          {section.count(loaded)} record(s)
        </span>
        <button
          type="button"
          class="btn btn-danger mt-3 justify-self-end"
          disabled={pending}
          onclick={() => deleteSection(section.confirm, section.path)}
        >
          Delete
        </button>
      </Card>
    {/each}
  </div>

  <Card title="Delete everything" level={2}>
    <p class="text-sm text-muted">
      This removes every record across all sections above in one action. This
      cannot be undone.
    </p>
    <button
      type="button"
      class="btn btn-danger mt-3 justify-self-end"
      disabled={pending}
      onclick={() =>
        deleteSection(
          "Delete ALL of your data across every section? This CANNOT be undone.",
          "/delete-data",
        )}
    >
      Delete
    </button>
  </Card>
{/if}

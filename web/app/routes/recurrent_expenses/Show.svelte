<script lang="ts">
  // Destructive actions call window.confirm() directly before the request
  // (§5, Phase 5 note in docs/spa-migration.md covers confirm() for
  // account-wide actions; this is the same mechanism at resource scale).
  import { Repeat, SquarePen } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
  import { APIRequestError, del, get, post } from "../../lib/api";
  import { formatCurrency } from "../../lib/currency";
  import { BASE_PATH, navigate } from "../../router";
  import type { RecurrentExpense } from "./types";

  let { id }: { id: string } = $props();

  let recurrentExpense = $state<RecurrentExpense | null>(null);
  let error = $state("");
  let pending = $state(false);

  $effect(() => {
    let cancelled = false;

    get<RecurrentExpense>(`/recurrent-expenses/${id}`)
      .then((result) => {
        if (cancelled) return;
        recurrentExpense = result;
        // Clearing on success matters as much as setting on failure: a
        // retried unarchive that works must not leave the previous
        // attempt's banner sitting above the record it just fixed.
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

  async function unarchive(): Promise<void> {
    if (
      !confirm(
        "Unarchive this recurrent expense? Its run count goes back to zero.",
      )
    ) {
      return;
    }

    pending = true;
    try {
      recurrentExpense = await post<RecurrentExpense>(
        `/recurrent-expenses/${id}/unarchive`,
      );
      error = "";
    } catch (err) {
      error =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
    } finally {
      pending = false;
    }
  }

  async function deleteRecurrentExpense(): Promise<void> {
    if (!confirm("Delete this recurrent expense?")) {
      return;
    }

    pending = true;
    try {
      await del(`/recurrent-expenses/${id}`);
      navigate("/recurrent-expenses");
    } catch (err) {
      error =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
      pending = false;
    }
  }
</script>

<Card title="Recurrent expense" actionsLabel="Recurrent expense navigation">
  {#snippet actions()}
    <CardAction
      icon={SquarePen}
      label="Edit recurrent expense"
      href={`${BASE_PATH}/recurrent-expenses/${id}/edit`}
    />
    <CardAction
      icon={Repeat}
      label="Recurrent expenses"
      href={`${BASE_PATH}/recurrent-expenses`}
    />
  {/snippet}
  {#if error}
    <p class="text-danger">{error}</p>
  {/if}
  {#if recurrentExpense}
    {@const re = recurrentExpense}
    <div class="overflow-x-auto">
      <table class="data-table">
        <tbody>
          <tr>
            <th>Category</th>
            <td>{re.category_name}</td>
          </tr>
          <tr>
            <th>Description</th>
            <td>{re.description}</td>
          </tr>
          <tr>
            <th>Amount</th>
            <td class="font-semibold text-fg">{formatCurrency(re.amount)}</td>
          </tr>
          <tr>
            <th>Period (months)</th>
            <td>{re.period}</td>
          </tr>
          <tr>
            <th>Runs</th>
            <td>
              {#if re.occurrence_limit}
                {re.occurrence_count} of {re.occurrence_limit}
              {:else}
                <span class="chip">Unlimited</span>
              {/if}
            </td>
          </tr>
          <tr>
            <th>Status</th>
            <td>
              <span class="chip chip-tag">
                {re.archived ? "Archived" : "Active"}
              </span>
            </td>
          </tr>
          <tr>
            <th>Tags</th>
            <td>
              {#if re.tags.length > 0}
                <div class="flex flex-wrap gap-2">
                  {#each re.tags as tag (tag)}
                    <span class="chip chip-tag">{tag}</span>
                  {/each}
                </div>
              {:else}
                <span class="chip">No tags</span>
              {/if}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div class="mt-3 flex justify-end gap-2">
      {#if re.archived}
        <button
          type="button"
          class="btn btn-primary"
          disabled={pending}
          onclick={unarchive}
        >
          Unarchive
        </button>
      {/if}
      <button
        type="button"
        class="btn btn-danger"
        disabled={pending}
        onclick={deleteRecurrentExpense}
      >
        Delete
      </button>
    </div>
  {/if}
</Card>

<script lang="ts">
  // data-turbo-confirm in the legacy show.html is Turbo's wrapper around
  // window.confirm() before it submits the form — calling confirm() directly
  // here is the same deliberateness, not a weaker substitute (§5, Phase 5
  // note in docs/spa-migration.md about confirm() applies to destructive
  // account-wide actions; this is the same mechanism at resource scale).
  import { Repeat, SquarePen } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import { APIRequestError, del, get, post } from "../../lib/api";
  import { formatCurrency } from "../../lib/currency";
  import { navigate } from "../../router";
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

<section class="card" aria-labelledby="recurrent-expense-card-title">
  <header class="card-header">
    <h1 id="recurrent-expense-card-title" class="card-title">
      Recurrent expense
    </h1>
    <nav class="card-actions" aria-label="Recurrent expense navigation">
      <a
        href={`/app/recurrent-expenses/${id}/edit`}
        class="card-action-link"
        aria-label="Edit recurrent expense"
        title="Edit recurrent expense"
      >
        <Icon icon={SquarePen} class="card-action-icon" />
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
  {#if error}
    <p class="form-error-text">{error}</p>
  {/if}
  {#if recurrentExpense}
    {@const re = recurrentExpense}
    <table>
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
          <td class="amount-value">{formatCurrency(re.amount)}</td>
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
              <span class="chip chip-empty">Unlimited</span>
            {/if}
          </td>
        </tr>
        <tr>
          <th>Status</th>
          <td>
            {#if re.archived}
              <span class="chip chip-tag">Archived</span>
            {:else}
              <span class="chip chip-tag">Active</span>
            {/if}
          </td>
        </tr>
        <tr>
          <th>Tags</th>
          <td>
            {#if re.tags.length > 0}
              <div class="chip-list">
                {#each re.tags as tag (tag)}
                  <span class="chip chip-tag">{tag}</span>
                {/each}
              </div>
            {:else}
              <span class="chip chip-empty">No tags</span>
            {/if}
          </td>
        </tr>
      </tbody>
    </table>
    {#if re.archived}
      <button
        type="button"
        class="btn-primary form-submit"
        disabled={pending}
        onclick={unarchive}
      >
        Unarchive
      </button>
    {/if}
    <button
      type="button"
      class="btn-danger form-submit"
      disabled={pending}
      onclick={deleteRecurrentExpense}
    >
      Delete
    </button>
  {/if}
</section>

<script lang="ts">
  // Ports recurrent_expenses/_form.html. The legacy form shows one combined
  // error message rather than per-field highlighting (there is no per-field
  // markup in the template), so this does the same — see .form-error-text
  // below, matching common/_form_error.html.
  import { untrack } from "svelte";
  import { fetchCategories, type Category } from "../../lib/categories";
  import { centsToInputValue, inputValueToCents } from "../../lib/currency";
  import { joinTagNames, parseTagsInput } from "../../lib/tags";
  import type { RecurrentExpense, RecurrentExpenseRequestBody } from "./types";

  interface Props {
    initial?: RecurrentExpense;
    error?: string;
    pending?: boolean;
    submitLabel: string;
    onSubmit: (body: RecurrentExpenseRequestBody) => void;
  }

  let {
    initial,
    error = "",
    pending = false,
    submitLabel,
    onSubmit,
  }: Props = $props();

  // Seeds the editable fields once from `initial` and then diverges as the
  // user types — not a mirror of the prop, so the read is deliberately
  // untracked rather than a `$derived` that would overwrite in-progress edits
  // if the parent ever re-rendered with a new `initial`.
  const seed = untrack(() => initial);

  let categories = $state<Category[]>([]);
  let categoryId = $state(seed?.category_id ?? 0);
  let description = $state(seed?.description ?? "");
  // `<input type="number">`'s bind:value coerces to a live number once the
  // user edits it, no matter what type this is seeded with — see
  // inputValueToCents's note on why it accepts both.
  let amountInput = $state<number | string>(
    seed ? Number(centsToInputValue(seed.amount)) : "",
  );
  let tagsInput = $state(seed ? joinTagNames(seed.tags) : "");
  let period = $state(seed?.period ?? 1);
  let occurrenceLimit = $state(seed?.occurrence_limit ?? 0);
  let amountError = $state("");

  $effect(() => {
    let cancelled = false;

    fetchCategories()
      .then((result) => {
        if (cancelled) return;

        categories = result;
        if (!initial && categoryId === 0 && result.length > 0) {
          categoryId = result[0].id;
        }
      })
      .catch(() => {
        // The form still renders; the select is just empty until a retry.
      });

    return () => {
      cancelled = true;
    };
  });

  function handleSubmit(event: SubmitEvent): void {
    event.preventDefault();

    const amount = inputValueToCents(amountInput);
    if (amount === null) {
      amountError = "Amount must be a valid number.";
      return;
    }
    amountError = "";

    onSubmit({
      category_id: categoryId,
      description,
      amount,
      period,
      occurrence_limit: occurrenceLimit,
      tags: parseTagsInput(tagsInput),
    });
  }
</script>

{#if amountError || error}
  <p class="form-error-text">{amountError || error}</p>
{/if}

<form onsubmit={handleSubmit}>
  <label>
    Category
    <select bind:value={categoryId}>
      {#each categories as category (category.id)}
        <option value={category.id}>{category.name}</option>
      {/each}
    </select>
  </label>
  <label>
    Description
    <input type="text" bind:value={description} placeholder="New purchase..." />
  </label>
  <label>
    Amount
    <input type="number" min="0" step="0.01" bind:value={amountInput} />
  </label>
  <label>
    Tags
    <input
      type="text"
      placeholder="Semicolon separated"
      bind:value={tagsInput}
    />
  </label>
  <label>
    Period (months)
    <input type="number" min="1" step="1" bind:value={period} />
  </label>
  <label>
    Occurrence limit (0 = unlimited)
    <input type="number" min="0" step="1" bind:value={occurrenceLimit} />
  </label>
  <button type="submit" class="btn-primary form-submit" disabled={pending}>
    {pending ? "Saving..." : submitLabel}
  </button>
</form>

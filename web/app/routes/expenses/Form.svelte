<script lang="ts">
  // Ports expenses/_form.html (define "expense_form"). Same combined-error
  // convention as recurrent_expenses/Form.svelte: one .form-error-text string,
  // no per-field markup in the legacy template to port.
  import { untrack } from "svelte";
  import { fetchCategories, type Category } from "../../lib/categories";
  import { centsToInputValue, inputValueToCents } from "../../lib/currency";
  import {
    calendarDateToUnix,
    todayCalendarDate,
    unixToCalendarDate,
  } from "../../lib/dates";
  import { joinTagNames, parseTagsInput } from "../../lib/tags";
  import type { Expense, ExpenseRequestBody } from "./types";

  interface Props {
    initial?: Expense;
    error?: string;
    pending?: boolean;
    submitLabel: string;
    onSubmit: (body: ExpenseRequestBody) => void;
  }

  let {
    initial,
    error = "",
    pending = false,
    submitLabel,
    onSubmit,
  }: Props = $props();

  // Seeded once from `initial`, then diverges as the user types — see
  // recurrent_expenses/Form.svelte's identical comment on `untrack` here.
  const seed = untrack(() => initial);

  let categories = $state<Category[]>([]);
  let categoryId = $state(seed?.category_id ?? 0);
  let description = $state(seed?.description ?? "");
  // See recurrent_expenses/Form.svelte and lib/currency.ts's note on why this
  // accepts both types: bind:value on <input type="number"> hands back a live
  // number once the user edits it, whatever type it was seeded with.
  let amountInput = $state<number | string>(
    seed ? Number(centsToInputValue(seed.amount)) : "",
  );
  let tagsInput = $state(seed ? joinTagNames(seed.tags) : "");
  let dateInput = $state(
    seed ? unixToCalendarDate(seed.date) : todayCalendarDate(),
  );
  let formError = $state("");

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
      formError = "Amount must be a valid number.";
      return;
    }

    let date: number;
    try {
      date = calendarDateToUnix(dateInput);
    } catch {
      formError = "Date must be a valid calendar date.";
      return;
    }
    formError = "";

    onSubmit({
      category_id: categoryId,
      description,
      amount,
      date,
      tags: parseTagsInput(tagsInput),
    });
  }
</script>

{#if formError || error}
  <p class="form-error-text">{formError || error}</p>
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
    Date
    <input type="date" bind:value={dateInput} />
  </label>
  <button type="submit" class="btn-primary form-submit" disabled={pending}>
    {pending ? "Saving..." : submitLabel}
  </button>
</form>

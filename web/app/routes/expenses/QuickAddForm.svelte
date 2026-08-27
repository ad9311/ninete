<script lang="ts">
  // Ports expenses/_quick_form.html (define "quick_expense_form"). tz_offset
  // rides on the request explicitly rather than a Turbo hook injecting it —
  // §3.6 of docs/spa-migration.md, "Consumer 2": quick-add's relative dates
  // ("today", "yesterday") still need the client's own zone, and that has
  // nothing to do with the named date ranges §3.6 retires from the API.
  import { APIRequestError, post } from "../../lib/api";
  import { fetchCategories, type Category } from "../../lib/categories";
  import { navigate } from "../../router";
  import DateHelp from "../../components/DateHelp.svelte";
  import type { Expense, QuickExpenseRequestBody } from "./types";

  let quickInput = $state("");
  let categories = $state<Category[]>([]);
  let needsCategory = $state(false);
  let categoryId = $state(0);
  let error = $state("");
  let pending = $state(false);

  $effect(() => {
    let cancelled = false;

    fetchCategories()
      .then((result) => {
        if (!cancelled) categories = result;
      })
      .catch(() => {
        // The form still renders; the category picker just stays empty.
      });

    return () => {
      cancelled = true;
    };
  });

  async function handleSubmit(event: SubmitEvent): Promise<void> {
    event.preventDefault();
    pending = true;

    const body: QuickExpenseRequestBody = {
      quick_input: quickInput,
      tz_offset: new Date().getTimezoneOffset(),
    };
    if (needsCategory && categoryId > 0) {
      body.category_id = categoryId;
    }

    try {
      await post<Expense>("/expenses/quick", body);
      navigate("/expenses");
    } catch (err) {
      if (err instanceof APIRequestError && err.fields.category_id) {
        needsCategory = true;
        error =
          "First time for this description. Pick a category to remember it.";
      } else {
        error =
          err instanceof APIRequestError
            ? err.message
            : "Something went wrong.";
      }
    } finally {
      pending = false;
    }
  }
</script>

<form onsubmit={handleSubmit}>
  {#if error}
    <p class="form-error-text">{error}</p>
  {/if}
  <label>
    <span class="quick-label-row">
      Quick add
      <DateHelp label="Show accepted date formats">
        <p class="date-help-title">Date can be:</p>
        <ul>
          <li>
            <code>today</code>, <code>yesterday</code>, <code>tomorrow</code>
          </li>
          <li><code>next month</code> (1st of next month)</li>
          <li><code>12 July 2026</code> or <code>12 Jul 2026</code></li>
          <li><code>2026-07-12</code> (ISO)</li>
          <li><code>12/07/2026</code> (day/month/year)</li>
        </ul>
      </DateHelp>
    </span>
    <input
      type="text"
      bind:value={quickInput}
      placeholder="Description, amount, date, tags"
    />
  </label>
  <p class="quick-hint">
    Example: Uber, 3344.22, today, travel; work — tags are optional,
    semicolon-separated.
  </p>
  {#if needsCategory}
    <label>
      Category
      <select bind:value={categoryId}>
        <option value={0}>Select a category</option>
        {#each categories as category (category.id)}
          <option value={category.id}>{category.name}</option>
        {/each}
      </select>
    </label>
  {/if}
  <button type="submit" class="btn-primary form-submit" disabled={pending}>
    {pending ? "Saving..." : "Submit"}
  </button>
</form>

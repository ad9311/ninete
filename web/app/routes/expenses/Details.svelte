<script lang="ts">
  // The expense's details table, shared by its own page (Show.svelte) and the
  // list's quick-view dialog (DetailsDialog.svelte), so the two cannot drift.
  import LocalDate from "../../components/LocalDate.svelte";
  import { formatCurrency } from "../../lib/currency";
  import type { Expense } from "./types";

  let { expense }: { expense: Expense } = $props();
</script>

<div class="overflow-x-auto">
  <table class="data-table">
    <tbody>
      <tr>
        <th>Category</th>
        <td>{expense.category_name}</td>
      </tr>
      <tr>
        <th>Description</th>
        <td>{expense.description}</td>
      </tr>
      <tr>
        <th>Amount</th>
        <td class="amount font-semibold text-fg">
          {formatCurrency(expense.amount)}
        </td>
      </tr>
      <tr>
        <th>Billed month</th>
        <td><LocalDate value={expense.date} month /></td>
      </tr>
      <tr>
        <th>Created</th>
        <td><LocalDate value={expense.created_at} datetime /></td>
      </tr>
      <tr>
        <th>Tags</th>
        <td>
          {#if expense.tags.length > 0}
            <div class="flex flex-wrap gap-2">
              {#each expense.tags as tag (tag)}
                <span class="chip chip-tag">{tag}</span>
              {/each}
            </div>
          {:else}
            <span class="chip">No tags</span>
          {/if}
        </td>
      </tr>
      {#if expense.note}
        <tr>
          <th>Note</th>
          <!-- pre-line keeps the line breaks typed into the textarea;
            wrap-anywhere lets one long unbroken word wrap instead of
            widening the table past the card. -->
          <td class="wrap-anywhere whitespace-pre-line">{expense.note}</td>
        </tr>
      {/if}
    </tbody>
  </table>
</div>

<script lang="ts">
  // The expense list's quick view: a plain click on a row's description opens
  // this instead of navigating, so the list (its filters, page and scroll)
  // stays where it was. The expense's own page is unchanged and still the
  // target of the row's href, which is what a modifier click or a shared link
  // reaches.
  //
  // It renders the row the list already holds rather than fetching
  // /expenses/:id again: the list response is the same apiExpense shape, note
  // included. That also keeps the dialog off lib/pending.ts's spinner, which
  // would sit under the dialog's top layer and never show.
  //
  // A native <dialog> opened with showModal() gives the focus trap, Escape,
  // inert background and ::backdrop for free.
  import { Maximize2, SquarePen, X } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
  import { APIRequestError, del } from "../../lib/api";
  import { BASE_PATH } from "../../router";
  import Details from "./Details.svelte";
  import type { Expense } from "./types";

  interface Props {
    /** The expense to show; null keeps the dialog closed. */
    expense: Expense | null;
    onclose: () => void;
    /** Called after a successful delete, so the list can reload. */
    ondeleted: () => void;
  }

  let { expense, onclose, ondeleted }: Props = $props();

  let dialog = $state<HTMLDialogElement | null>(null);
  let error = $state("");

  $effect(() => {
    if (!dialog) return;
    if (expense && !dialog.open) {
      error = "";
      dialog.showModal();
    } else if (!expense && dialog.open) {
      dialog.close();
    }
  });

  // A click whose target is the <dialog> itself landed on the backdrop: the
  // card fills the dialog's box, so every click inside it targets a child.
  function onBackdropClick(event: MouseEvent): void {
    if (event.target === dialog) dialog?.close();
  }

  async function deleteExpense(): Promise<void> {
    if (!expense) return;
    if (!window.confirm("Delete this expense?")) return;

    try {
      await del(`/expenses/${expense.id}`);
      dialog?.close();
      ondeleted();
    } catch (err) {
      error =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
    }
  }
</script>

<!-- `close` fires for Escape, the close button and the backdrop alike, so it
  is the one place the parent hears about it. The backdrop click needs no key
  twin: Escape already closes a modal dialog. -->
<dialog
  bind:this={dialog}
  class="m-auto w-full max-w-lg bg-transparent backdrop:bg-backdrop max-md:max-w-[calc(100%-2rem)]"
  aria-label="Expense details"
  {onclose}
  onclick={onBackdropClick}
>
  {#if expense}
    <Card title="Expense" level={2} actionsLabel="Expense actions">
      {#snippet actions()}
        <CardAction
          icon={SquarePen}
          label="Edit expense"
          href={`${BASE_PATH}/expenses/${expense.id}/edit`}
        />
        <CardAction
          icon={Maximize2}
          label="Open expense page"
          href={`${BASE_PATH}/expenses/${expense.id}`}
        />
        <CardAction icon={X} label="Close" onclick={() => dialog?.close()} />
      {/snippet}
      {#if error}
        <p class="text-danger">{error}</p>
      {/if}
      <Details {expense} />
      <button
        type="button"
        class="btn btn-danger mt-3 justify-self-end"
        onclick={deleteExpense}
      >
        Delete
      </button>
    </Card>
  {/if}
</dialog>

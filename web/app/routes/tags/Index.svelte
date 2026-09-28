<script lang="ts">
  // Every tag the account has, with how many records carry it. Tags are
  // created as free text on the expense forms, so this is the one place they
  // are managed: retagged (renamed or merged) through POST /api/tags/retag,
  // and deleted through DELETE /api/tags/{id} or /api/tags/unused.
  //
  // A retag keeps its source tags, unused, on purpose (CLAUDE.md), which is
  // why unused tags are listed here and can be cleared in one go.
  import { ArrowLeft } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
  import { APIRequestError, del, get, post } from "../../lib/api";
  import { BASE_PATH } from "../../router";
  import type {
    DeleteUnusedTagsResponse,
    RetagResponse,
    TagListResponse,
    TagUsage,
  } from "./types";

  // logic.RetagParams caps "from" at 20 names. The server does not send the
  // number, so it is held here a second time; a mismatch surfaces as a 422.
  const RETAG_SOURCE_LIMIT = 20;

  let tags = $state<TagUsage[]>([]);
  let loaded = $state(false);
  let loadError = $state("");

  let selectedIDs = $state<number[]>([]);
  let target = $state("");
  let retagging = $state(false);
  let retagError = $state("");
  let retagFieldErrors = $state<Record<string, string>>({});
  let retagMessage = $state("");

  let deleting = $state(false);
  let deleteError = $state("");
  let deleteMessage = $state("");

  const unusedCount = $derived(tags.filter((t) => usage(t) === 0).length);
  const atSourceLimit = $derived(selectedIDs.length >= RETAG_SOURCE_LIMIT);
  const selectedNames = $derived(
    tags.filter((t) => selectedIDs.includes(t.id)).map((t) => t.name),
  );

  function usage(tag: TagUsage): number {
    return tag.expense_count + tag.recurrent_expense_count;
  }

  function plural(n: number, word: string): string {
    return `${n} ${word}${n === 1 ? "" : "s"}`;
  }

  function usageLabel(tag: TagUsage): string {
    if (usage(tag) === 0) return "Unused";

    const parts: string[] = [];
    if (tag.expense_count > 0) parts.push(plural(tag.expense_count, "expense"));
    if (tag.recurrent_expense_count > 0) {
      parts.push(plural(tag.recurrent_expense_count, "recurrent expense"));
    }

    return parts.join(" · ");
  }

  function errorMessage(err: unknown): string {
    return err instanceof APIRequestError
      ? err.message
      : "Something went wrong.";
  }

  async function load(): Promise<void> {
    const result = await get<TagListResponse>("/tags");
    tags = result.data;
    loadError = "";
    // A tag deleted or merged away must not stay selected by id.
    selectedIDs = selectedIDs.filter((id) => tags.some((t) => t.id === id));
    loaded = true;
  }

  $effect(() => {
    let cancelled = false;

    load().catch((err) => {
      if (cancelled) return;
      loadError = errorMessage(err);
    });

    return () => {
      cancelled = true;
    };
  });

  // Every action starts from a clean slate: an error or notice left by an
  // earlier action, of either kind, must not read as this one's outcome.
  function clearFeedback(): void {
    retagError = "";
    retagFieldErrors = {};
    retagMessage = "";
    deleteError = "";
    deleteMessage = "";
  }

  function toggle(id: number, checked: boolean): void {
    retagMessage = "";
    selectedIDs = checked
      ? [...selectedIDs, id]
      : selectedIDs.filter((selected) => selected !== id);
  }

  // Reloads after a change that already happened, so a failed reload is
  // reported as a load problem rather than as the change having failed.
  async function reload(): Promise<void> {
    await load().catch((err) => {
      loadError = errorMessage(err);
    });
  }

  async function retag(event: SubmitEvent): Promise<void> {
    event.preventDefault();
    retagging = true;
    clearFeedback();

    let done = false;
    try {
      const result = await post<RetagResponse>("/tags/retag", {
        from: selectedNames,
        to: target,
      });
      retagMessage =
        `Moved ${plural(result.retagged, "record")} onto "${result.tag.name}". ` +
        "The old tags are kept, now unused.";
      selectedIDs = [];
      target = "";
      done = true;
    } catch (err) {
      retagError = errorMessage(err);
      if (err instanceof APIRequestError) retagFieldErrors = err.fields;
    } finally {
      retagging = false;
    }

    if (done) await reload();
  }

  async function deleteTag(tag: TagUsage): Promise<void> {
    const question =
      usage(tag) === 0
        ? `Delete "${tag.name}"?`
        : `Delete "${tag.name}"? It comes off ${usageLabel(tag).replace(" · ", " and ")}, ` +
          "and out of the monthly report if it groups by it. The records themselves stay.";
    if (!confirm(question)) return;

    deleting = true;
    clearFeedback();

    let done = false;
    try {
      await del(`/tags/${tag.id}`);
      deleteMessage = `Deleted "${tag.name}".`;
      done = true;
    } catch (err) {
      deleteError = errorMessage(err);
    } finally {
      deleting = false;
    }

    if (done) await reload();
  }

  async function deleteUnused(): Promise<void> {
    if (!confirm(`Delete ${plural(unusedCount, "unused tag")}?`)) return;

    deleting = true;
    clearFeedback();

    let done = false;
    try {
      const result = await del<DeleteUnusedTagsResponse>("/tags/unused");
      deleteMessage = `Deleted ${plural(result.deleted, "unused tag")}.`;
      done = true;
    } catch (err) {
      deleteError = errorMessage(err);
    } finally {
      deleting = false;
    }

    if (done) await reload();
  }
</script>

<Card title="Tags" actionsLabel="Tag actions">
  {#snippet actions()}
    <CardAction
      icon={ArrowLeft}
      label="Back to account"
      href={`${BASE_PATH}/account`}
    />
  {/snippet}
  <p class="text-sm text-muted">
    Tags are created by adding them to an expense. Pick tags below and give a
    name to move every expense and recurrent expense carrying them onto that
    tag: a new name renames, an existing one merges. The picked tags are kept,
    unused, until you delete them.
  </p>

  {#if loadError}
    <p class="text-danger">{loadError}</p>
  {/if}

  <form onsubmit={retag} class="grid max-w-form gap-3">
    <p class="text-sm text-muted">
      {#if selectedNames.length === 0}
        No tags picked.
      {:else}
        Retagging {selectedNames.join(", ")}.
      {/if}
    </p>
    <label>
      Retag as
      <input
        type="text"
        maxlength="20"
        list="tag-names"
        placeholder="transport"
        bind:value={target}
        aria-invalid={retagFieldErrors.to ? "true" : undefined}
      />
    </label>
    <datalist id="tag-names">
      {#each tags as tag (tag.id)}
        <option value={tag.name}></option>
      {/each}
    </datalist>

    {#if retagError}
      <p class="text-danger">{retagError}</p>
    {/if}
    {#if retagMessage}
      <p class="text-sm text-muted">{retagMessage}</p>
    {/if}

    <button
      type="submit"
      class="btn btn-primary justify-self-end"
      disabled={retagging ||
        !loaded ||
        selectedIDs.length === 0 ||
        target.trim() === ""}
    >
      {retagging ? "Retagging..." : "Retag"}
    </button>
  </form>
</Card>

{#if loaded}
  <Card title="Your tags" level={2}>
    {#if deleteError}
      <p class="text-danger">{deleteError}</p>
    {/if}
    {#if deleteMessage}
      <p class="text-sm text-muted">{deleteMessage}</p>
    {/if}
    {#if tags.length === 0}
      <p class="text-sm text-muted">
        You have no tags yet. Add some to your expenses first.
      </p>
    {:else}
      <ul class="grid gap-2">
        {#each tags as tag (tag.id)}
          {@const checked = selectedIDs.includes(tag.id)}
          <li
            class="flex flex-wrap items-center justify-between gap-3 rounded-xs border border-line px-4 py-3"
          >
            <label class="inline-flex items-center gap-2">
              <input
                type="checkbox"
                {checked}
                disabled={!checked && atSourceLimit}
                onchange={(event) => {
                  toggle(tag.id, event.currentTarget.checked);
                }}
              />
              <span class="grid gap-1">
                <span class="font-bold">{tag.name}</span>
                <span class="text-sm text-muted">{usageLabel(tag)}</span>
              </span>
            </label>
            <button
              type="button"
              class="btn btn-danger"
              disabled={deleting}
              aria-label={`Delete ${tag.name}`}
              onclick={() => void deleteTag(tag)}
            >
              Delete
            </button>
          </li>
        {/each}
      </ul>
      <button
        type="button"
        class="btn btn-danger justify-self-end"
        disabled={deleting || unusedCount === 0}
        onclick={() => void deleteUnused()}
      >
        Delete unused ({unusedCount})
      </button>
    {/if}
  </Card>
{/if}

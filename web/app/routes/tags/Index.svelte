<script lang="ts">
  // Every tag the account has, with how many records carry it. Tags are
  // created as free text on the expense forms, so this is the one place they
  // are managed, and every action sits on the tags it affects:
  //
  // - Rename, on a row, retags that one tag onto a new name. Typing a name that
  //   already exists merges into it, and the row says so before saving.
  // - Merge appears once two or more rows are ticked, and retags all of them
  //   onto one name: a new one, one of their own, or any other existing tag.
  // - Delete, on a row, removes one tag; "Delete unused" clears the tags a
  //   retag leaves behind.
  //
  // Rename and merge are both POST /api/tags/retag, which keeps its source
  // tags, unused, on purpose (CLAUDE.md) — hence the unused count up top.
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
  const MERGE_SOURCE_LIMIT = 20;

  let tags = $state<TagUsage[]>([]);
  let loaded = $state(false);
  let loadError = $state("");

  // One message area for every action, so an outcome always shows in the
  // same place whichever control produced it.
  let actionError = $state("");
  let notice = $state("");
  let busy = $state(false);

  let editingID = $state<number | null>(null);
  let newName = $state("");

  let selectedIDs = $state<number[]>([]);
  let mergeTarget = $state("");

  const unusedCount = $derived(tags.filter((t) => usage(t) === 0).length);
  const selectedNames = $derived(
    tags.filter((t) => selectedIDs.includes(t.id)).map((t) => t.name),
  );
  const atMergeLimit = $derived(selectedIDs.length >= MERGE_SOURCE_LIMIT);

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

  // The server stores names trimmed and lowercased, so that is how two names
  // compare here.
  function normalize(name: string): string {
    return name.trim().toLowerCase();
  }

  function existingTag(name: string, exceptID: number): TagUsage | undefined {
    const key = normalize(name);

    return tags.find((t) => t.id !== exceptID && t.name === key);
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

  // Every action starts from a clean slate: a message left by an earlier
  // action must not read as this one's outcome.
  function clearFeedback(): void {
    actionError = "";
    notice = "";
  }

  // Runs one change and reloads after it. The reload sits outside the
  // change's try: the change already happened, so a failed reload is a load
  // problem, not a failed change to be retried.
  async function run(change: () => Promise<string>): Promise<boolean> {
    clearFeedback();
    busy = true;

    let done = false;
    try {
      notice = await change();
      done = true;
    } catch (err) {
      actionError = errorMessage(err);
    } finally {
      busy = false;
    }

    if (done) {
      await load().catch((err) => {
        loadError = errorMessage(err);
      });
    }

    return done;
  }

  function retagNotice(result: RetagResponse): string {
    return (
      `Moved ${plural(result.retagged, "record")} onto "${result.tag.name}". ` +
      "The old tags are kept, now unused."
    );
  }

  function startRename(tag: TagUsage): void {
    clearFeedback();
    editingID = tag.id;
    newName = tag.name;
  }

  function cancelRename(): void {
    editingID = null;
    newName = "";
  }

  async function saveRename(event: SubmitEvent, tag: TagUsage): Promise<void> {
    event.preventDefault();

    const renamed = await run(async () =>
      retagNotice(
        await post<RetagResponse>("/tags/retag", {
          from: [tag.name],
          to: newName,
        }),
      ),
    );
    if (renamed) cancelRename();
  }

  function toggle(id: number, checked: boolean): void {
    selectedIDs = checked
      ? [...selectedIDs, id]
      : selectedIDs.filter((selected) => selected !== id);
  }

  async function merge(event: SubmitEvent): Promise<void> {
    event.preventDefault();

    // Merging into one of the ticked tags keeps that tag: it is the target,
    // not a source, and the server refuses a target that is also a source.
    const target = normalize(mergeTarget);
    const from = selectedNames.filter((name) => name !== target);

    const merged = await run(async () =>
      retagNotice(
        await post<RetagResponse>("/tags/retag", { from, to: mergeTarget }),
      ),
    );
    if (merged) {
      selectedIDs = [];
      mergeTarget = "";
    }
  }

  async function deleteTag(tag: TagUsage): Promise<void> {
    const question =
      usage(tag) === 0
        ? `Delete "${tag.name}"?`
        : `Delete "${tag.name}"? It comes off ${usageLabel(tag).replace(" · ", " and ")}, ` +
          "and out of the monthly report if it groups by it. The records themselves stay.";
    if (!confirm(question)) return;

    await run(async () => {
      await del(`/tags/${tag.id}`);

      return `Deleted "${tag.name}".`;
    });
  }

  async function deleteUnused(): Promise<void> {
    if (!confirm(`Delete ${plural(unusedCount, "unused tag")}?`)) return;

    await run(async () => {
      const result = await del<DeleteUnusedTagsResponse>("/tags/unused");

      return `Deleted ${plural(result.deleted, "unused tag")}.`;
    });
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
    Tags are created by adding them to an expense. Rename a tag to change it
    everywhere, or tick two or more to merge them into one. Renamed and merged
    tags are kept, unused, until you delete them.
  </p>

  {#if loadError}
    <p class="text-danger">{loadError}</p>
  {/if}
  {#if actionError}
    <p class="text-danger">{actionError}</p>
  {/if}
  {#if notice}
    <p class="text-sm text-muted">{notice}</p>
  {/if}

  {#if loaded}
    {#if tags.length === 0}
      <p class="text-sm text-muted">
        You have no tags yet. Add some to your expenses first.
      </p>
    {:else}
      <button
        type="button"
        class="btn btn-danger justify-self-end"
        disabled={busy || unusedCount === 0}
        onclick={() => void deleteUnused()}
      >
        Delete unused ({unusedCount})
      </button>

      <ul class="grid gap-2">
        {#each tags as tag (tag.id)}
          {@const checked = selectedIDs.includes(tag.id)}
          {@const clash =
            editingID === tag.id ? existingTag(newName, tag.id) : undefined}
          <li class="grid gap-2 rounded-xs border border-line px-4 py-3">
            {#if editingID === tag.id}
              <form
                class="flex flex-wrap items-center gap-3"
                onsubmit={(event) => void saveRename(event, tag)}
              >
                <input
                  type="text"
                  class="min-w-0 flex-1"
                  maxlength="20"
                  aria-label={`New name for ${tag.name}`}
                  bind:value={newName}
                  onkeydown={(event) => {
                    if (event.key === "Escape") cancelRename();
                  }}
                />
                <span class="text-sm text-muted">{usageLabel(tag)}</span>
                <span class="flex gap-2">
                  <button
                    type="submit"
                    class="btn btn-primary"
                    disabled={busy ||
                      normalize(newName) === "" ||
                      normalize(newName) === tag.name}
                  >
                    {clash ? "Merge" : "Save"}
                  </button>
                  <button
                    type="button"
                    class="btn btn-neutral"
                    onclick={cancelRename}
                  >
                    Cancel
                  </button>
                </span>
              </form>
              {#if clash}
                <p class="text-sm text-muted">
                  "{clash.name}" already exists — saving merges "{tag.name}"
                  into it.
                </p>
              {/if}
            {:else}
              <div class="flex flex-wrap items-center justify-between gap-3">
                <label class="inline-flex items-center gap-2">
                  <input
                    type="checkbox"
                    {checked}
                    disabled={!checked && atMergeLimit}
                    onchange={(event) => {
                      toggle(tag.id, event.currentTarget.checked);
                    }}
                  />
                  <span class="grid gap-1">
                    <span class="font-bold">{tag.name}</span>
                    <span class="text-sm text-muted">{usageLabel(tag)}</span>
                  </span>
                </label>
                <span class="flex gap-2">
                  <button
                    type="button"
                    class="btn btn-neutral"
                    disabled={busy}
                    aria-label={`Rename ${tag.name}`}
                    onclick={() => startRename(tag)}
                  >
                    Rename
                  </button>
                  <button
                    type="button"
                    class="btn btn-danger"
                    disabled={busy}
                    aria-label={`Delete ${tag.name}`}
                    onclick={() => void deleteTag(tag)}
                  >
                    Delete
                  </button>
                </span>
              </div>
            {/if}
          </li>
        {/each}
      </ul>

      {#if selectedIDs.length >= 2}
        <!-- Sticky so the bar stays in reach while ticking down a long list. -->
        <form
          class="sticky bottom-4 flex flex-wrap items-center gap-3 rounded-xs border border-line-strong bg-surface p-3 shadow-card"
          onsubmit={(event) => void merge(event)}
        >
          <label class="flex min-w-0 flex-1 flex-wrap items-center gap-2">
            <span class="font-bold">
              Merge {selectedIDs.length} tags into
            </span>
            <input
              type="text"
              class="min-w-0 flex-1"
              maxlength="20"
              list="tag-names"
              placeholder={selectedNames[0]}
              bind:value={mergeTarget}
            />
          </label>
          <datalist id="tag-names">
            {#each tags as tag (tag.id)}
              <option value={tag.name}></option>
            {/each}
          </datalist>
          <button
            type="submit"
            class="btn btn-primary"
            disabled={busy || normalize(mergeTarget) === ""}
          >
            Merge
          </button>
        </form>
      {/if}
    {/if}
  {/if}
</Card>

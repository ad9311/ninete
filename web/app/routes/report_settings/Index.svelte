<script lang="ts">
  // The monthly report's settings (docs/monthly-report.md, Phase 1). Nothing
  // consumes these yet: the PDF lands in Phase 2 and the scheduled email in
  // Phase 3.
  //
  // The tag list rides along on GET /api/report-settings rather than coming
  // from a listing endpoint of its own — tags are created as free text on the
  // expense forms and have never had one.
  import { Download } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
  import { APIRequestError, get, put } from "../../lib/api";
  import { lastCalendarMonth } from "../../lib/dates";
  import { browserTimezone, timezoneOptions } from "./timezones";
  import type { ReportSettingsResponse, ReportTag } from "./types";

  let tags = $state<ReportTag[]>([]);
  let selectedTagIDs = $state<number[]>([]);
  let timezone = $state("");
  let tagLimit = $state(0);
  // Nothing may be submitted before the GET lands. PUT replaces the tag list
  // wholesale, so a failed load would otherwise leave an empty, enabled form
  // whose Save wipes every grouping tag the user had — the page cannot tell
  // "no tags selected" from "we never found out".
  let loaded = $state(false);
  let loadError = $state("");
  let saveError = $state("");
  let saved = $state(false);
  let saving = $state(false);
  // The month the download asks for, defaulting to the month that just ended
  // — the period the scheduled report will send.
  let month = $state(lastCalendarMonth());

  const zones = $derived(timezoneOptions(timezone));
  const atTagLimit = $derived(
    tagLimit > 0 && selectedTagIDs.length >= tagLimit,
  );

  $effect(() => {
    let cancelled = false;

    get<ReportSettingsResponse>("/report-settings")
      .then((result) => {
        if (cancelled) return;
        tags = result.tags;
        selectedTagIDs = result.selected_tag_ids;
        // An unsaved account gets the browser's zone as a starting point;
        // the server's own fallback is UTC, which would otherwise look like
        // a choice the user had made.
        timezone = result.configured ? result.timezone : browserTimezone();
        tagLimit = result.tag_limit;
        loadError = "";
        loaded = true;
      })
      .catch((err) => {
        if (cancelled) return;
        loadError =
          err instanceof APIRequestError
            ? err.message
            : "Something went wrong.";
      });

    return () => {
      cancelled = true;
    };
  });

  function toggleTag(id: number, checked: boolean): void {
    saved = false;
    selectedTagIDs = checked
      ? [...selectedTagIDs, id]
      : selectedTagIDs.filter((tagID) => tagID !== id);
  }

  async function saveSettings(event: SubmitEvent): Promise<void> {
    event.preventDefault();
    saving = true;
    saveError = "";
    saved = false;

    try {
      await put("/report-settings", {
        timezone,
        tag_ids: selectedTagIDs,
      });
      saved = true;
    } catch (err) {
      saveError =
        err instanceof APIRequestError ? err.message : "Something went wrong.";
    } finally {
      saving = false;
    }
  }
</script>

<Card title="Monthly report" actionsLabel="Report actions">
  {#snippet actions()}
    <!-- A plain anchor to the page-chain endpoint, outside lib/api.ts, for the
         reason routes/exports/Index.svelte spells out: a fetch response cannot
         reach the browser's save flow, and an expired session must be able to
         redirect. rel="external" keeps router.ts's onLinkClick off it; it is
         deliberately not `download`, which would save a redirect to the login
         page as a file. The path is handlers.ReportExpensesPath, held here a
         second time because a bundle cannot import a Go constant. -->
    <CardAction
      icon={Download}
      label="Download report"
      href={`/reports/expenses.pdf?month=${month}`}
      rel="external"
    />
  {/snippet}
  <p class="text-sm text-muted">
    The monthly report covers one billed month — every expense whose date falls
    in it, whatever day you actually made the purchase.
  </p>

  <label class="grid max-w-form gap-1">
    <span class="font-bold">Report month</span>
    <span class="text-sm text-muted">
      Which month to download. Defaults to the one that just ended.
    </span>
    <!-- Clearing the box snaps back to the default rather than leaving it
         empty. An empty value sends "?month=", which the server reads as
         absent and answers with last month — a blank picker handing back a
         specific month. `required` cannot help: this input is outside a form,
         so nothing validates it, and CardAction's `disabled` only reaches its
         button branch, not the anchor this link renders as. -->
    <input
      type="month"
      bind:value={month}
      onchange={() => {
        if (month === "") month = lastCalendarMonth();
      }}
    />
  </label>

  {#if loadError}
    <p class="text-danger">{loadError}</p>
  {/if}

  <form onsubmit={saveSettings} class="grid max-w-form gap-6">
    <fieldset class="grid gap-2 border-0 p-0">
      <legend class="mb-1 font-bold">Group by tags</legend>
      <p class="text-sm text-muted">
        The report prints a section per tag you pick here, ordered by total.
        <strong class="font-bold text-fg">
          An expense appears in one section only
        </strong>
        — the one for whichever of these tags you added to it first. Expenses carrying
        none of them go to a final "Untagged" section, and a tag with nothing in the
        month is left out. Pick none and the report is one flat list.
      </p>

      {#if !loaded}
        <!-- Deliberately silent until the load settles: "You have no tags yet"
             would otherwise be shown for a request that simply failed. -->
      {:else if tags.length === 0}
        <p class="text-sm text-muted">
          You have no tags yet. Add some to your expenses first.
        </p>
      {:else}
        <p class="text-sm text-muted">
          {selectedTagIDs.length} of {tagLimit} selected
        </p>
        <div class="grid gap-2 sm:grid-cols-2">
          {#each tags as tag (tag.id)}
            {@const checked = selectedTagIDs.includes(tag.id)}
            <label class="inline-flex items-center gap-2">
              <input
                type="checkbox"
                {checked}
                disabled={!checked && atTagLimit}
                onchange={(event) => {
                  toggleTag(tag.id, event.currentTarget.checked);
                }}
              />
              {tag.name}
            </label>
          {/each}
        </div>
      {/if}
    </fieldset>

    <label class="grid gap-1">
      <span class="font-bold">Time zone</span>
      <span class="text-sm text-muted">
        Decides which month the scheduled report calls "last month" when it
        runs. A report you download yourself uses the month you pick instead.
      </span>
      <select
        bind:value={timezone}
        disabled={!loaded}
        onchange={() => {
          saved = false;
        }}
      >
        {#each zones as zone (zone)}
          <option value={zone}>{zone}</option>
        {/each}
      </select>
    </label>

    {#if saveError}
      <p class="text-danger">{saveError}</p>
    {/if}
    {#if saved}
      <p class="text-sm text-muted">Settings saved.</p>
    {/if}

    <button
      type="submit"
      name="save_report_settings"
      class="btn btn-primary justify-self-end"
      disabled={saving || !loaded}
    >
      {saving ? "Saving..." : "Save settings"}
    </button>
  </form>
</Card>

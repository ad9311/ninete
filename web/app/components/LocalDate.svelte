<script lang="ts">
  // Ports localDateController.ts's two display modes (§3.6 of
  // docs/spa-migration.md), plus the billed-month mode added when the expense
  // billed date stopped showing its day: a calendar date formatted with UTC
  // getters as a full date or as a month, or an instant formatted in the
  // viewer's zone with the full local timestamp as its title tooltip. Shared by
  // every dated resource (expenses first, in Phase 3) — §3.9 rule 2 reserves
  // components/ for exactly this.
  import {
    formatDate,
    formatDateTime,
    formatDateUTC,
    formatMonthUTC,
    type UnixOrDate,
  } from "../lib/dates";

  interface Props {
    value: UnixOrDate;
    /** True for an instant (created_at); omitted for a calendar date. */
    datetime?: boolean;
    /**
     * True to render a calendar date as its month alone (the billed date).
     * Ignored when `datetime` is set — an instant has no month mode.
     */
    month?: boolean;
  }

  let { value, datetime = false, month = false }: Props = $props();
</script>

{#if datetime}
  <span title={formatDateTime(value)}>{formatDate(value)}</span>
{:else if month}
  <span>{formatMonthUTC(value)}</span>
{:else}
  <span>{formatDateUTC(value)}</span>
{/if}

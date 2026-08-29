<script lang="ts">
  // Ports localDateController.ts's two display modes (§3.6 of
  // docs/spa-migration.md): a calendar date formatted with UTC getters, or an
  // instant formatted in the viewer's zone with the full local timestamp as
  // its title tooltip. Shared by every dated resource (expenses first, in
  // Phase 3) — §3.9 rule 2 reserves components/ for exactly this.
  import {
    formatDate,
    formatDateTime,
    formatDateUTC,
    type UnixOrDate,
  } from "../lib/dates";

  interface Props {
    value: UnixOrDate;
    /** True for an instant (created_at); omitted for a calendar date. */
    datetime?: boolean;
  }

  let { value, datetime = false }: Props = $props();
</script>

{#if datetime}
  <span title={formatDateTime(value)}>{formatDate(value)}</span>
{:else}
  <span>{formatDateUTC(value)}</span>
{/if}

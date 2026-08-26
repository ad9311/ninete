<script lang="ts">
  // Router-level pending flag, replacing Turbo's progress bar (§3.7 of
  // docs/spa-migration.md). Shares layout.css's spinner styling with the
  // Turbo one it replaces rather than duplicating it — see .route-progress-bar
  // there.
  let { visible = false }: { visible?: boolean } = $props();
</script>

{#if visible}
  <div class="route-progress-bar" aria-hidden="true"></div>
{/if}

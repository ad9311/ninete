<script lang="ts">
  // Router-level pending flag, replacing Turbo's progress bar (§3.7 of
  // docs/spa-migration.md). Styling lives in layout.css under
  // .route-progress-bar.
  let { visible = false }: { visible?: boolean } = $props();
</script>

{#if visible}
  <div class="route-progress-bar" aria-hidden="true"></div>
{/if}

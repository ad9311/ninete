<script lang="ts">
  // Renders the router-level pending flag (§3.7 of docs/spa-migration.md).
  // Styling lives in web/app/app.css under .route-progress-bar: the ring is a
  // ::before with an animation, which is not something a utility can carry.
  let { visible = false }: { visible?: boolean } = $props();
</script>

{#if visible}
  <div class="route-progress-bar" aria-hidden="true"></div>
{/if}

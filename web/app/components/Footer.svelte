<script lang="ts">
  // The shell prints the build stamp into a <meta> tag (see App.svelte's
  // document); the footer reads it rather than fetching it, since it never
  // changes within a page load.
  const version =
    document.querySelector<HTMLMetaElement>('meta[name="version"]')?.content ??
    "";
</script>

<footer class="mt-8 flex justify-end border-t border-line pt-3">
  <span class="text-sm text-muted tabular-nums">{version}</span>
</footer>

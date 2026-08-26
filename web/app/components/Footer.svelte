<script lang="ts">
  // The shell prints the build stamp into a <meta> tag (see App.svelte's
  // document); the footer reads it rather than fetching it, since it never
  // changes within a page load.
  const version =
    document.querySelector<HTMLMetaElement>('meta[name="version"]')?.content ??
    "";
</script>

<footer class="site-footer">
  <span class="site-footer-version">{version}</span>
</footer>

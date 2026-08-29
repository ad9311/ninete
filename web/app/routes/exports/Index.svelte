<script lang="ts">
  // Ports exports/index.html. The download link stays a plain anchor to a
  // real endpoint (§5, Phase 5 note in docs/spa-migration.md: "Export links
  // stay plain anchors") — going through lib/api.ts would fetch the file into
  // memory as a Response with no way to hand it to the browser's save flow,
  // for no benefit, since a GET needs no CSRF token anyway. The `download`
  // attribute is what keeps router.ts's onLinkClick off it — since Phase 7
  // flattened BASE_PATH to "", its same-origin check matches every internal
  // path, /api/* included, so dropping `download` would have the client
  // router swallow this link and render "Not found."
  import { Download } from "lucide";
  import Icon from "../../components/Icon.svelte";
</script>

<section class="card" aria-labelledby="exports-card-title">
  <header class="card-header">
    <h1 id="exports-card-title" class="card-title">Exports</h1>
    <nav class="card-actions" aria-label="Export actions">
      <a
        href="/api/exports/expenses.json"
        class="card-action-link"
        aria-label="Download expenses"
        title="Download expenses"
        download
      >
        <Icon icon={Download} class="card-action-icon" />
      </a>
    </nav>
  </header>
  <p>
    Download all your expenses as a JSON file. Includes category and tag names.
    All dates are Unix timestamps in UTC.
  </p>
</section>

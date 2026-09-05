<script lang="ts">
  // Ports exports/index.html. The download link stays a plain anchor to a
  // real endpoint (§5, Phase 5 note in docs/spa-migration.md: "Export links
  // stay plain anchors") — going through lib/api.ts would fetch the file into
  // memory as a Response with no way to hand it to the browser's save flow,
  // for no benefit, since a GET needs no CSRF token anyway.
  //
  // Two things keep that anchor working, and both are easy to undo:
  //
  // The href is "/exports/...", not "/api/exports/...". The endpoint sits on
  // the page chain so an expired session answers with a redirect to the login
  // page. Under /api it answered 401 with no Location, and a navigation has
  // nothing to follow — the browser saved the JSON error envelope as
  // expenses.json, with no sign-in prompt and no visible failure. This path is
  // handlers.ExportExpensesPath; a bundle cannot import a Go constant, so the
  // two literals have to change together.
  //
  // rel="external" keeps router.ts's onLinkClick off it — since Phase 7
  // flattened BASE_PATH to "", its same-origin check matches every internal
  // path, so without an opt-out the client router would swallow this link and
  // render "Not found." The `download` attribute used to do that job, but it
  // also makes the browser save whatever comes back, a redirect included: with
  // it, an expired session saved the login page as a file. The server's
  // Content-Disposition: attachment is what starts the save now, so the
  // browser downloads only when the response really is the export.
  import { Download } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
</script>

<Card title="Exports" actionsLabel="Export actions">
  {#snippet actions()}
    <CardAction
      icon={Download}
      label="Download expenses"
      href="/exports/expenses.json"
      rel="external"
    />
  {/snippet}
  <p class="text-muted">
    Download all your expenses as a JSON file. Includes category and tag names.
    All dates are Unix timestamps in UTC.
  </p>
</Card>

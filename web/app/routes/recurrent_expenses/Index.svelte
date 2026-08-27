<script lang="ts">
  import { Archive, Plus } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import { BASE_PATH } from "../../router";
  import List from "./List.svelte";

  let { search = "" }: { search?: string } = $props();
</script>

<section class="card" aria-labelledby="recurrent-expenses-card-title">
  <header class="card-header">
    <h1 id="recurrent-expenses-card-title" class="card-title">
      Recurrent expenses
    </h1>
    <nav class="card-actions" aria-label="Recurrent expense actions">
      <a
        href={`${BASE_PATH}/recurrent-expenses/new`}
        class="card-action-link"
        aria-label="New recurrent expense"
        title="New recurrent expense"
      >
        <Icon icon={Plus} class="card-action-icon" />
      </a>
      <a
        href={`${BASE_PATH}/recurrent-expenses/archived`}
        class="card-action-link"
        aria-label="Archived recurrent expenses"
        title="Archived recurrent expenses"
      >
        <Icon icon={Archive} class="card-action-icon" />
      </a>
    </nav>
  </header>
  <List archived={false} basePath="/recurrent-expenses" {search} />
</section>

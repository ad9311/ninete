<script lang="ts">
  import { Repeat } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import { BASE_PATH } from "../../router";
  import List from "./List.svelte";

  let { search = "" }: { search?: string } = $props();
</script>

<section class="card" aria-labelledby="archived-recurrent-expenses-card-title">
  <header class="card-header">
    <h1 id="archived-recurrent-expenses-card-title" class="card-title">
      Archived recurrent expenses
    </h1>
    <nav class="card-actions" aria-label="Archived recurrent expense actions">
      <a
        href={`${BASE_PATH}/recurrent-expenses`}
        class="card-action-link"
        aria-label="Recurrent expenses"
        title="Recurrent expenses"
      >
        <Icon icon={Repeat} class="card-action-icon" />
      </a>
    </nav>
  </header>
  <List archived={true} basePath="/recurrent-expenses/archived" {search} />
</section>

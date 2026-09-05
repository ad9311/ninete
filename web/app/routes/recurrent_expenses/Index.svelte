<script lang="ts">
  import { Archive, Plus } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
  import { BASE_PATH } from "../../router";
  import List from "./List.svelte";

  let { search = "" }: { search?: string } = $props();
</script>

<Card title="Recurrent expenses" actionsLabel="Recurrent expense actions">
  {#snippet actions()}
    <CardAction
      icon={Plus}
      label="New recurrent expense"
      href={`${BASE_PATH}/recurrent-expenses/new`}
    />
    <CardAction
      icon={Archive}
      label="Archived recurrent expenses"
      href={`${BASE_PATH}/recurrent-expenses/archived`}
    />
  {/snippet}
  <List archived={false} basePath="/recurrent-expenses" {search} />
</Card>

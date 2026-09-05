<script lang="ts">
  import { Repeat } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
  import { BASE_PATH } from "../../router";
  import List from "./List.svelte";

  let { search = "" }: { search?: string } = $props();
</script>

<Card
  title="Archived recurrent expenses"
  actionsLabel="Archived recurrent expense actions"
>
  {#snippet actions()}
    <CardAction
      icon={Repeat}
      label="Recurrent expenses"
      href={`${BASE_PATH}/recurrent-expenses`}
    />
  {/snippet}
  <List archived={true} basePath="/recurrent-expenses/archived" {search} />
</Card>

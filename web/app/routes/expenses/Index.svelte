<script lang="ts">
  import { ChartColumn, Plus, Repeat, Target } from "lucide";
  import Card from "../../components/Card.svelte";
  import CardAction from "../../components/CardAction.svelte";
  import { BASE_PATH } from "../../router";
  import List from "./List.svelte";

  interface Props {
    search?: string;
  }

  let { search = "" }: Props = $props();
</script>

<Card title="Expenses" actionsLabel="Expense actions">
  {#snippet actions()}
    <CardAction
      icon={Plus}
      label="New expense"
      href={`${BASE_PATH}/expenses/new`}
    />
    <CardAction
      icon={Repeat}
      label="Recurrent expenses"
      href={`${BASE_PATH}/recurrent-expenses`}
    />
    <CardAction
      icon={ChartColumn}
      label="Stats"
      href={`${BASE_PATH}/expenses/stats`}
    />
    <CardAction
      icon={Target}
      label="Budgets"
      href={`${BASE_PATH}/expenses/budgets`}
    />
  {/snippet}
  <List {search} />
</Card>

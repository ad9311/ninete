<script lang="ts">
  import { ChartColumn, Plus, Repeat, Target } from "lucide";
  import Icon from "../../components/Icon.svelte";
  import { BASE_PATH } from "../../router";
  import List from "./List.svelte";

  interface Props {
    search?: string;
  }

  let { search = "" }: Props = $props();
</script>

<section class="card" aria-labelledby="expenses-card-title">
  <header class="card-header">
    <h1 id="expenses-card-title" class="card-title">Expenses</h1>
    <nav class="card-actions" aria-label="Expense actions">
      <a
        href={`${BASE_PATH}/expenses/new`}
        class="card-action-link"
        aria-label="New expense"
        title="New expense"
      >
        <Icon icon={Plus} class="card-action-icon" />
      </a>
      <a
        href={`${BASE_PATH}/recurrent-expenses`}
        class="card-action-link"
        aria-label="Recurrent expenses"
        title="Recurrent expenses"
      >
        <Icon icon={Repeat} class="card-action-icon" />
      </a>
      <a
        href={`${BASE_PATH}/expenses/stats`}
        class="card-action-link"
        aria-label="Stats"
        title="Stats"
      >
        <Icon icon={ChartColumn} class="card-action-icon" />
      </a>
      <a
        href={`${BASE_PATH}/expenses/budgets`}
        class="card-action-link"
        aria-label="Budgets"
        title="Budgets"
      >
        <Icon icon={Target} class="card-action-icon" />
      </a>
    </nav>
  </header>
  <List {search} />
</section>

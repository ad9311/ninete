<script lang="ts">
	import type { Ledger } from '$lib/server/db/schema';
	import Card from '$lib/components/ui/Card.svelte';
	import LedgerItem from './LedgerItem.svelte';

	const { ledgers }: { ledgers: Ledger[] } = $props();
</script>

<Card>
	{#snippet header()}
		<h3 class="card-title">Loans</h3>
	{/snippet}
	<ul class="flex flex-col gap-1">
		{#each ledgers as ledger (ledger.id)}
			<li>
				<LedgerItem {ledger} />
			</li>
		{/each}
	</ul>
</Card>

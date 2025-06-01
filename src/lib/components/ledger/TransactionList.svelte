<script lang="ts">
	import type { Transaction } from '$lib/server/db/schema';
	import Amount from '$lib/components/ledger/Amount.svelte';
	import Card from '$lib/components/ui/Card.svelte';

	export let transactions: Transaction[];
	export let type: 'credit' | 'debit';
	export let ledgerType: 'budget' | 'payable/receivable';

	const ledgerPath = ledgerType === 'budget' ? 'budgets' : 'accounts';
	const title = type === 'credit' ? 'Credits' : 'Debits';
</script>

<Card>
	{#snippet header()}
		<h3 class="card-title">{title}</h3>
	{/snippet}
	{#if transactions.length > 0}
		{#each transactions as transaction (transaction.id)}
			<p>{transaction.description}</p>
			<p><Amount value={transaction.amount} {type} /></p>
			<a href={`/ledgers/${ledgerPath}/${transaction.ledgerId}/transactions/${transaction.id}`}
				>View</a
			>
		{/each}
	{:else}
		<p class="text-muted text-center text-sm">No transactions</p>
	{/if}
</Card>

<script lang="ts">
	import type { Transaction } from '$lib/server/db/schema';
	import Amount from '$lib/components/Amount.svelte';

	export let transactions: Transaction[];
	export let type: 'credit' | 'debit';
	export let ledgerType: 'budget' | 'payable/receivable';

	const ledgerPath = ledgerType === 'budget' ? 'budgets' : 'accounts';
</script>

<div>
	{#each transactions as transaction (transaction.id)}
		<p>{transaction.description}</p>
		<p><Amount value={transaction.amount} {type} /></p>
		<a href={`/ledgers/${ledgerPath}/${transaction.ledgerId}/transactions/${transaction.id}`}
			>View</a
		>
	{/each}
</div>

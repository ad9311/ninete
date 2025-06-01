<script lang="ts">
	import type { Transaction } from '$lib/server/db/schema';
	import type { Action } from '$lib/shared';
	import Card from '$lib/components/ui/Card.svelte';
	import Amount from '$lib/components/ledger/Amount.svelte';

	export let transaction: Transaction;
	export let ledgerType: 'budget' | 'payable/receivable';

	const type = transaction.type === 'credit' ? 'credit' : 'debit';
	const ledgerPath = ledgerType === 'budget' ? 'budgets' : 'accounts';
	const actions: Action[] = [
		{
			label: 'Edit',
			href: `/ledgers/${ledgerPath}/${transaction.ledgerId}/transactions/${transaction.id}/edit`
		},
		{
			label: 'Delete',
			onClick: () => {
				// TODO: Implement delete transaction
			}
		}
	];
</script>

<Card>
	{#snippet header()}
		<h2>{transaction.description}</h2>
	{/snippet}
	<p>{transaction.type}</p>
	<p>{transaction.category}</p>
	<p>{transaction.date}</p>
	<p>Amount: <Amount value={transaction.amount} {type} /></p>
	{#snippet footer()}
		<div>
			{#each actions as action (action.label)}
				{#if action.href}
					<a href={action.href}>{action.label}</a>
				{:else}
					<button onclick={action.onClick}>{action.label}</button>
				{/if}
			{/each}
		</div>
	{/snippet}
</Card>

<script lang="ts">
	import type { Transaction } from '$lib/server/db/schema';
	import type { Action } from '$lib/shared';
	import Amount from '../Amount.svelte';

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

<div>
	<h2>{transaction.description}</h2>
	<p>{transaction.type}</p>
	<p>{transaction.category}</p>
	<p>{transaction.date}</p>
	<p>Amount: <Amount value={transaction.amount} {type} /></p>
	<hr />
	<div>
		{#each actions as action (action.label)}
			{#if action.href}
				<a href={action.href}>{action.label}</a>
			{:else}
				<button onclick={action.onClick}>{action.label}</button>
			{/if}
		{/each}
	</div>
</div>

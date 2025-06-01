<script lang="ts">
	import type { Transaction } from '$lib/server/db/schema';
	import type { Action } from '$lib/shared';
	import Card from '$lib/components/ui/Card.svelte';
	import Amount from '$lib/components/ledger/Amount.svelte';
	import { enhance } from '$app/forms';

	export let transaction: Transaction;
	export let ledgerType: 'budget' | 'payable/receivable';

	const type = transaction.type === 'credit' ? 'credit' : 'debit';
	const ledgerPath = ledgerType === 'budget' ? 'budgets' : 'accounts';
	const actions: Action[] = [
		{
			label: 'Delete',
			form: 'delete-transaction-form',
			className: 'btn-destructive',
			submit: true
		},
		{
			label: 'Edit',
			href: `/ledgers/${ledgerPath}/${transaction.ledgerId}/transactions/${transaction.id}/edit`,
			className: 'btn-primary'
		}
	];
</script>

<form method="POST" use:enhance id="delete-transaction-form"></form>

<Card>
	{#snippet header()}
		<h2 class="card-title">{transaction.type}</h2>
	{/snippet}
	<p class="text-muted text-sm italic">{transaction.category}</p>
	<p>{transaction.description}</p>
	<p><Amount value={transaction.amount} {type} /></p>
	<p class="text-muted text-sm italic">{transaction.date}</p>
	{#snippet footer()}
		<div class="flex justify-end gap-2">
			{#each actions as action (action.label)}
				{#if action.href}
					<a href={action.href} class={action.className}>{action.label}</a>
				{:else}
					<button
						type={action.submit ? 'submit' : 'button'}
						onclick={action.onClick}
						class={action.className}
						form={action.form}>{action.label}</button
					>
				{/if}
			{/each}
		</div>
	{/snippet}
</Card>

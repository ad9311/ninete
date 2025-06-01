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
		<div class="flex flex-col gap-1">
			{#each transactions as transaction (transaction.id)}
				<div class="border-muted rounded-xs border p-2 leading-none">
					<div class="mb-2 flex items-center justify-between">
						<p class="text-muted text-sm italic">{transaction.category}</p>
						<a
							href={`/ledgers/${ledgerPath}/${transaction.ledgerId}/transactions/${transaction.id}`}
							class="link">View</a
						>
					</div>
					<p class="text-sm">{transaction.description}</p>
					<p class="text-sm"><Amount value={transaction.amount} {type} /></p>
				</div>
			{/each}
		</div>
	{:else}
		<p class="text-muted text-center text-sm">No transactions</p>
	{/if}
</Card>

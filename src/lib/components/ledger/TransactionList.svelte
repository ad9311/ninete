<script lang="ts">
	import type { Transaction } from '$lib/server/db/schema';
	import Amount from '$lib/components/ledger/Amount.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import * as m from '$lib/paraglide/messages';
	import type { LEDGER_TYPE } from '$lib/shared';

	const {
		transactions,
		type,
		ledgerType
	}: {
		transactions: Transaction[];
		type: 'credit' | 'debit';
		ledgerType: LEDGER_TYPE;
	} = $props();

	const ledgerPath = `${ledgerType}s`;
	const title = type === 'credit' ? 'Credits' : 'Debits';
</script>

<Card>
	{#snippet header()}
		<h3 class="card-title">{title}</h3>
	{/snippet}
	{#if transactions.length > 0}
		<ul class="flex flex-col gap-1">
			{#each transactions as transaction (transaction.id)}
				<li class="border-muted rounded-xs border bg-neutral-50 p-2 leading-normal">
					<div class="mb-1 flex items-center justify-between">
						<p class="text-xs text-zinc-600 italic">
							{m[`transactions.categories.${transaction.category}`]()}
						</p>
						<a
							href={`/ledgers/${ledgerPath}/${transaction.ledgerId}/transactions/${transaction.id}`}
							class="link !text-xs">View Details</a
						>
					</div>
					<div class="mt-4 flex items-center justify-between">
						<p class="text-sm">{transaction.description}</p>
						<p class="text-sm"><Amount value={transaction.amount} {type} /></p>
					</div>
				</li>
			{/each}
		</ul>
	{:else}
		<p class="text-center text-sm text-zinc-600">No transactions found.</p>
	{/if}
</Card>

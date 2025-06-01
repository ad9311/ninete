<script lang="ts">
	import type { Ledger } from '$lib/server/db/schema';
	import { getBalance } from '$lib/shared/ledger';
	import Amount from '$lib/components/ledger/Amount.svelte';
	import Card from '../ui/Card.svelte';

	const { ledger }: { ledger: Ledger } = $props();

	const title = ledger.type === 'budget' ? `${ledger.month}/${ledger.year} Budget` : ledger.title;
	const linkRoot = ledger.type === 'budget' ? '/ledgers/budgets' : '/ledgers/accounts';
</script>

<Card>
	{#snippet header()}
		<div class="flex w-full items-center justify-between">
			<h3 class="card-title">{title}</h3>
			<a href={`${linkRoot}/${ledger.id}`} class="text-xs text-zinc-300 underline hover:text-white"
				>View Details</a
			>
		</div>
	{/snippet}
	<div class="space-y-1 text-sm text-zinc-800">
		<p>
			<span class="font-semibold">Credits:</span>
			<Amount value={ledger.totalCredits} type="credit" />
		</p>
		<p>
			<span class="font-semibold">Debits:</span>
			<Amount value={ledger.totalDebits} type="debit" />
		</p>
		<hr class="my-1 border-t border-zinc-300" />
		<p>
			<span class="font-bold">Balance:</span>
			<Amount value={getBalance(ledger)} type="balance" />
		</p>
	</div>
</Card>

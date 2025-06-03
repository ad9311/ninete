<script lang="ts">
	import type { Ledger } from '$lib/server/db/schema';
	import { getBalance } from '$lib/shared/ledger';
	import Amount from '$lib/components/ledger/Amount.svelte';
	import Card from '../ui/Card.svelte';
	import { formatMonthYear } from '$lib/shared';

	const { ledger }: { ledger: Ledger } = $props();

	const title =
		ledger.type === 'budget'
			? `${formatMonthYear(ledger.month, ledger.year)} Budget`
			: ledger.title;
	const linkRoot = ledger.type === 'budget' ? '/ledgers/budgets' : '/ledgers/accounts';
</script>

<Card>
	{#snippet header()}
		<div class="flex w-full items-center justify-between">
			<h3 class="card-title">{title}</h3>
			<a href={`${linkRoot}/${ledger.id}`} class="link">View Details</a>
		</div>
	{/snippet}
	<div class="space-y-4 text-sm">
		<div class="flex items-center">
			<p class="w-20 font-semibold">Credits:</p>
			<Amount value={ledger.totalCredits} type="credit" />
		</div>
		<div class="flex items-center">
			<p class="w-20 font-semibold">Debits:</p>
			<Amount value={ledger.totalDebits} type="debit" />
		</div>
		<div class="flex items-center">
			<p class="w-20 font-bold">Balance:</p>
			<Amount value={getBalance(ledger)} type="balance" />
		</div>
	</div>
</Card>

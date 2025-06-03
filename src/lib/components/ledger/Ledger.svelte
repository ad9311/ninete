<script lang="ts">
	import type { Ledger } from '$lib/server/db/schema';
	import { getBalance } from '$lib/shared/ledger';
	import Amount from '$lib/components/ledger/Amount.svelte';
	import { formatMonthYear, type Action } from '$lib/shared';
	import Card from '$lib/components/ui/Card.svelte';
	import CardActions from '../ui/CardActions.svelte';

	const { ledger, actions }: { ledger: Ledger; actions: Action[] } = $props();
	const title =
		ledger.type === 'budget'
			? `${formatMonthYear(ledger.month, ledger.year)} Budget`
			: ledger.title;
</script>

<Card className="w-xl">
	{#snippet header()}
		<h2 class="card-title">{title}</h2>
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
	{#snippet footer()}
		<CardActions {actions} />
	{/snippet}
</Card>

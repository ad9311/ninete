<script lang="ts">
	import type { Ledger } from '$lib/server/db/schema';
	import { getBalance } from '$lib/shared/ledger';
	import Amount from '$lib/components/ledger/Amount.svelte';
	import type { Action } from '$lib/shared';
	import Card from '$lib/components/ui/Card.svelte';
	import CardActions from '../ui/CardActions.svelte';

	const { ledger, actions }: { ledger: Ledger; actions: Action[] } = $props();
	const title = ledger.type === 'budget' ? `${ledger.month}/${ledger.year} Budget` : ledger.title;
</script>

<Card>
	{#snippet header()}
		<h2 class="card-title">{title}</h2>
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
	{#snippet footer()}
		<CardActions {actions} />
	{/snippet}
</Card>

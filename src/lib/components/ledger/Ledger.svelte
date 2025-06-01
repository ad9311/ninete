<script lang="ts">
	import type { Ledger } from '$lib/server/db/schema';
	import { getBalance } from '$lib/shared/ledger';
	import Amount from '$lib/components/ledger/Amount.svelte';
	import type { Action } from '$lib/shared';
	import Card from '$lib/components/ui/Card.svelte';

	export let ledger: Ledger;
	export let actions: Action[] = [];

	const title = ledger.type === 'budget' ? `${ledger.month}/${ledger.year} Budget` : ledger.title;
</script>

<Card>
	{#snippet header()}
		<h2 class="card-title">{title}</h2>
	{/snippet}
	<p>Credits: <Amount value={ledger.totalCredits} type="credit" /></p>
	<p>Debits: <Amount value={ledger.totalDebits} type="debit" /></p>
	<p>Balance: <Amount value={getBalance(ledger)} type="balance" /></p>
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

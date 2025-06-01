<script lang="ts">
	import type { Ledger } from '$lib/server/db/schema';
	import { getBalance } from '$lib/shared/ledger';
	import Amount from '$lib/components/ledger/Amount.svelte';
	import type { Action } from '$lib/shared';

	export let ledger: Ledger;
	export let actions: Action[] = [];

	const title = ledger.type === 'budget' ? `${ledger.month}/${ledger.year} Budget` : ledger.title;
</script>

<div>
	<h2 class="text-2xl font-bold">{title}</h2>
	<p>Credits: <Amount value={ledger.totalCredits} type="credit" /></p>
	<p>Debits: <Amount value={ledger.totalDebits} type="debit" /></p>
	<p>Balance: <Amount value={getBalance(ledger)} type="balance" /></p>
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

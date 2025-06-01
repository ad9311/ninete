<script lang="ts">
	import type { Ledger } from '$lib/server/db/schema';
	import { getBalance } from '$lib/shared/ledger';
	import Amount from '$lib/components/ledger/Amount.svelte';
	import Card from '../ui/Card.svelte';

	export let ledger: Ledger;

	const title = ledger.type === 'budget' ? `${ledger.month}/${ledger.year} Budget` : ledger.title;
	const linkRoot = ledger.type === 'budget' ? '/ledgers/budgets' : '/ledgers/accounts';
</script>

<Card>
	{#snippet header()}
		<div class="flex items-center justify-between">
			<h3 class="card-title">{title}</h3>
			<a href={`${linkRoot}/${ledger.id}`}>View</a>
		</div>
	{/snippet}
	<p>Credits: <Amount value={ledger.totalCredits} type="credit" /></p>
	<p>Debits: <Amount value={ledger.totalDebits} type="debit" /></p>
	<p>Balance: <Amount value={getBalance(ledger)} type="balance" /></p>
</Card>

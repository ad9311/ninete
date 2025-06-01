<script lang="ts">
	import Ledger from '$lib/components/ledger/Ledger.svelte';
	import TransactionList from '$lib/components/ledger/TransactionList.svelte';
	import type { Action } from '$lib/shared';
	import type { PageData } from './$types';

	const { data }: { data: PageData } = $props();
	const { budget, credits, debits } = data;
	const actions: Action[] = [
		{
			label: 'New Transaction',
			href: `/ledgers/budgets/${budget.id}/transactions/new`
		}
	];
</script>

<Ledger ledger={budget} {actions} />
<br />
<TransactionList transactions={credits} ledgerType="budget" type="credit" />
<br />
<TransactionList transactions={debits} ledgerType="budget" type="debit" />

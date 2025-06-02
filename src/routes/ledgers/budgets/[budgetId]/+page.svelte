<script lang="ts">
	import Ledger from '$lib/components/ledger/Ledger.svelte';
	import TransactionList from '$lib/components/ledger/TransactionList.svelte';
	import Breadcrumb from '$lib/components/ui/Breadcrumb.svelte';
	import type { BreadcrumbItem } from '$lib/client';
	import type { Action } from '$lib/shared';
	import type { PageData } from './$types';

	const { data }: { data: PageData } = $props();

	const { budget, credits, debits } = data;
	const breadcrumbItems: BreadcrumbItem[] = [
		{ label: 'Home', href: '/' },
		{ label: 'Budgets', href: '/ledgers/budgets' },
		{ label: 'TODO' }
	];
	const actions: Action[] = [
		{
			label: 'New Transaction',
			href: `/ledgers/budgets/${budget.id}/transactions/new`
		}
	];
</script>

<Breadcrumb items={breadcrumbItems} />
<Ledger ledger={budget} {actions} />
<br />
<TransactionList transactions={credits} ledgerType="budget" type="credit" />
<br />
<TransactionList transactions={debits} ledgerType="budget" type="debit" />

<script lang="ts">
	import Ledger from '$lib/components/ledger/Ledger.svelte';
	import Breadcrumb from '$lib/components/ui/Breadcrumb.svelte';
	import type { BreadcrumbItem } from '$lib/client';
	import type { PageData } from './$types';
	import type { Action } from '$lib/shared';
	import TransactionList from '$lib/components/ledger/TransactionList.svelte';

	const { data }: { data: PageData } = $props();

	const { loan, credits, debits } = data;
	const breadcrumbItems: BreadcrumbItem[] = [
		{ label: 'Home', href: '/' },
		{ label: 'Loans', href: '/ledgers/loans' },
		{ label: loan.title as string }
	];
	const actions: Action[] = [
		{
			label: 'New Transaction',
			href: `/ledgers/loans/${loan.id}/transactions/new`
		}
	];
</script>

<Breadcrumb items={breadcrumbItems} />
<Ledger ledger={loan} {actions} />
<br />
<TransactionList transactions={credits} ledgerType="loan" type="credit" />
<br />
<TransactionList transactions={debits} ledgerType="loan" type="debit" />

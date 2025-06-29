<script lang="ts">
	import Ledger from '$lib/components/ledger/Ledger.svelte';
	import Breadcrumb from '$lib/components/ui/Breadcrumb.svelte';
	import type { BreadcrumbItem } from '$lib/client';
	import type { PageData } from './$types';
	import type { Action } from '$lib/shared';
	import TransactionList from '$lib/components/ledger/TransactionList.svelte';

	const { data }: { data: PageData } = $props();

	const { payable, credits, debits } = data;
	const breadcrumbItems: BreadcrumbItem[] = [
		{ label: 'Home', href: '/' },
		{ label: 'Loans', href: '/ledgers/loans' },
		{ label: payable.title as string }
	];
	const actions: Action[] = [
		{
			label: 'New Transaction',
			href: `/ledgers/loans/${payable.id}/transactions/new`
		}
	];
</script>

<Breadcrumb items={breadcrumbItems} />
<Ledger ledger={payable} {actions} />
<br />
<TransactionList transactions={credits} ledgerType="payable" type="credit" />
<br />
<TransactionList transactions={debits} ledgerType="payable" type="debit" />

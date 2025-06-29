<script lang="ts">
	import Transaction from '$lib/components/ledger/Transaction.svelte';
	import type { BreadcrumbItem } from '$lib/client';
	import Breadcrumb from '$lib/components/ui/Breadcrumb.svelte';
	import type { PageData } from './$types';

	const { data }: { data: PageData } = $props();

	const { transaction, loan } = data;
	const breadcrumbItems: BreadcrumbItem[] = [
		{ label: 'Home', href: '/' },
		{ label: 'Loans', href: '/ledgers/loans' },
		{ label: 'Loan', href: `/ledgers/loans/${loan.id}` },
		{ label: 'Transaction' }
	];
</script>

<Breadcrumb items={breadcrumbItems} />
<Transaction {transaction} ledgerType="loan" />

<script lang="ts">
	import Transaction from '$lib/components/ledger/Transaction.svelte';
	import type { BreadcrumbItem } from '$lib/client';
	import Breadcrumb from '$lib/components/ui/Breadcrumb.svelte';
	import type { PageData } from './$types';

	const { data }: { data: PageData } = $props();

	const { transaction, payable } = data;
	const breadcrumbItems: BreadcrumbItem[] = [
		{ label: 'Home', href: '/' },
		{ label: 'Payables', href: '/ledgers/payables' },
		{ label: 'Payable', href: `/ledgers/payables/${payable.id}` },
		{ label: 'Transaction' }
	];
</script>

<Breadcrumb items={breadcrumbItems} />
<Transaction {transaction} ledgerType="payable" />

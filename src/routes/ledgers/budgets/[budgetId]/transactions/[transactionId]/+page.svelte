<script lang="ts">
	import Transaction from '$lib/components/ledger/Transaction.svelte';
	import type { BreadcrumbItem } from '$lib/client';
	import Breadcrumb from '$lib/components/ui/Breadcrumb.svelte';
	import type { PageData } from './$types';

	const { data }: { data: PageData } = $props();

	const { transaction, budget } = data;
	const breadcrumbItems: BreadcrumbItem[] = [
		{ label: 'Home', href: '/' },
		{ label: 'Budgets', href: '/ledgers/budgets' },
		{ label: 'Budget', href: `/ledgers/budgets/${budget.id}` },
		{ label: 'Transaction' }
	];
</script>

<Breadcrumb items={breadcrumbItems} />
<Transaction {transaction} ledgerType="budget" />

<script lang="ts">
	import type { BreadcrumbItem } from '$lib/client';
	import LedgerList from '$lib/components/ledger/LedgerList.svelte';
	import Breadcrumb from '$lib/components/ui/Breadcrumb.svelte';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();

	const { budgets } = data;
	const breadcrumbItems: BreadcrumbItem[] = [{ label: 'Home', href: '/' }, { label: 'Budgets' }];
</script>

<Breadcrumb items={breadcrumbItems} />
<LedgerList ledgers={budgets} />

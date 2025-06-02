<script lang="ts">
	import type { BreadcrumbItem } from '$lib/client';
	import Breadcrumb from '$lib/components/ui/Breadcrumb.svelte';
	import LedgerSummary from '$lib/components/ledger/LedgerSummary.svelte';
	import type { PageData } from './$types';

	const { data }: { data: PageData } = $props();
	const { currentBudget } = data;

	const breadcrumbItems: BreadcrumbItem[] = [{ label: 'Home' }];
</script>

<Breadcrumb items={breadcrumbItems} />
<LedgerSummary ledger={currentBudget} />

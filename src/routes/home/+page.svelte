<script lang="ts">
	import LedgerSummary from '$lib/components/ledger/LedgerSummary.svelte';
	import type { PageData } from './$types';

	const { data }: { data: PageData } = $props();
	const { currentBudget } = data;
</script>

<LedgerSummary ledger={currentBudget} />

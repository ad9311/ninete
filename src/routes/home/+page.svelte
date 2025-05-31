<script lang="ts">
	import LedgerSummary from '$lib/components/ledger/LedgerSummary.svelte';
	import type { PageData } from './$types';

	const { data }: { data: PageData } = $props();
</script>

<LedgerSummary ledger={data.currentBudget} />

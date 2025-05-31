<script lang="ts">
	import Ledger from '$lib/components/ledger/Ledger.svelte';
	import type { Action } from '$lib/shared';
	import type { PageData } from './$types';

	const { data }: { data: PageData } = $props();
	const { budget } = data;
	const actions: Action[] = [
		{
			label: 'New Transaction',
			href: `/ledgers/budgets/${budget.id}/transactions/new`
		}
	];
</script>

<Ledger ledger={budget} {actions} />

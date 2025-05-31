<script lang="ts">
	import Transaction from '$lib/components/ledger/Transaction.svelte';
	import type { PageData } from './$types';

	const { data }: { data: PageData } = $props();
	const { transaction } = data;
</script>

<Transaction {transaction} ledgerType="budget" />

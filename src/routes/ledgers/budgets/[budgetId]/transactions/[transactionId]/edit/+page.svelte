<script lang="ts">
	import { type BreadcrumbItem } from '$lib/client';
	import FormErrors from '$lib/components/form/FormErrors.svelte';
	import Breadcrumb from '$lib/components/ui/Breadcrumb.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Form from '$lib/components/transaction/Form.svelte';
	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();

	const { transaction, budget } = data;
	const breadcrumbItems: BreadcrumbItem[] = [
		{ label: 'Home', href: '/' },
		{ label: 'Budgets', href: '/ledgers/budgets' },
		{ label: 'Budget', href: `/ledgers/budgets/${budget.id}` },
		{ label: 'Transaction', href: `/ledgers/budgets/${budget.id}/transactions/${transaction.id}` },
		{ label: 'Edit' }
	];
</script>

<Breadcrumb items={breadcrumbItems} />
<Card className="max-w-xl">
	{#snippet header()}
		<h3 class="card-title">Edit Transaction</h3>
	{/snippet}
	<FormErrors errors={form?.errors} />
	<Form ledgerType="budget" {transaction} />
</Card>

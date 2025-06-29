<script lang="ts">
	import FormErrors from '$lib/components/form/FormErrors.svelte';
	import Breadcrumb from '$lib/components/ui/Breadcrumb.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Form from '$lib/components/transaction/Form.svelte';
	import type { PageProps } from './$types';
	import type { BreadcrumbItem } from '$lib/client';

	const { data, form }: PageProps = $props();
	const { loan } = data;

	const breadcrumbItems: BreadcrumbItem[] = [
		{ label: 'Home', href: '/' },
		{ label: 'Loans', href: '/ledgers/loans' },
		{ label: 'Loan', href: `/ledgers/loans/${loan.id}` },
		{ label: 'New Transaction' }
	];
</script>

<Breadcrumb items={breadcrumbItems} />
<Card className="max-w-xl">
	{#snippet header()}
		<h3 class="card-title">New Transaction</h3>
	{/snippet}
	<FormErrors errors={form?.errors} />
	<Form ledgerType="loan" />
</Card>

<script lang="ts">
	import type { BreadcrumbItem } from '$lib/client';
	import FormErrors from '$lib/components/form/FormErrors.svelte';
	import Form from '$lib/components/ledger/loan/Form.svelte';
	import Breadcrumb from '$lib/components/ui/Breadcrumb.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import type { PageProps } from './$types';

	const { form }: PageProps = $props();

	const breadcrumbItems: BreadcrumbItem[] = [
		{ label: 'Home', href: '/' },
		{ label: 'Loans', href: '/ledgers/loans' },
		{ label: 'New' }
	];
</script>

<Breadcrumb items={breadcrumbItems} />
<Card className="max-w-xl">
	{#snippet header()}
		<h3 class="card-title">New Loan</h3>
	{/snippet}
	<FormErrors errors={form?.errors} />
	<Form />
</Card>

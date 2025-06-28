<script lang="ts">
	import FormErrors from '$lib/components/form/FormErrors.svelte';
	import Form from '$lib/components/ledger/payable/Form.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import type { PageProps } from './$types';

	const { form }: PageProps = $props();
</script>

<Card className="max-w-xl">
	{#snippet header()}
		<h3 class="card-title">New Payable</h3>
	{/snippet}
	<FormErrors errors={form?.errors} />
	<Form />
</Card>

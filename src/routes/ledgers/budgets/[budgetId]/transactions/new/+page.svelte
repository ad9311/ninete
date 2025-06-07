<script lang="ts">
	import { enhance } from '$app/forms';
	import { mapTransactionCategories, type BreadcrumbItem } from '$lib/client';
	import FormErrors from '$lib/components/form/FormErrors.svelte';
	import Breadcrumb from '$lib/components/ui/Breadcrumb.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import { TRANSACTION_TYPES } from '$lib/shared';
	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();
	const { budget } = data;

	const breadcrumbItems: BreadcrumbItem[] = [
		{ label: 'Home', href: '/' },
		{ label: 'Budgets', href: '/ledgers/budgets' },
		{ label: 'Budget', href: `/ledgers/budgets/${budget.id}` },
		{ label: 'New Transaction' }
	];
	const categories = mapTransactionCategories();
</script>

<Breadcrumb items={breadcrumbItems} />
<Card className="max-w-xl">
	{#snippet header()}
		<h3 class="card-title">New Transaction</h3>
	{/snippet}
	<FormErrors errors={form?.errors} />
	<form method="post" class="form" use:enhance>
		<div class="form-group">
			<label for="type">Type </label>
			<select id="type" name="type">
				{#each TRANSACTION_TYPES as type (type)}
					<option value={type}>{type === 'credit' ? 'Credit' : 'Debit'}</option>
				{/each}
			</select>
		</div>
		<div class="form-group">
			<label for="description">Description </label>
			<textarea id="description" name="description" placeholder="Description"></textarea>
		</div>
		<div class="form-group">
			<label for="category">Category</label>
			<select id="category" name="category">
				{#each categories as category (category)}
					<option value={category.value}>{category.label}</option>
				{/each}
			</select>
		</div>
		<div class="form-group">
			<label for="amount">Amount</label>
			<input id="amount" type="number" name="amount" placeholder="Amount" step="0.01" min="0" />
		</div>
		<div class="form-group">
			<label for="date">Date</label>
			<input id="date" type="date" name="date" />
		</div>
		<div class="form-group flex items-center gap-2">
			<label for="is_estimated">is_estimated</label>
			<input id="is_estimated" type="checkbox" name="is_estimated" />
		</div>
		<div class="form-actions">
			<button type="submit" class="btn-primary">Create</button>
		</div>
	</form>
</Card>

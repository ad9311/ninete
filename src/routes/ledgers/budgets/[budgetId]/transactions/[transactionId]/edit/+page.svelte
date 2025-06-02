<script lang="ts">
	import { enhance } from '$app/forms';
	import type { BreadcrumbItem } from '$lib/client';
	import FormErrors from '$lib/components/form/FormErrors.svelte';
	import Breadcrumb from '$lib/components/ui/Breadcrumb.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import { TRANSACTION_CATEGORIES } from '$lib/shared';
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
<Card>
	{#snippet header()}
		<h3 class="card-title">Edit Transaction</h3>
	{/snippet}
	<FormErrors errors={form?.errors} />
	<form method="post" use:enhance class="form">
		<p>{transaction.type}</p>
		<div class="form-group">
			<label for="description">Description</label>
			<input
				id="description"
				type="text"
				name="description"
				placeholder="Description"
				value={transaction.description}
			/>
		</div>
		<div class="form-group">
			<label for="category">Category</label>
			<select id="category" name="category" value={transaction.category}>
				{#each TRANSACTION_CATEGORIES as category (category)}
					<option value={category}>{category}</option>
				{/each}
			</select>
		</div>
		<div class="form-group">
			<label for="amount">Amount</label>
			<input
				id="amount"
				type="number"
				name="amount"
				placeholder="Amount"
				step="0.01"
				min="0"
				value={transaction.amount}
			/>
		</div>
		<div class="form-group">
			<label for="date">Date</label>
			<input
				id="date"
				type="date"
				name="date"
				value={transaction.date.toISOString().split('T')[0]}
			/>
		</div>
		<div class="form-actions">
			<button type="submit" class="btn-primary">Update</button>
		</div>
	</form>
</Card>

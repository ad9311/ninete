<script lang="ts">
	import { enhance } from '$app/forms';
	import { mapTransactionCategories, formatDateForInput } from '$lib/client';
	import type { Transaction } from '$lib/server/db/schema';
	import { TRANSACTION_TYPES } from '$lib/shared';

	const { transaction }: { transaction?: Transaction } = $props();

	const categories = mapTransactionCategories();
</script>

<form method="post" class="form" use:enhance>
	<div class="form-group">
		<label for="type">Type </label>
		<select id="type" name="type" value={transaction?.type}>
			{#each TRANSACTION_TYPES as type (type)}
				<option value={type}>{type === 'credit' ? 'Credit' : 'Debit'}</option>
			{/each}
		</select>
	</div>
	<div class="form-group">
		<label for="description">Description </label>
		<textarea id="description" name="description" placeholder="Description"
			>{transaction?.description}</textarea
		>
	</div>
	<div class="form-group">
		<label for="category">Category</label>
		<select id="category" name="category" value={transaction?.category}>
			{#each categories as category (category)}
				<option value={category.value}>{category.label}</option>
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
			value={transaction?.amount}
		/>
	</div>
	<div class="form-group">
		<label for="date">Date</label>
		<input
			id="date"
			type="date"
			name="date"
			value={transaction?.date ? formatDateForInput(transaction.date) : undefined}
		/>
	</div>
	<div class="form-group flex items-center gap-2">
		<label for="is_estimated">Estimated?</label>
		<input id="is_estimated" type="checkbox" name="is_estimated" value={transaction?.isEstimated} />
	</div>
	<div class="form-actions">
		<button type="submit" class="btn-primary">Create</button>
	</div>
</form>

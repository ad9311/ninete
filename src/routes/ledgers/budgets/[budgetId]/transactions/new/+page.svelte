<script lang="ts">
	import { enhance } from '$app/forms';
	import Card from '$lib/components/ui/Card.svelte';
	import { TRANSACTION_CATEGORIES, TRANSACTION_TYPES } from '$lib/shared';
</script>

<Card>
	{#snippet header()}
		<h3 class="card-title">New Transaction</h3>
	{/snippet}
	<form method="post" use:enhance class="form">
		<div class="form-group">
			<label for="type">Type </label>
			<select id="type" name="type">
				{#each TRANSACTION_TYPES as type (type)}
					<option value={type}>{type}</option>
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
				{#each TRANSACTION_CATEGORIES as category (category)}
					<option value={category}>{category}</option>
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
		<div class="form-actions">
			<button type="submit" class="btn-primary">Create</button>
		</div>
	</form>
</Card>

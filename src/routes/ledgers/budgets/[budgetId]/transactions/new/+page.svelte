<script lang="ts">
	import { enhance } from '$app/forms';
	import { TRANSACTION_CATEGORIES, TRANSACTION_TYPES } from '$lib/shared';
</script>

<form method="post" use:enhance>
	<label for="type">Type</label>
	<select id="type" name="type">
		{#each TRANSACTION_TYPES as type (type)}
			<option value={type}>{type}</option>
		{/each}
	</select>
	<label for="description">Description</label>
	<input id="description" type="text" name="description" placeholder="Description" />
	<label for="category">Category</label>
	<select id="category" name="category">
		{#each TRANSACTION_CATEGORIES as category (category)}
			<option value={category}>{category}</option>
		{/each}
	</select>
	<label for="amount">Amount</label>
	<input id="amount" type="number" name="amount" placeholder="Amount" step="0.01" min="0" />
	<label for="date">Date</label>
	<input id="date" type="date" name="date" />
	<button type="submit">Create</button>
</form>

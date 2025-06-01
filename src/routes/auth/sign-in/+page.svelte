<script lang="ts">
	import { enhance } from '$app/forms';
	import Card from '$lib/components/ui/Card.svelte';
</script>

<Card>
	{#snippet header()}
		<h3 class="card-title">Welcome back!</h3>
	{/snippet}
	<form method="POST" use:enhance>
		<label for="email">Email</label>
		<input type="email" name="email" id="email" />

		<label for="password">Password</label>
		<input type="password" name="password" id="password" />

		<button type="submit">Sign in</button>
	</form>
</Card>

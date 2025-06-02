<script lang="ts">
	import { enhance } from '$app/forms';
	import Card from '$lib/components/ui/Card.svelte';
	import CardActions from '$lib/components/ui/CardActions.svelte';
</script>

<Card>
	{#snippet header()}
		<h3 class="card-title">Welcome back!</h3>
	{/snippet}
	<form method="POST" use:enhance class="form">
		<div class="form-group">
			<label for="email">Email</label>
			<input type="email" name="email" id="email" placeholder="example@mail.com" />
		</div>

		<div class="form-group">
			<label for="password">Password</label>
			<input type="password" name="password" id="password" placeholder="********" />
		</div>

		<div class="form-actions">
			<button type="submit" class="btn-primary">Sign in</button>
		</div>
	</form>
	{#snippet footer()}
		<CardActions actions={[{ label: 'Create an account', href: '/auth/sign-up' }]} />
	{/snippet}
</Card>

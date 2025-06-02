<script lang="ts">
	import { enhance } from '$app/forms';
	import Card from '$lib/components/ui/Card.svelte';
</script>

<Card>
	{#snippet header()}
		<h3 class="card-title">Create an account</h3>
	{/snippet}
	<form method="POST" use:enhance class="form">
		<div class="form-group">
			<label for="username">Username</label>
			<input type="text" name="username" id="username" placeholder="john_doe" />
		</div>
		<div class="form-group">
			<label for="email">Email</label>
			<input type="email" name="email" id="email" placeholder="example@mail.com" />
		</div>
		<div class="form-group">
			<label for="password">Password</label>
			<input type="password" name="password" id="password" placeholder="********" />
		</div>
		<div class="form-group">
			<label for="password-confirmation">Password Confirmation</label>
			<input
				type="password"
				name="password-confirmation"
				id="password-confirmation"
				placeholder="********"
			/>
		</div>
		<div class="form-actions">
			<button type="submit" class="btn-primary">Sign up</button>
		</div>
	</form>
	{#snippet footer()}
		<div class="flex justify-between text-sm">
			<a href="/auth/sign-in" class="link">Already have an account?</a>
		</div>
	{/snippet}
</Card>

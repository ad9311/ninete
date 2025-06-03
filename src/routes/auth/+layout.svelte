<script lang="ts">
	const { children } = $props();
</script>

<main class="container mx-auto">
	<h1 class="my-20 text-center text-3xl font-bold text-primary">NineTe</h1>
	<div class="mx-auto max-w-xl">
		{@render children()}
	</div>
</main>

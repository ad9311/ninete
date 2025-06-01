<script lang="ts">
	import '../app.css';

	const { children, data } = $props();
</script>

<svelte:head>
	<title>NineTe</title>
	<meta name="description" content="NineTe is a personal app for managing my life" />
</svelte:head>

{#if data.isUserSignedIn}
	<main class="p-2 md:p-4">
		{@render children()}
	</main>
{:else}
	<main class="p-2 md:p-4">
		{@render children()}
	</main>
{/if}

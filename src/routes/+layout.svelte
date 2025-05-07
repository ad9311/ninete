<script lang="ts">
	import '../app.css';

	const { children, data } = $props();
</script>

<svelte:head>
	<title>NineTe</title>
	<meta name="description" content="NineTe is a personal app for managing my life" />
</svelte:head>

{#if data.isUserSignedIn}
	<main>
		{@render children()}
	</main>
{:else}
	<main>
		{@render children()}
	</main>
{/if}

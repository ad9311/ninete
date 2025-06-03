<script lang="ts">
	import Sidebar from '$lib/components/ui/Sidebar.svelte';
	import '../app.css';

	const { children, data } = $props();
</script>

<svelte:head>
	<title>NineTe</title>
	<meta name="description" content="NineTe is a personal app for managing my life" />
</svelte:head>

{#if data.isUserSignedIn}
	<div class="flex">
		<Sidebar>NineTe</Sidebar>
		<main class="w-full p-2 pt-13 md:p-4">
			{@render children()}
		</main>
	</div>
{:else}
	<main class="p-2 md:p-4">
		{@render children()}
	</main>
{/if}

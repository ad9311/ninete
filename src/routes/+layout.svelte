<script lang="ts">
	import Sidebar from '$lib/components/ui/Sidebar.svelte';
	import SidebarContent from '$lib/components/nav/SidebarContent.svelte';
	import type { NavLink } from '$lib/client';
	import { BanknoteArrowDown, House, WalletCards } from 'lucide-svelte';
	import '../app.css';

	const { children, data } = $props();

	const navLinks: NavLink[] = [
		{
			label: 'Home',
			path: '/home',
			icon: House
		},
		{
			label: 'Budgets',
			path: '/ledgers/budgets',
			icon: WalletCards
		},
		{
			label: 'Loans',
			path: '/ledgers/loans',
			icon: BanknoteArrowDown
		}
	];
</script>

<svelte:head>
	<title>NineTe</title>
	<meta name="description" content="NineTe is a personal app for managing my life" />
</svelte:head>

{#if data.isUserSignedIn}
	<div class="flex">
		<Sidebar>
			<SidebarContent {navLinks} />
		</Sidebar>
		<main class="w-full p-2 pt-13 md:p-4">
			{@render children()}
		</main>
	</div>
{:else}
	<main class="p-2 md:p-4">
		{@render children()}
	</main>
{/if}

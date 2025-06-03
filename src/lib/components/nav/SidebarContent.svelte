<script lang="ts">
	import { CircleUserRound } from 'lucide-svelte';
	import type { NavLink } from '$lib/client';

	const { navLinks = [] }: { navLinks: NavLink[] } = $props();
</script>

<form action="/auth/sign-out" method="POST" id="sign-out-formb"></form>
<div class="relative flex items-center justify-between">
	<h1 class="text-2xl font-bold">NineTe</h1>
	<CircleUserRound size={30} strokeWidth={1} style="color: var(--color-primary);" />
</div>
<hr class="my-4 border-zinc-400" />
<ul class="space-y-2">
	{#each navLinks as link (link.path)}
		<li>
			{#if link.icon}
				<link.icon />
			{/if}
			<a
				href={link.path}
				class="border-primary text-primary block w-full rounded-xs border bg-zinc-50 px-2 py-1 transition-colors hover:bg-zinc-100"
				>{link.lable}</a
			>
		</li>
	{/each}
	<div class="absolute bottom-8 w-full text-center">
		<button
			type="submit"
			form="sign-out-form"
			class="cursor-pointer text-sm text-red-600 hover:text-red-800 hover:underline"
			>Sign out</button
		>
	</div>
</ul>

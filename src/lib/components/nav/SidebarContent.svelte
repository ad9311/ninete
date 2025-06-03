<script lang="ts">
	import { CircleUserRound } from 'lucide-svelte';
	import type { NavLink } from '$lib/client';

	const { navLinks = [] }: { navLinks: NavLink[] } = $props();
</script>

<div class="flex items-center justify-between">
	<h1 class="text-2xl font-bold">NineTe</h1>
	<CircleUserRound size={30} strokeWidth={1} style="color: var(--color-primary);" />
</div>
<hr class="my-4 border-zinc-400" />
<ul class="space-y-2">
	{#each navLinks as link (link.path)}
		<li>
			<a
				href={link.path}
				class="border-primary text-primary block w-full rounded-xs border bg-zinc-50 px-2 py-1 transition-colors hover:bg-zinc-100"
				>{link.lable}</a
			>
		</li>
	{/each}
</ul>

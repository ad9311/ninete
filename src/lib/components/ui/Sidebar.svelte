<script lang="ts">
	import { Menu } from 'lucide-svelte';
	import type { Snippet } from 'svelte';

	type SidebarProps = {
		children: Snippet;
	};

	const { children }: SidebarProps = $props();

	let toggle: boolean = $state(false);

	const handleToggle = () => {
		toggle = !toggle;
	};
</script>

<div
	class={`fixed top-0 left-0 z-40 h-screen w-64 transform transition-transform duration-200
	${toggle ? 'translate-x-0' : '-translate-x-full'} md:static md:translate-x-0 md:transform-none`}
>
	<aside class="border-primary h-full w-64 border bg-white p-2 md:p-4">
		{@render children()}
	</aside>
	<button
		class="absolute top-2 {toggle
			? '-right-11'
			: '-right-10'} border-primary cursor-pointer rounded-xs border bg-white p-1 md:hidden"
		onclick={handleToggle}
	>
		<Menu />
	</button>
</div>

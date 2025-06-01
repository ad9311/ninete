<script lang="ts">
	import type { Snippet } from 'svelte';

	type CardProps = {
		children?: Snippet;
		header?: Snippet;
		footer?: Snippet;
		className?: string;
	};

	const { children, header, footer, className }: CardProps = $props();
</script>

<div class={className ?? ''}>
	<div
		class="rounded-sm border-2 border-t-zinc-50 border-r-zinc-400 border-b-zinc-400 border-l-zinc-50 bg-zinc-200 font-sans text-sm shadow-md"
	>
		<header class="flex cursor-default items-center bg-blue-600 px-2 py-1 text-white">
			{#if header}
				<span
					class="mr-2 inline-block h-3 w-3 border border-t-zinc-50 border-r-zinc-400 border-b-zinc-400 border-l-zinc-50 bg-zinc-300"
				></span>
				{@render header()}
			{/if}
		</header>

		<section
			class="border-t-primary border-l-primary m-1 min-h-[50px] border-2 border-r-zinc-300 border-b-zinc-300 bg-white p-3"
		>
			{#if children}
				<div>{@render children()}</div>
			{:else}
				<div class="text-zinc-500 italic">No content provided.</div>
			{/if}
		</section>

		{#if footer}
			<footer
				class="text-primary mt-1 border-t-2 border-t-zinc-400 bg-zinc-200 px-2 py-0.5 text-xs"
			>
				{@render footer()}
			</footer>
		{/if}
	</div>
</div>

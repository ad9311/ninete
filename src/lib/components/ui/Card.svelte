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
	<section class="border-primary rounded-xs border p-4">
		{#if header}
			<header>{@render header()}</header>
			<hr class="border-primary my-4 border-t" />
		{/if}
		{#if children}
			<div>{@render children()}</div>
		{/if}
		{#if footer}
			<hr class="my-2" />
			<footer>{@render footer()}</footer>
		{/if}
	</section>
</div>

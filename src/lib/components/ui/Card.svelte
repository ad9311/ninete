<script lang="ts">
	import type { Snippet } from 'svelte';

	const {
		children,
		header,
		footer,
		className
	}: { children?: Snippet; header?: Snippet; footer?: Snippet; className?: string } = $props();
</script>

<div class={className ?? ''}>
	<section class="border-primary rounded-xs border bg-white p-4">
		<header>
			{#if header}
				{@render header()}
			{/if}
			<hr class="my-4 border-zinc-400" />
		</header>
		<div>
			{#if children}
				<div>{@render children()}</div>
			{:else}
				<div class="">No content provided.</div>
			{/if}
		</div>
		{#if footer}
			<hr class="my-4 border-zinc-400" />
			<div>
				{@render footer()}
			</div>
		{/if}
	</section>
</div>

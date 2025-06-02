<script lang="ts">
	import type { BreadcrumbItem } from '$lib/client';

	type Props = {
		items: BreadcrumbItem[];
		className?: string;
	};

	const { items, className = '' }: Props = $props();
</script>

{#if items && items.length > 0}
	<nav aria-label="Breadcrumb" class="win98-breadcrumb-bar {className}">
		<ol class="flex items-center">
			{#each items as item, i (item.label)}
				<li>
					{#if item.href && i < items.length - 1}
						<a
							href={item.href}
							class="link rounded-sm px-1 py-0.5 text-xs hover:text-blue-600 hover:underline"
							>{item.label}</a
						>
					{:else}
						<span
							class="px-1 py-0.5 text-xs {i === items.length - 1
								? 'font-semibold text-black'
								: 'text-zinc-700'}">{item.label}</span
						>
					{/if}
				</li>
				{#if i < items.length - 1}
					<li aria-hidden="true">
						<span class="mx-0.5 text-xs text-zinc-500">/</span>
					</li>
				{/if}
			{/each}
		</ol>
	</nav>
{/if}

<style lang="postcss">
	.win98-breadcrumb-bar {
		user-select: none;
		background-color: var(--color-zinc-200);
		padding-left: 0.5rem;
		padding-right: 0.5rem;
		padding-top: 0.25rem;
		padding-bottom: 0.25rem;
		font-size: 0.75rem;
		line-height: 1rem;
		border: 1px solid;
		border-top-color: var(--color-zinc-50);
		border-left-color: var(--color-zinc-50);
		border-right-color: var(--color-zinc-400);
		border-bottom-color: var(--color-zinc-400);
	}
</style>

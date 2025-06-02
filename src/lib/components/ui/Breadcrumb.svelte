<script lang="ts">
	import type { BreadcrumbItem } from '$lib/client';

	type Props = {
		items: BreadcrumbItem[];
	};

	const { items }: Props = $props();

	const MAX_ITEMS_VISIBLE = 4;
	let processedItems: BreadcrumbItem[] = $state([]);

	$effect(() => {
		if (items.length > MAX_ITEMS_VISIBLE) {
			processedItems = [
				{ label: '...', href: undefined },
				...items.slice(items.length - (MAX_ITEMS_VISIBLE - 1))
			];
		} else {
			processedItems = items;
		}
	});
</script>

{#if processedItems && processedItems.length > 0}
	<nav
		aria-label="Breadcrumb"
		class="breadcrumb-bar border-primary mb-4 rounded-xs border bg-white p-2 text-xs"
	>
		<ol class="flex items-center">
			{#each processedItems as pItem, pIndex (pItem)}
				<li>
					{#if pItem.href && pItem !== items[items.length - 1]}
						<a href={pItem.href} class="link">{pItem.label}</a>
					{:else}
						<span class="text-sm text-zinc-600">{pItem.label}</span>
					{/if}
				</li>
				{#if pIndex < processedItems.length - 1}
					<li aria-hidden="true">
						<span class="mx-0.5 text-xs text-zinc-500">/</span>
					</li>
				{/if}
			{/each}
		</ol>
	</nav>
{/if}

<style lang="postcss">
	.breadcrumb-bar {
		user-select: none;
		/* padding-left: 0.5rem;
		padding-right: 0.5rem;
		padding-top: 0.25rem;
		padding-bottom: 0.25rem;
		margin-bottom: 1rem;
		font-size: 0.75rem;
		line-height: 1rem;
		border: 1px solid var(--color-primary);
		border-radius: var(--radius-xs); */
	}
</style>

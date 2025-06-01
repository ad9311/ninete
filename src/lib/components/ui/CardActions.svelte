<script lang="ts">
	import type { Action } from '$lib/shared';

	type Props = {
		actions: Action[];
	};

	const { actions }: Props = $props();
</script>

{#if actions.length > 0}
	<div class="flex w-full items-center justify-end gap-2">
		{#each actions as action (action.label)}
			{#if action.href}
				<a href={action.href} class="link text-xs {action.className ?? ''}">{action.label}</a>
			{:else if action.onClick}
				<button
					type="button"
					class="btn-primary text-xs {action.className ?? ''}"
					onclick={action.onClick}
				>
					{action.label}
				</button>
			{/if}
		{/each}
	</div>
{/if}

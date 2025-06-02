<script lang="ts">
	export let errors: string[] | undefined = [];
</script>

{#if errors && errors.length > 0 && errors[0] != ''}
	<ul class="mb-4 rounded-xs border border-red-500 bg-red-100 p-2">
		{#each errors as error, index (error + index)}
			<li class="text-xs text-red-500">{error}</li>
		{/each}
	</ul>
{/if}

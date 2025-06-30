<script lang="ts">
	import type { Ledger } from '$lib/server/db/schema';
	import { formatMonthYear } from '$lib/shared';

	const { ledger }: { ledger: Ledger } = $props();

	const getTitle = () => {
		if (ledger.type === 'budget') {
			return formatMonthYear(ledger.month, ledger.year);
		}

		return ledger.title;
	};

	const getPath = () => {
		switch (ledger.type) {
			case 'budget':
				return `/ledgers/budgets/${ledger.id}`;
			case 'loan':
				return `/ledgers/loans/${ledger.id}`;
			default:
				return '#';
		}
	};
</script>

<div class="border-muted rounded-xs border bg-neutral-50 p-2 leading-normal">
	<div class="mb-1 flex items-center justify-between">
		<p class="text-zinc-600">
			{getTitle()}
		</p>
		<a href={getPath()} class="link !text-xs">View Details</a>
	</div>
	<div class="mt-4 flex items-center justify-between">
		<p class="text-sm">{ledger.description}</p>
	</div>
</div>

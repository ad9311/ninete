<script lang="ts">
	import type { Transaction } from '$lib/server/db/schema';
	import { formatDateToMonthYear, type Action, type LEDGER_TYPE } from '$lib/shared';
	import Card from '$lib/components/ui/Card.svelte';
	import Amount from '$lib/components/ledger/Amount.svelte';
	import * as m from '$lib/paraglide/messages';
	import { enhance } from '$app/forms';

	const { transaction, ledgerType }: { transaction: Transaction; ledgerType: LEDGER_TYPE } =
		$props();

	const transactionDisplayType = transaction.type === 'credit' ? 'Credit' : 'Debit';
	const ledgerPath = ledgerType === 'budget' ? 'budgets' : 'accounts';
	const actions: Action[] = [
		{
			label: 'Delete',
			form: 'delete-transaction-form',
			className: 'btn-destructive',
			submit: true
		},
		{
			label: 'Edit',
			href: `/ledgers/${ledgerPath}/${transaction.ledgerId}/transactions/${transaction.id}/edit`,
			className: 'btn-primary'
		}
	];

	const formattedDate = formatDateToMonthYear(transaction.date, { includeDay: true });
</script>

<form method="POST" use:enhance id="delete-transaction-form">
	<input type="hidden" name="transactionId" value={transaction.id} />
</form>

<Card className="max-w-xl">
	{#snippet header()}
		<h2 class="card-title">Transaction Details</h2>
	{/snippet}
	<div class="space-y-1.5 text-sm text-zinc-800">
		<p>
			<span class="font-semibold">Date:</span>
			<span class="text-zinc-600">{formattedDate}</span>
		</p>
		<p>
			<span class="font-semibold">Type:</span>
			<span
				class="font-semibold {transaction.type === 'credit' ? 'text-green-700' : 'text-red-700'}"
				>{transactionDisplayType}</span
			>
		</p>
		<p>
			<span class="font-semibold">Category:</span>
			<span class="text-zinc-600 italic"
				>{m[`transactions.categories.${transaction.category}`]() ?? 'N/A'}</span
			>
		</p>
		<p class="flex items-center gap-2">
			<span class="font-semibold">Estiamted?:</span>
			<input type="checkbox" checked={transaction.isEstimated} readonly disabled />
		</p>
		<div>
			<span class="mb-0.5 block font-semibold">Description:</span>
			<div
				class="border-muted max-h-[96px] min-h-[48px] w-full overflow-y-auto rounded-xs border bg-neutral-100 p-2 leading-normal"
			>
				{transaction.description}
			</div>
		</div>
		<p class="mt-4">
			<span class="font-semibold">Amount:</span>
			<Amount value={transaction.amount} type={transaction.type} />
		</p>
	</div>
	{#snippet footer()}
		<div class="flex w-full items-center justify-end gap-2">
			{#each actions as action (action.label)}
				{#if action.href}
					<a href={action.href} class="{action.className} text-xs">{action.label}</a>
				{:else}
					<button
						type={action.submit ? 'submit' : 'button'}
						onclick={action.onClick}
						class="{action.className} text-xs"
						form={action.form}>{action.label}</button
					>
				{/if}
			{/each}
		</div>
	{/snippet}
</Card>

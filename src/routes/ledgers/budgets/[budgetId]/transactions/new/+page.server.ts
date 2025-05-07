import type { TRANSACTION_CATEGORIES, TRANSACTION_TYPES } from '$lib';
import { createTransaction } from '$lib/server/models/transactions';
import { fail, redirect } from '@sveltejs/kit';
import type { Actions } from './$types';

export const actions: Actions = {
	default: async (event) => {
		const formData = await event.request.formData();

		const budgetId = Number(event.params.budgetId);
		const amount = formData.get('amount') as string;
		const description = formData.get('description') as string;
		const category = formData.get('category') as Exclude<
			(typeof TRANSACTION_CATEGORIES)[number],
			'payable' | 'receivable'
		>;
		const date = new Date(formData.get('date') as string);
		const type = formData.get('type') as (typeof TRANSACTION_TYPES)[number];

		try {
			await createTransaction({
				ledgerId: budgetId,
				amount,
				description,
				category,
				date,
				type
			});
		} catch (error) {
			return fail(400, { message: error instanceof Error ? error.message : 'Unknown error' });
		}

		return redirect(302, `/ledgers/budgets/${budgetId}/transactions/new`);
	}
};

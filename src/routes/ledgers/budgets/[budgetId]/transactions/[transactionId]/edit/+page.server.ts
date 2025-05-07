import { fail, redirect } from '@sveltejs/kit';
import { updateTransaction } from '$lib/server/models/transactions';
import type { Actions } from './$types';
import type { TRANSACTION_CATEGORIES } from '$lib';

export const actions: Actions = {
	default: async (event) => {
		const formData = await event.request.formData();

		const { budgetId, transactionId } = event.params;

		const description = formData.get('description') as string;
		const amount = formData.get('amount') as string;
		const category = formData.get('category') as Exclude<
			(typeof TRANSACTION_CATEGORIES)[number],
			'payable' | 'receivable'
		>;
		const date = new Date(formData.get('date') as string);

		const params = {
			description,
			amount,
			category,
			date
		};

		try {
			await updateTransaction(Number(budgetId), Number(transactionId), params);
		} catch (error) {
			return fail(400, { message: error instanceof Error ? error.message : 'Unknown error' });
		}

		return redirect(302, `/ledgers/budgets/${budgetId}`);
	}
};

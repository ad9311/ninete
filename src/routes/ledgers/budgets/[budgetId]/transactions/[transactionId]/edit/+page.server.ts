import { fail, redirect } from '@sveltejs/kit';
import { updateTransaction } from '$lib/server/models/transactions';
import type { Actions } from './$types';
import { formatFormErrors, type TRANSACTION_CATEGORIES } from '$lib/shared';
import type { ZodError } from 'zod';

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
		const isEstimated = formData.get('is_estimated');

		const params = {
			description,
			amount,
			category,
			date,
			isEstimated: isEstimated === 'on'
		};

		try {
			await updateTransaction(Number(budgetId), 'budget', Number(transactionId), params);
		} catch (e) {
			const errors = formatFormErrors(e as Error | ZodError);
			return fail(400, { errors });
		}

		redirect(303, `/ledgers/budgets/${budgetId}`);
	}
};

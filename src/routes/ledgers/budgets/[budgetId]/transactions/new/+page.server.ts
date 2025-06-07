import { formatFormErrors, type TRANSACTION_CATEGORIES, type TRANSACTION_TYPES } from '$lib/shared';
import { createTransaction } from '$lib/server/models/transactions';
import { fail } from '@sveltejs/kit';
import type { Actions } from './$types';
import type { ZodError } from 'zod';

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
		const isEstimated = formData.get('is_estimated');

		try {
			await createTransaction({
				ledgerId: budgetId,
				amount,
				description,
				category,
				date,
				type,
				isEstimated: isEstimated === 'on'
			});
		} catch (e) {
			const errors = formatFormErrors(e as Error | ZodError);
			return fail(400, { errors });
		}

		return { success: true };
	}
};

import { formatFormErrors, type TRANSACTION_CATEGORY, type TRANSACTION_TYPE } from '$lib/shared';
import { createTransaction } from '$lib/server/models/transactions';
import { fail } from '@sveltejs/kit';
import type { Actions } from './$types';
import type { ZodError } from 'zod';

export const actions: Actions = {
	default: async (event) => {
		const formData = await event.request.formData();

		const loanId = Number(event.params.loanId);
		const amount = formData.get('amount') as string;
		const description = formData.get('description') as string;
		const category = formData.get('category') as TRANSACTION_CATEGORY;
		const date = new Date(formData.get('date') as string);
		const type = formData.get('type') as TRANSACTION_TYPE;
		const isEstimated = formData.get('is_estimated');

		try {
			await createTransaction('loan', {
				ledgerId: loanId,
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

import { formatFormErrors } from '$lib/shared';
import type { ZodError } from 'zod';
import type { Actions } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { createLoan, loanNewSchema, type NewLoanParams } from '$lib/server/models/ledger/loans';
import type { Ledger } from '$lib/server/db/schema';

export const actions: Actions = {
	default: async (event) => {
		const formData = await event.request.formData();
		const userId = Number(event.locals.user?.id);

		const title = formData.get('title') as string;
		const description = formData.get('description') as string;
		const date = new Date(formData.get('date') as string);

		let loan: Ledger;

		try {
			const params: NewLoanParams = {
				userId,
				date,
				title,
				description,
				type: 'loan',
				status: 'pending'
			};
			const validated = loanNewSchema.parse(params);
			loan = await createLoan(validated);
		} catch (e) {
			const errors = formatFormErrors(e as Error | ZodError);

			return fail(400, { errors });
		}

		redirect(303, `/ledgers/loans/${loan.id}`);
	}
};

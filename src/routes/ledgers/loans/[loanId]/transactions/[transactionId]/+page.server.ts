import { deleteTransaction } from '$lib/server/models/transactions';
import { fail, redirect, type Actions } from '@sveltejs/kit';

export const actions: Actions = {
	default: async (event) => {
		const { transactionId, loanId } = event.params;

		try {
			await deleteTransaction(Number(loanId), 'loan', Number(transactionId));
		} catch (error) {
			return fail(400, { message: error instanceof Error ? error.message : 'Unknown error' });
		}

		redirect(303, `/ledgers/loans/${loanId}`);
	}
};

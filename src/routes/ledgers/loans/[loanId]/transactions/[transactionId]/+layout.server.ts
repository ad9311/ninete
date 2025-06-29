import { findTransactionWithLedgerId } from '$lib/server/models/transactions';
import { error } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async (event) => {
	const { transactionId, loanId } = event.params;

	const transaction = await findTransactionWithLedgerId(Number(loanId), Number(transactionId));

	if (!transaction) {
		error(404, 'Transaction not found');
	}

	return { transaction };
};

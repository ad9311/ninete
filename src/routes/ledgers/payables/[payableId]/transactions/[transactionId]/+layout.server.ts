import { findTransaction } from '$lib/server/models/transactions';
import { error } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async (event) => {
	const { transactionId } = event.params;

	const transaction = await findTransaction(Number(transactionId));

	if (!transaction) {
		error(404, 'Transaction not found');
	}

	return { transaction };
};

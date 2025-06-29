import { findLedger, findLedgetCredtis, findLedgetDebits } from '$lib/server/models/ledger';
import { error } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async (event) => {
	const { user } = event.locals;
	const { loanId } = event.params;

	const loan = await findLedger(Number(user?.id), Number(loanId), 'loan');

	if (!loan) {
		error(404, 'loan not found!');
	}

	const credits = await findLedgetCredtis(Number(loanId));
	const debits = await findLedgetDebits(Number(loanId));

	return { loan, credits, debits };
};

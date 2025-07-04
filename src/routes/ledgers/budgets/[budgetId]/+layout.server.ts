import { error } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';
import { findLedger, findLedgetCredtis } from '$lib/server/models/ledger';
import { findLedgetDebits } from '$lib/server/models/ledger';

export const load: LayoutServerLoad = async (event) => {
	const { user } = event.locals;
	const { budgetId } = event.params;

	const budget = await findLedger(Number(user?.id), Number(budgetId), 'budget');

	if (!budget) {
		error(404, 'Budget not found');
	}

	const credits = await findLedgetCredtis(Number(budgetId));
	const debits = await findLedgetDebits(Number(budgetId));

	return { budget, credits, debits };
};

import { findUserBudget } from '$lib/server/models/ledger/budget';
import { error } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';
import { findLedgetCredtis } from '$lib/server/models/ledger';
import { findLedgetDebits } from '$lib/server/models/ledger';

export const load: LayoutServerLoad = async (event) => {
	const { user } = event.locals;
	const { budgetId } = event.params;

	const budget = await findUserBudget(Number(user?.id), Number(budgetId));

	if (!budget) {
		return error(404, 'Budget not found');
	}

	const credits = await findLedgetCredtis(Number(budgetId));
	const debits = await findLedgetDebits(Number(budgetId));

	return { budget, credits, debits };
};

import { findLedger, findLedgetCredtis, findLedgetDebits } from '$lib/server/models/ledger';
import { error } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async (event) => {
	const { user } = event.locals;
	const { payableId } = event.params;

	const payable = await findLedger(Number(user?.id), Number(payableId), 'payable');

	if (!payable) {
		error(404, 'payable not found!');
	}

	const credits = await findLedgetCredtis(Number(payableId));
	const debits = await findLedgetDebits(Number(payableId));

	return { payable, credits, debits };
};

import { findLedgers } from '$lib/server/models/ledger';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const payables = await findLedgers(Number(event.locals.user?.id), 'payable');

	return { payables };
};

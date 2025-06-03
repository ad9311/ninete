import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { findOrCreateBudget } from '$lib/server/models/ledger/budget';

export const load: PageServerLoad = async (event) => {
	const { user } = event.locals;

	if (!user) {
		redirect(303, '/login');
	}

	const currentBudget = await findOrCreateBudget(user.id);

	return {
		currentBudget
	};
};

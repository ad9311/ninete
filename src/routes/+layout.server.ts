import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async (event) => {
	return {
		isUserSignedIn: !!event.locals?.session && !!event.locals?.user,
		currentUser: event.locals?.user
	};
};

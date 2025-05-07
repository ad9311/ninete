import { fail, redirect } from '@sveltejs/kit';
import type { Actions } from './$types';
import { createUser } from '$lib/server/models/user';
import * as auth from '$lib/server/auth';

export const actions: Actions = {
	default: async (event) => {
		const formData = await event.request.formData();

		const username = formData.get('username') as string;
		const email = formData.get('email') as string;
		const password = formData.get('password') as string;
		const passwordConfirmation = formData.get('password-confirmation') as string;

		if (password !== passwordConfirmation) {
			return fail(400, { message: 'Passwords do not match' });
		}

		try {
			const params = { username, email, password };
			const user = await createUser(params);

			const sessionToken = auth.generateSessionToken();
			const session = await auth.createSession(sessionToken, user.id);
			auth.setSessionTokenCookie(event, sessionToken, session.expiresAt);
		} catch (error) {
			const msg = (error as Error).message;
			return fail(400, { message: `Failed to create user: ${msg}` });
		}

		return redirect(302, '/home');
	}
};

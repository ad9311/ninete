import { fail, redirect } from '@sveltejs/kit';
import * as auth from '$lib/server/auth';
import { findUserByEmail, verfifyPassword } from '$lib/server/models/user';
import type { Actions } from './$types';

export const actions: Actions = {
	default: async (event) => {
		const formData = await event.request.formData();

		const email = formData.get('email') as string;
		const password = formData.get('password') as string;

		try {
			const existingUser = await findUserByEmail(email);

			if (!existingUser) {
				return fail(400, { message: 'Incorrect email or password.' });
			}

			const validPassword = await verfifyPassword(existingUser.passwordHash, password);

			if (!validPassword) {
				return fail(400, { message: 'Incorrect email or password.' });
			}

			const sessionToken = auth.generateSessionToken();
			const session = await auth.createSession(sessionToken, existingUser.id);
			auth.setSessionTokenCookie(event, sessionToken, session.expiresAt);
		} catch (e) {
			console.error('Sign-in error:', e);
			return fail(400, { message: 'An unexpected error occurred during sign-in.' });
		}

		return redirect(302, '/');
	}
};

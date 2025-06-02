import { fail, redirect } from '@sveltejs/kit';
import * as auth from '$lib/server/auth';
import { findUserByEmail, verfifyPassword } from '$lib/server/models/user';
import type { Actions } from './$types';
import type { ZodError } from 'zod';
import { formatFormErrors } from '$lib/shared';

const INCORRECT_EMAIL_PASSWORD = 'Incorrect email or password';

export const actions: Actions = {
	default: async (event) => {
		const formData = await event.request.formData();

		const email = formData.get('email') as string;
		const password = formData.get('password') as string;

		try {
			const existingUser = await findUserByEmail(email);

			if (!existingUser) {
				throw new Error(INCORRECT_EMAIL_PASSWORD);
			}

			const validPassword = await verfifyPassword(existingUser.passwordHash, password);

			if (!validPassword) {
				throw new Error(INCORRECT_EMAIL_PASSWORD);
			}

			const sessionToken = auth.generateSessionToken();
			const session = await auth.createSession(sessionToken, existingUser.id);
			auth.setSessionTokenCookie(event, sessionToken, session.expiresAt);
		} catch (e) {
			const errors = formatFormErrors(e as Error | ZodError);
			return fail(400, { errors });
		}

		return redirect(302, '/');
	}
};

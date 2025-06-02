import { fail, redirect } from '@sveltejs/kit';
import type { Actions } from './$types';
import { createUser } from '$lib/server/models/user';
import * as auth from '$lib/server/auth';
import type { ZodError } from 'zod';
import { formatFormErrors } from '$lib/shared';

export const actions: Actions = {
	default: async (event) => {
		const formData = await event.request.formData();

		const username = formData.get('username') as string;
		const email = formData.get('email') as string;
		const password = formData.get('password') as string;
		const passwordConfirmation = formData.get('password-confirmation') as string;

		if (password !== passwordConfirmation) {
			throw new Error('Passwords do not match');
		}

		try {
			const params = { username, email, password };
			const user = await createUser(params);

			const sessionToken = auth.generateSessionToken();
			const session = await auth.createSession(sessionToken, user.id);
			auth.setSessionTokenCookie(event, sessionToken, session.expiresAt);
		} catch (e) {
			const errors = formatFormErrors(e as Error | ZodError);
			return fail(400, { errors });
		}

		return redirect(302, '/home');
	}
};

import { redirect, fail } from '@sveltejs/kit';
import * as auth from '$lib/server/auth'; // Ensure this path is correct
import type { Actions } from './$types';
import { sha256 } from '@oslojs/crypto/sha2'; // Needed to hash the token for invalidation
import { encodeHexLowerCase } from '@oslojs/encoding'; // Needed for encoding the hash

export const load = async (event) => {
	const { request } = event;

	if (request.method !== 'POST') {
		return redirect(303, '/home');
	}
};

export const actions: Actions = {
	default: async (event) => {
		const sessionToken = event.cookies.get(auth.sessionCookieName);

		if (!sessionToken) {
			// This case should ideally be caught by the load function or hooks,
			// but it's good for robustness.
			return fail(401, { message: 'Not authenticated. No session token found.' });
		}

		try {
			// Derive the sessionId (hashed token) from the raw sessionToken
			// This matches how sessionId is generated in your auth.ts (e.g., in createSession or validateSessionToken)
			const sessionId = encodeHexLowerCase(sha256(new TextEncoder().encode(sessionToken)));

			// Invalidate the session in the database using the derived sessionId
			await auth.invalidateSession(sessionId);

			// Delete the session cookie from the browser
			auth.deleteSessionTokenCookie(event);
		} catch (e) {
			console.error('Sign-out error:', e);
			// Provide a generic error message to the user
			return fail(500, {
				message: 'An error occurred while trying to sign out. Please try again.'
			});
		}

		// Redirect to the sign-in page after successful sign-out
		throw redirect(303, '/auth/sign-in');
	}
};

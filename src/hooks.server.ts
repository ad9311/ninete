import { sequence } from '@sveltejs/kit/hooks';
import * as auth from '$lib/server/auth.js';
import type { Handle } from '@sveltejs/kit';
import { paraglideMiddleware } from '$lib/paraglide/server';
import { redirect } from '@sveltejs/kit';

const handleParaglide: Handle = ({ event, resolve }) =>
	paraglideMiddleware(event.request, ({ request, locale }) => {
		event.request = request;

		return resolve(event, {
			transformPageChunk: ({ html }) => html.replace('%paraglide.lang%', locale)
		});
	});

const handleAuth: Handle = async ({ event, resolve }) => {
	const sessionToken = event.cookies.get(auth.sessionCookieName);
	const { pathname } = event.url;

	const unprotectedRoutes = ['/auth/sign-in', '/auth/sign-up'];
	const isUnprotectedRoute = unprotectedRoutes.includes(pathname);

	if (!sessionToken) {
		event.locals.user = null;
		event.locals.session = null;
		if (!isUnprotectedRoute) {
			throw redirect(303, '/auth/sign-in');
		}
		return resolve(event);
	}

	const { session, user } = await auth.validateSessionToken(sessionToken);

	if (session) {
		auth.setSessionTokenCookie(event, sessionToken, session.expiresAt);
	} else {
		// Session is invalid or expired
		auth.deleteSessionTokenCookie(event);
		event.locals.user = null;
		event.locals.session = null;
		if (!isUnprotectedRoute) {
			throw redirect(303, '/auth/sign-in');
		}
		return resolve(event);
	}

	event.locals.user = user;
	event.locals.session = session;

	// If user is logged in and tries to access auth pages, redirect to home
	if (user && isUnprotectedRoute) {
		throw redirect(303, '/');
	}

	return resolve(event);
};

export const handle: Handle = sequence(handleParaglide, handleAuth);

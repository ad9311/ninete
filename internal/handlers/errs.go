package handlers

import "errors"

var (
	ErrParseForm = errors.New("failed to parse form")

	// ErrLoginUnavailable is what the browser sees when the account lookup
	// itself failed. The underlying error goes to the log instead, so a
	// database fault does not describe itself on a public page.
	ErrLoginUnavailable = errors.New("login is temporarily unavailable, please try again")

	// ErrRegistrationUnavailable is the sign-up counterpart of
	// ErrLoginUnavailable: the applicant caused nothing, so they get a generic
	// message and the real error goes to the log.
	ErrRegistrationUnavailable = errors.New("registration is temporarily unavailable, please try again")

	ErrTooManyAttempts = errors.New("too many attempts, please wait a moment and try again")

	// API errors. Their messages reach the browser as a JSON body, so they say
	// what the client can act on and nothing about the server.
	ErrUnauthorized     = errors.New("authentication required")
	ErrInvalidCSRFToken = errors.New("invalid or missing CSRF token")
	ErrAPIRouteNotFound = errors.New("resource not found")
	ErrAPIUnavailable   = errors.New("the request could not be completed, please try again")
	ErrAPIInvalidJSON   = errors.New("request body must be valid JSON")

	ErrSearchDateFormat    = errors.New("dates must use the YYYY-MM-DD format")
	ErrSearchDateRange     = errors.New("the from date must be on or before the to date")
	ErrUnknownDateRange    = errors.New("unknown date range")
	ErrBudgetCategoryField = errors.New("invalid budget field name")
	ErrSearchTermTooLong   = errors.New("search terms must be at most 50 characters")

	// ErrAPIInvalidDateRange is the /api/expenses* twin of ErrUnknownDateRange:
	// the client resolves a named range to explicit bounds itself (§3.6 of
	// docs/spa-migration.md), so the API only ever sees start/end and rejects
	// anything that is not a well-formed half-open range.
	ErrAPIInvalidDateRange = errors.New("start and end must both be set, with start before end")
)

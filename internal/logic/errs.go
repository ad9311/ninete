package logic

import "errors"

var (
	ErrWithPasswords         = errors.New("failed to save passwords")
	ErrWrongEmailOrPassword  = errors.New("wrong email or password")
	ErrInvalidInvitationCode = errors.New("invalid invitation code")
	ErrInvitationCodeExists  = errors.New("invitation code already exists")
	ErrPasswordConfirmation  = errors.New("password and password confirmation do not match")
	ErrInvitationCodeVerify  = errors.New("failed to verify invitation code")
	ErrLoginLookup           = errors.New("failed to look up account")

	// ErrAccountExists names neither the field that collided nor the value, so
	// a holder of a valid invitation code cannot probe which addresses are
	// already registered.
	//
	// It is also the shape to copy for any future create endpoint that does a
	// plain insert on a unique column: catch repo.IsUniqueViolation, return a
	// sentinel from this file, and name that sentinel in the endpoint's
	// userErrors list so WriteAPIError answers 422. Left uncaught, SQLite's
	// "UNIQUE constraint failed: ..." reaches WriteAPIError, which does not
	// recognize the string and falls through to a generic 500 — right about not
	// leaking the driver's message, wrong about whose fault the request was.
	//
	// SignUp is the only live path that needs it, and deliberately the only one:
	// tags are created with INSERT OR IGNORE (ensureTagsForUserTx), expense
	// budgets upsert (repo/expense_budget.go), and categories have no create
	// route at all. Do not lift this into a shared helper — reject-and-report,
	// silently-reuse and upsert are three different conflict semantics, not one
	// rule written three times.
	ErrAccountExists = errors.New("an account with that username or email already exists")

	// ErrSignUpFailed marks a sign-up that failed for a reason the applicant
	// did nothing to cause. The underlying error is for the log, not the page.
	ErrSignUpFailed = errors.New("failed to create account")

	ErrValidationAssertion = errors.New("failed to assert error type")
	ErrValidationFailed    = errors.New("validation failed")

	ErrTagResolutionFailed = errors.New("failed to resolve tags")

	ErrReportTooManyTags = errors.New("too many grouping tags, 20 maximum")

	ErrQuickExpenseFormat      = errors.New("quick expense must be: description, amount, month[, tags]")
	ErrQuickExpenseDescription = errors.New("description must be between 3 and 50 characters")
	ErrQuickExpenseAmount      = errors.New("invalid amount")
	ErrQuickExpenseDate        = errors.New("invalid month, use last, current, next or Jul 2026")
	ErrQuickExpenseTags        = errors.New("too many tags, 10 maximum")
	ErrQuickExpenseTagName     = errors.New("each tag must be at most 20 characters")
)

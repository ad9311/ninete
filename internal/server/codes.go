package server

// Code represents a response code used in API responses.
type Code string

const (
	successCode = Code("SUCCESS") // Indicates a successful response

	// Error codes
	standardErrorCode          = Code("ERROR")                              // Generic error
	internalErrorCode          = Code("INTERNAL_ERROR")                     // Internal server error
	invalidFormFormatErrorCode = Code("INVALID_FORM_FORMAT")                // Invalid form format
	invalidFormErrorCode       = Code("INVALID_FORM")                       // Invalid form data
	invalidAuthCredsErrorCode  = Code("INVALID_AUTHENTICATION_CREDENTIALS") // Invalid authentication credentials
	routeNotFound              = Code("PATH_NOT_FOUND")                     // Route not found
	methodNotAllowed           = Code("METHOD_NOT_ALLOWED")                 // Method not allowed
)

// Package app provides application-wide configuration, error definitions, and utility/helper functions.
package app

// Context keys used for storing and retrieving values in context.Context.
const (
	CurrentUserIDKey = ContextKey("currentuserID") // Key for storing the current user's ID in context
	CurrentUserKey   = ContextKey("currentUser")   // Key for storing the current user object in context
)

// ContextKey is a dedicated type for context keys to avoid collisions in context.Context.
type ContextKey string

// SetDefaultString returns the default string d if v is empty; otherwise, it returns v.
func SetDefaultString(v, d string) string {
	if v == "" {
		return d
	}

	return v
}

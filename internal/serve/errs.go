package serve

import "errors"

var (
	ErrLayoutNotFound  = errors.New("layout template not found")
	ErrNonceGeneration = errors.New("failed to generate csp nonce")
)

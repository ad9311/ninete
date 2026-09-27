package config

import "errors"

var (
	ErrMissingURL   = errors.New(EnvURL + " is required, e.g. https://ninete.example.com")
	ErrInvalidURL   = errors.New(EnvURL + " must be an absolute http(s) URL with no path, query or fragment")
	ErrInsecureURL  = errors.New(EnvURL + " must use https unless it points at this machine")
	ErrMissingToken = errors.New(EnvToken + " is required; create one at /account/tokens")
	ErrInvalidToken = errors.New(EnvToken + ` does not look like a Ninete token (it starts with "nin_")`)
	ErrInvalidTZ    = errors.New(EnvTZ + ` must be an IANA zone name, e.g. "America/Bogota"`)
)

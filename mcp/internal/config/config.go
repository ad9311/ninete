// Package config reads the MCP server's settings from its environment. The
// MCP client (Claude Code, Claude Desktop) launches the binary and passes them
// in its server configuration; nothing is read from a file.
package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

const (
	EnvURL   = "NINETE_URL"
	EnvToken = "NINETE_TOKEN" //nolint:gosec // G101: the variable name, not a credential
	EnvTZ    = "NINETE_TZ"

	// tokenMarker opens every Ninete token (logic.apiTokenMarker on the
	// server). Checked here so a pasted session cookie or a stray value fails
	// at start-up with a message, not as a 401 on the first call.
	tokenMarker = "nin_"
)

var (
	ErrMissingURL   = errors.New(EnvURL + " is required, e.g. https://ninete.example.com")
	ErrInvalidURL   = errors.New(EnvURL + " must be an absolute http(s) URL with no path, query or fragment")
	ErrInsecureURL  = errors.New(EnvURL + " must use https unless it points at this machine")
	ErrMissingToken = errors.New(EnvToken + " is required; create one at /account/tokens")
	ErrInvalidToken = errors.New(EnvToken + ` does not look like a Ninete token (it starts with "nin_")`)
	ErrInvalidTZ    = errors.New(EnvTZ + ` must be an IANA zone name, e.g. "America/Bogota"`)
)

type Config struct {
	// BaseURL is the app's origin, with no trailing slash.
	BaseURL string
	Token   string
	// Location decides what "today" and "this month" mean, and which window of
	// instants a created-on day covers. It plays the role the browser's zone
	// plays for the web app.
	Location *time.Location
}

// Load builds a Config from getenv. It takes the lookup as a function so tests
// never touch the real environment.
func Load(getenv func(string) string) (Config, error) {
	var cfg Config

	baseURL, err := parseBaseURL(getenv(EnvURL))
	if err != nil {
		return cfg, err
	}

	token := strings.TrimSpace(getenv(EnvToken))
	switch {
	case token == "":
		return cfg, ErrMissingToken
	case !strings.HasPrefix(token, tokenMarker):
		return cfg, ErrInvalidToken
	}

	loc := time.Local
	if name := strings.TrimSpace(getenv(EnvTZ)); name != "" {
		loc, err = time.LoadLocation(name)
		if err != nil {
			return cfg, fmt.Errorf("%w: %w", ErrInvalidTZ, err)
		}
	}

	return Config{BaseURL: baseURL, Token: token, Location: loc}, nil
}

// parseBaseURL insists on https. The token is a bearer credential, so sending
// it in clear text hands it to anyone on the path. Plain http is accepted only
// for a loopback host, which is how the app runs in development.
func parseBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrMissingURL
	}

	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", ErrInvalidURL
	}

	if strings.Trim(u.Path, "/") != "" {
		return "", ErrInvalidURL
	}

	switch u.Scheme {
	case "https":
	case "http":
		if !isLoopback(u.Hostname()) {
			return "", ErrInsecureURL
		}
	default:
		return "", ErrInvalidURL
	}

	return u.Scheme + "://" + u.Host, nil
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}

	ip := net.ParseIP(host)

	return ip != nil && ip.IsLoopback()
}

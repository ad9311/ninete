// Package api is a small HTTP client for Ninete's /api/* JSON API, authenticated
// with a personal access token.
//
// The response types below are this module's own copies of the server's JSON
// shapes, on purpose: this module is outside github.com/ad9311/ninete, so it
// cannot import internal/handlers, and it should not want to. The API is the
// contract; a change there has to be made here too, visibly.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	apiPrefix      = "/api"
	requestTimeout = 15 * time.Second
	// maxResponseBytes bounds what one response may make this process hold.
	// The largest real answer is a 100-row expense page.
	maxResponseBytes = 4 << 20
)

// Error is a failure the API answered with its JSON envelope
// ({"error", "fields"}). Fields is set on a 422 and maps each rejected request
// field to the validation rule it broke ("required", "max", ...).
type Error struct {
	Status  int
	Message string
	Fields  map[string]string
}

func (e *Error) Error() string {
	if len(e.Fields) == 0 {
		return e.Message
	}

	parts := make([]string, 0, len(e.Fields))
	for field, rule := range e.Fields {
		parts = append(parts, field+": "+rule)
	}

	return fmt.Sprintf("%s (%s)", e.Message, strings.Join(parts, ", "))
}

type Client struct {
	baseURL   string
	token     string
	userAgent string
	http      *http.Client
}

func New(baseURL, token, version string) *Client {
	return &Client{
		baseURL:   strings.TrimSuffix(baseURL, "/"),
		token:     token,
		userAgent: "ninete-mcp/" + version,
		http: &http.Client{
			Timeout: requestTimeout,
			// A redirect from /api/* means the URL points at something that is
			// not the API — most likely the login page. Following it would
			// forward the Authorization header's request to wherever it leads
			// and hand back HTML; stopping surfaces the misconfiguration.
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// Get, Post and Put send one request to apiPrefix+path. out may be nil for a
// response with no body. query is nil or one of the query structs in query.go,
// encoded by EncodeQuery.
func (c *Client) Get(ctx context.Context, path string, query, out any) error {
	var values url.Values

	if query != nil {
		encoded, err := EncodeQuery(query)
		if err != nil {
			return err
		}

		values = encoded
	}

	return c.do(ctx, http.MethodGet, path, values, nil, out)
}

func (c *Client) Post(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPost, path, nil, body, out)
}

func (c *Client) Put(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPut, path, nil, body, out)
}

// There is deliberately no Delete. The server refuses DELETE to every token
// scope, and this client does not offer the method at all.

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	// A query written into the path ("/expenses?tags=food") would bypass the
	// query structs, so contracttest would never see its keys. Every query must
	// come through Get's struct; a "#" would silently cut the path short.
	if strings.ContainsAny(path, "?#") {
		return fmt.Errorf("%w: %q", ErrPathQuery, path)
	}

	target := c.baseURL + apiPrefix + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}

		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, target, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.http.Do(req)
	if err != nil {
		// The error names the URL but never the header, so the token cannot
		// leak through it.
		return fmt.Errorf("request to Ninete failed: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if res.StatusCode >= http.StatusMultipleChoices {
		return responseError(res, data)
	}

	if out == nil || res.StatusCode == http.StatusNoContent {
		return nil
	}

	if !isJSON(res.Header.Get("Content-Type")) {
		return fmt.Errorf("%w: expected JSON, got %q; is %s the right URL?",
			ErrUnexpected, res.Header.Get("Content-Type"), c.baseURL)
	}

	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("%w: %w", ErrUnexpected, err)
	}

	return nil
}

func responseError(res *http.Response, data []byte) error {
	switch res.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	case http.StatusNotFound:
		return ErrNotFound
	}

	if res.StatusCode < http.StatusBadRequest {
		return fmt.Errorf("%w: redirected (%d) to %q; is the URL right?",
			ErrUnexpected, res.StatusCode, res.Header.Get("Location"))
	}

	var envelope struct {
		Error  string            `json:"error"`
		Fields map[string]string `json:"fields"`
	}

	if isJSON(res.Header.Get("Content-Type")) && json.Unmarshal(data, &envelope) == nil && envelope.Error != "" {
		return &Error{Status: res.StatusCode, Message: envelope.Error, Fields: envelope.Fields}
	}

	return fmt.Errorf("%w: status %d", ErrUnexpected, res.StatusCode)
}

func isJSON(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)

	return err == nil && mediaType == "application/json"
}

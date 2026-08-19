package dbaas

import "net/http"

// Option configures a Client at construction time.
type Option func(*Client)

// WithHTTPClient overrides the default HTTP client - useful for setting a
// custom timeout, transport, or injecting a mock in tests.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithToken sets a static bearer token.
func WithToken(token string) Option {
	return func(c *Client) {
		c.token = token
	}
}

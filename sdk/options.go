package dbaas

import (
	"net/http"
	"time"
)

// Option configures a Client at construction time.
type Option func(*Client)

// WithHTTPClient overrides the default HTTP client - useful for setting a
// custom timeout or transport (proxies, tracing, test doubles). Token
// requests use it too.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithPollInterval sets how often InstancesService.WaitUntilRunning checks
// the instance's status. The default is 5 seconds.
func WithPollInterval(d time.Duration) Option {
	return func(c *Client) {
		c.pollInterval = d
	}
}

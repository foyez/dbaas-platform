package dbaas

import (
	"context"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	defaultTimeout      = 30 * time.Second
	defaultPollInterval = 5 * time.Second
)

// Client is the entry point for interacting with the DBaaS Platform API.
// It is safe for concurrent use by multiple goroutines.
type Client struct {
	baseURL    string
	httpClient *http.Client

	// tokenSource supplies the bearer token for each request; nil sends
	// requests unauthenticated (only /health allows that).
	tokenSource oauth2.TokenSource
	// newTokenSource builds tokenSource once all options are applied, so
	// it can fetch tokens with the final httpClient whatever the order of
	// the options.
	newTokenSource func(*http.Client) oauth2.TokenSource

	pollInterval time.Duration

	Instances *InstancesService
}

// New creates a Client. baseURL is required; everything else has sane
// defaults and can be overridden with Options.
func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		pollInterval: defaultPollInterval,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.newTokenSource != nil {
		c.tokenSource = c.newTokenSource(c.httpClient)
	}

	c.Instances = &InstancesService{client: c}

	return c
}

// tokenContext is the context the token source fetches tokens with. Token
// fetches outlive a single API call (the token is cached and reused), so it
// is not tied to any request's context.
func tokenContext(hc *http.Client) context.Context {
	return context.WithValue(context.Background(), oauth2.HTTPClient, hc)
}

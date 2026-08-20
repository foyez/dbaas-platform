package dbaas

import (
	"net/http"
	"time"
)

// Client is the entry point for interacting with the DBaaS Platform API.
type Client struct {
	baseURL    string
	httpClient *http.Client

	Instances *InstancesService
}

// New creates a Client. baseURL is required; everything else has sane
// defaults and can be overridden with Options.
func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	c.Instances = &InstancesService{client: c}

	return c
}

package dbaas_test

import (
	"errors"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/foyez/dbaas-platform/sdk/dbaas"
	"github.com/foyez/dbaas-platform/sdk/dbaas/dbaastest"
)

func TestClientCredentials(t *testing.T) {
	srv := dbaastest.NewServer()
	defer srv.Close()

	// Count token requests through the custom HTTP client: WithHTTPClient
	// comes after WithClientCredentials, and must still be used for tokens.
	var tokenRequests atomic.Int32
	hc := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/oauth/v2/token" {
			tokenRequests.Add(1)
		}
		return http.DefaultTransport.RoundTrip(r)
	})}

	c := dbaas.New(srv.URL,
		dbaas.WithClientCredentials(srv.TokenURL, dbaastest.ClientID, dbaastest.ClientSecret, nil),
		dbaas.WithHTTPClient(hc),
	)

	for range 3 {
		if _, err := c.Instances.List(t.Context(), dbaas.ListOptions{}); err != nil {
			t.Fatal(err)
		}
	}

	if n := tokenRequests.Load(); n != 1 {
		t.Errorf("token requests = %d, want 1 (the token is cached)", n)
	}
}

func TestClientCredentialsWrongSecret(t *testing.T) {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL, dbaas.WithClientCredentials(srv.TokenURL, dbaastest.ClientID, "wrong", nil))

	_, err := c.Instances.List(t.Context(), dbaas.ListOptions{})

	if err == nil {
		t.Fatal("want an error for a wrong client secret")
	}
	if errors.Is(err, dbaas.ErrUnauthorized) {
		t.Error("a failed token fetch is not an API error")
	}
}

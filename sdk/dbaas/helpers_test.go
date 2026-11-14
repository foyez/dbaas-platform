package dbaas_test

// Helpers shared by the unit tests. Unit tests sit next to the file they
// test (request_test.go tests request.go, ...); the examples, which double
// as documentation, are in example_test.go.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// recordServer replies with status and body to every request, and records
// the last request it got.
func recordServer(t *testing.T, status int, body string) (*httptest.Server, *http.Request) {
	t.Helper()

	var got http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = *r.Clone(context.Background())
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	return srv, &got
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

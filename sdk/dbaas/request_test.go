package dbaas_test

import (
	"net/http"
	"testing"

	"github.com/foyez/dbaas-platform/sdk/dbaas"
)

func TestCreateSendsHeaders(t *testing.T) {
	srv, got := recordServer(t, http.StatusAccepted, `{"id":"instance-1"}`)
	c := dbaas.New(srv.URL+"/", dbaas.WithToken("tok"))

	_, err := c.Instances.Create(t.Context(), dbaas.CreateInstanceInput{
		Name: "orders", Version: 16, Storage: "1Gi", IdempotencyKey: "key-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	checks := map[string]string{
		"Authorization":   "Bearer tok",
		"Idempotency-Key": "key-1",
		"Content-Type":    "application/json",
		"Accept":          "application/json",
		"User-Agent":      "dbaas-go/" + dbaas.Version,
	}
	for header, want := range checks {
		if v := got.Header.Get(header); v != want {
			t.Errorf("%s = %q, want %q", header, v, want)
		}
	}
	if got.URL.Path != "/api/v1/instances" {
		t.Errorf("path = %q, want /api/v1/instances (trailing slash of base URL trimmed)", got.URL.Path)
	}
}

func TestIDIsPathEscaped(t *testing.T) {
	srv, got := recordServer(t, http.StatusOK, `{}`)
	c := dbaas.New(srv.URL)

	_, _ = c.Instances.Get(t.Context(), "../audit-logs")

	if want := "/api/v1/instances/..%2Faudit-logs"; got.URL.RawPath != want {
		t.Errorf("raw path = %q, want %q", got.URL.RawPath, want)
	}
}

func TestListSendsPageQuery(t *testing.T) {
	srv, got := recordServer(t, http.StatusOK, `{"items":[]}`)
	c := dbaas.New(srv.URL)

	_, err := c.Instances.List(t.Context(), dbaas.ListOptions{Limit: 5, Cursor: "abc"})
	if err != nil {
		t.Fatal(err)
	}

	if q := got.URL.RawQuery; q != "cursor=abc&limit=5" {
		t.Errorf("query = %q, want cursor=abc&limit=5", q)
	}
}

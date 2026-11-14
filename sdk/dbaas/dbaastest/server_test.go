package dbaastest_test

// The fake is only useful while it behaves like the real API. These tests
// pin the parts of the contract the SDK relies on, in raw HTTP, so a change
// to the fake that drifts from the API fails here and not in a confusing
// example. integration_test.go checks the same behaviour on the real API.

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/foyez/dbaas-platform/sdk/dbaas/dbaastest"
)

func send(t *testing.T, srv *dbaastest.Server, method, path, token, key, body string) (int, string) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), method, srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func TestContract(t *testing.T) {
	const create = `{"name":"orders","version":16,"storage":"1Gi"}`

	tests := []struct {
		name                     string
		method, path, token, key string
		body                     string
		wantStatus               int
		wantBody                 string
	}{
		{"health needs no token", "GET", "/health", "", "", "", 200, `"status":"ok"`},
		{"no token is a plain-text 401", "GET", "/api/v1/instances", "", "", "", 401, "unauthorized"},
		{"create is 202 Pending", "POST", "/api/v1/instances", dbaastest.Token, "k", create, 202, `"status":"Pending"`},
		{"create needs a key", "POST", "/api/v1/instances", dbaastest.Token, "", create, 400, "MISSING_IDEMPOTENCY_KEY"},
		{"create validates", "POST", "/api/v1/instances", dbaastest.Token, "k", `{"name":"x"}`, 400, "INVALID_REQUEST"},
		{"list is an items envelope", "GET", "/api/v1/instances", dbaastest.Token, "", "", 200, `"items":[]`},
		{"unknown instance is 404", "GET", "/api/v1/instances/nope", dbaastest.Token, "", "", 404, "NOT_FOUND"},
		{"update needs a field", "PATCH", "/api/v1/instances/nope", dbaastest.Token, "", `{}`, 400, "NO_UPDATE_FIELDS"},
		{"audit logs need admin", "GET", "/api/v1/audit-logs", dbaastest.Token, "", "", 403, "forbidden"},
		{"admin reads audit logs", "GET", "/api/v1/audit-logs", dbaastest.AdminToken, "", "", 200, `"logs":[]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := dbaastest.NewServer()
			defer srv.Close()

			status, body := send(t, srv, tt.method, tt.path, tt.token, tt.key, tt.body)

			if status != tt.wantStatus || !strings.Contains(body, tt.wantBody) {
				t.Errorf("got %d %s, want %d containing %s", status, body, tt.wantStatus, tt.wantBody)
			}
		})
	}
}

func TestContractCreateIsIdempotent(t *testing.T) {
	srv := dbaastest.NewServer()
	defer srv.Close()

	const create = `{"name":"orders","version":16,"storage":"1Gi"}`
	_, first := send(t, srv, "POST", "/api/v1/instances", dbaastest.Token, "same", create)
	_, second := send(t, srv, "POST", "/api/v1/instances", dbaastest.Token, "same", create)
	_, list := send(t, srv, "GET", "/api/v1/instances", dbaastest.Token, "", "")

	if first != second {
		t.Errorf("retry returned %s, want %s", second, first)
	}
	if n := strings.Count(list, `"id"`); n != 1 {
		t.Errorf("list has %d instances, want 1", n)
	}
}

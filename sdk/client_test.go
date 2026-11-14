package dbaas_test

// Unit tests for behaviour the examples don't show: what goes over the wire,
// error decoding, token handling and waiting.

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	dbaas "github.com/foyez/dbaas-platform/sdk"
	"github.com/foyez/dbaas-platform/sdk/dbaastest"
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

func TestAPIErrorDecoding(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		wantErr  error
		wantCode string
		wantMsg  string
	}{
		{
			name:     "error envelope",
			status:   http.StatusConflict,
			body:     `{"error":{"code":"INSTANCE_ALREADY_EXISTS","message":"instance already exists"}}`,
			wantErr:  dbaas.ErrAlreadyExists,
			wantCode: "INSTANCE_ALREADY_EXISTS",
			wantMsg:  "instance already exists",
		},
		{
			name:     "code wins over status",
			status:   http.StatusBadRequest,
			body:     `{"error":{"code":"NO_UPDATE_FIELDS","message":"no fields"}}`,
			wantErr:  dbaas.ErrNoUpdateFields,
			wantCode: "NO_UPDATE_FIELDS",
			wantMsg:  "no fields",
		},
		{
			name:    "plain text from the auth middleware",
			status:  http.StatusUnauthorized,
			body:    "invalid token\n",
			wantErr: dbaas.ErrUnauthorized,
			wantMsg: "invalid token",
		},
		{
			name:    "empty body",
			status:  http.StatusForbidden,
			wantErr: dbaas.ErrForbidden,
			wantMsg: "Forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := recordServer(t, tt.status, tt.body)
			c := dbaas.New(srv.URL)

			_, err := c.Instances.Get(t.Context(), "x")

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("errors.Is(%v, %v) = false", err, tt.wantErr)
			}

			var apiErr *dbaas.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("error %v is not an *APIError", err)
			}
			if apiErr.StatusCode != tt.status || apiErr.Code != tt.wantCode || apiErr.Message != tt.wantMsg {
				t.Errorf("got {%d %q %q}, want {%d %q %q}",
					apiErr.StatusCode, apiErr.Code, apiErr.Message, tt.status, tt.wantCode, tt.wantMsg)
			}
		})
	}
}

func TestAPIErrorIsOnlyItsOwnSentinel(t *testing.T) {
	err := &dbaas.APIError{StatusCode: http.StatusNotFound, Code: "NOT_FOUND"}

	if errors.Is(err, dbaas.ErrForbidden) {
		t.Error("NOT_FOUND matched ErrForbidden")
	}
	if _, ok := err.QueryDetails(); ok {
		t.Error("QueryDetails ok for a NOT_FOUND error")
	}
}

func TestMissingIdempotencyKey(t *testing.T) {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL, dbaas.WithToken(dbaastest.Token))

	_, err := c.Instances.Create(t.Context(), dbaas.CreateInstanceInput{Name: "orders", Version: 16, Storage: "1Gi"})

	if !errors.Is(err, dbaas.ErrMissingIdempotencyKey) {
		t.Errorf("err = %v, want ErrMissingIdempotencyKey", err)
	}
}

func TestNoTokenIsUnauthorized(t *testing.T) {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL)

	_, err := c.Instances.List(t.Context(), dbaas.ListOptions{})

	if !errors.Is(err, dbaas.ErrUnauthorized) {
		t.Errorf("err = %v, want ErrUnauthorized", err)
	}
}

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

func TestWaitUntilRunningFailed(t *testing.T) {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL, dbaas.WithToken(dbaastest.Token), dbaas.WithPollInterval(time.Millisecond))

	inst, err := c.Instances.Create(t.Context(), dbaas.CreateInstanceInput{
		Name: "orders", Version: 16, Storage: "1Gi", IdempotencyKey: "k",
	})
	if err != nil {
		t.Fatal(err)
	}
	srv.SetStatus(inst.ID, dbaas.StatusFailed)

	got, err := c.Instances.WaitUntilRunning(t.Context(), inst.ID)

	if !errors.Is(err, dbaas.ErrInstanceFailed) {
		t.Errorf("err = %v, want ErrInstanceFailed", err)
	}
	if got == nil || got.Status != dbaas.StatusFailed {
		t.Errorf("instance = %+v, want the failed instance", got)
	}
}

func TestWaitUntilRunningHonoursContext(t *testing.T) {
	// An instance that stays Pending forever.
	srv, _ := recordServer(t, http.StatusOK, `{"id":"x","status":"Pending"}`)
	c := dbaas.New(srv.URL, dbaas.WithPollInterval(time.Millisecond))

	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()

	_, err := c.Instances.WaitUntilRunning(ctx, "x")

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want context.DeadlineExceeded", err)
	}
}

func TestAllYieldsError(t *testing.T) {
	srv, _ := recordServer(t, http.StatusInternalServerError, `{"error":{"code":"INTERNAL_ERROR","message":"boom"}}`)
	c := dbaas.New(srv.URL)

	var errs int
	for inst, err := range c.Instances.All(t.Context(), 0) {
		if inst != nil {
			t.Errorf("got instance %v with an error", inst)
		}
		if err != nil {
			errs++
		}
	}

	if errs != 1 {
		t.Errorf("errors yielded = %d, want 1", errs)
	}
}

func TestAllStopsWhenLoopBreaks(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte(`{"items":[{"id":"a"},{"id":"b"}],"nextCursor":"more"}`))
	}))
	defer srv.Close()
	c := dbaas.New(srv.URL)

	for range c.Instances.All(t.Context(), 2) {
		break
	}

	if n := requests.Load(); n != 1 {
		t.Errorf("requests = %d, want 1: breaking out must not fetch more pages", n)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

package dbaas_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/foyez/dbaas-platform/sdk/dbaas"
	"github.com/foyez/dbaas-platform/sdk/dbaas/dbaastest"
)

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

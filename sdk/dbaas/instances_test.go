package dbaas_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/foyez/dbaas-platform/sdk/dbaas"
	"github.com/foyez/dbaas-platform/sdk/dbaas/dbaastest"
)

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

	for _, err := range c.Instances.All(t.Context(), 2) {
		if err != nil {
			t.Fatal(err)
		}
		break // after the first of the page's two instances
	}

	if n := requests.Load(); n != 1 {
		t.Errorf("requests = %d, want 1: breaking out must not fetch more pages", n)
	}
}

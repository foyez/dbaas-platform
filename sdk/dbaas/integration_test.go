//go:build integration

package dbaas_test

// Integration test: the SDK against a real API, normally the local platform
// from `make dev`. It creates a real instance, so it's behind a build tag
// and only runs with `make test-integration` (which loads .env).

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/foyez/dbaas-platform/sdk/dbaas"
)

// integrationClient builds a client from the DBAAS_* environment variables
// documented in .env.example, skipping the test when they're missing.
func integrationClient(t *testing.T) *dbaas.Client {
	t.Helper()

	env := map[string]string{}
	for _, name := range []string{"DBAAS_API_URL", "DBAAS_TOKEN_URL", "DBAAS_CLIENT_ID", "DBAAS_CLIENT_SECRET", "DBAAS_PROJECT_ID"} {
		env[name] = os.Getenv(name)
		if env[name] == "" {
			t.Skipf("%s is not set; copy .env.example to .env and fill it in", name)
		}
	}

	return dbaas.New(env["DBAAS_API_URL"],
		dbaas.WithClientCredentials(env["DBAAS_TOKEN_URL"], env["DBAAS_CLIENT_ID"], env["DBAAS_CLIENT_SECRET"], []string{
			"openid",
			"urn:zitadel:iam:org:project:id:" + env["DBAAS_PROJECT_ID"] + ":aud",
			"urn:zitadel:iam:org:project:roles",
		}),
	)
}

func TestIntegrationInstanceLifecycle(t *testing.T) {
	c := integrationClient(t)

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Minute)
	defer cancel()

	if _, err := c.Health(ctx); err != nil {
		t.Fatalf("health: %v", err)
	}

	// A fresh key per run, so each run creates a new instance.
	key := fmt.Sprintf("sdk-it-%d", time.Now().UnixNano())

	inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
		Name:           "sdkit",
		Version:        16,
		Storage:        "1Gi",
		IdempotencyKey: key,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Logf("created %s", inst.ID)

	t.Cleanup(func() {
		// t.Context() is already cancelled during cleanup.
		if err := c.Instances.Delete(context.Background(), inst.ID); err != nil && !errors.Is(err, dbaas.ErrNotFound) {
			t.Errorf("delete: %v", err)
		}
	})

	replay, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
		Name: "sdkit", Version: 16, Storage: "1Gi", IdempotencyKey: key,
	})
	if err != nil {
		t.Fatalf("create retry: %v", err)
	}
	if replay.ID != inst.ID {
		t.Errorf("create retry returned %s, want %s", replay.ID, inst.ID)
	}

	if _, err := c.Instances.WaitUntilRunning(ctx, inst.ID); err != nil {
		t.Fatalf("wait: %v", err)
	}

	found := false
	for got, err := range c.Instances.All(ctx, 0) {
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		found = found || got.ID == inst.ID
	}
	if !found {
		t.Errorf("list doesn't contain %s", inst.ID)
	}

	creds, err := c.Instances.Credentials(ctx, inst.ID)
	if err != nil {
		t.Fatalf("credentials: %v", err)
	}
	if creds.Database != "sdkit" || creds.URI == "" {
		t.Errorf("credentials = %+v, want database sdkit and a URI", creds)
	}

	res, err := c.Instances.Query(ctx, inst.ID, "SELECT 1")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if res.RowCount != 1 {
		t.Errorf("query row count = %d, want 1", res.RowCount)
	}

	_, err = c.Instances.Query(ctx, inst.ID, "SELEC 1")
	var apiErr *dbaas.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("bad query: err = %v, want an *APIError", err)
	}
	if details, ok := apiErr.QueryDetails(); !ok || details.SQLState != "42601" {
		t.Errorf("bad query details = %+v, want SQLSTATE 42601", details)
	}

	if err := c.Instances.Delete(ctx, inst.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

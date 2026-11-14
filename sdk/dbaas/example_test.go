package dbaas_test

// Testable examples: go test runs each one and compares what it prints with
// its "// Output:" comment, and pkg.go.dev shows them in the docs. They run
// against the in-memory fake API from package dbaastest, so they need no
// cluster. integration_test.go runs the same calls against a real API.

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/foyez/dbaas-platform/sdk/dbaas"
	"github.com/foyez/dbaas-platform/sdk/dbaas/dbaastest"
)

// The whole lifecycle of an instance: create it, wait for it, connect,
// change it, delete it.
func Example() {
	srv := dbaastest.NewServer()
	defer srv.Close()

	c := dbaas.New(srv.URL,
		dbaas.WithToken(dbaastest.Token),
		dbaas.WithPollInterval(time.Millisecond), // the fake is instant
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
		Name:           "orders",
		Version:        16,
		Storage:        "10Gi",
		IdempotencyKey: "orders-db",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("created:", inst.Status)

	inst, err = c.Instances.WaitUntilRunning(ctx, inst.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("now:", inst.Status)

	creds, err := c.Instances.Credentials(ctx, inst.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("database:", creds.Database)

	storage := "20Gi"
	inst, err = c.Instances.Update(ctx, inst.ID, dbaas.UpdateInstanceInput{Storage: &storage})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("resizing to", inst.Storage, inst.Status)

	if err := c.Instances.Delete(ctx, inst.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted")

	// Output:
	// created: Pending
	// now: Running
	// database: orders
	// resizing to 20Gi Pending
	// deleted
}

func ExampleClient_Health() {
	srv := dbaastest.NewServer()
	defer srv.Close()

	// /health needs no token.
	c := dbaas.New(srv.URL)

	status, err := c.Health(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(status.Status)
	// Output: ok
}

// Machine users (CI jobs, scripts) log in with the OAuth2 client credentials
// grant. The SDK fetches a token on the first call and reuses it.
func ExampleWithClientCredentials() {
	srv := dbaastest.NewServer()
	defer srv.Close()

	projectID := "123456789"
	c := dbaas.New(srv.URL,
		dbaas.WithClientCredentials(srv.TokenURL, dbaastest.ClientID, dbaastest.ClientSecret, []string{
			"openid",
			"urn:zitadel:iam:org:project:id:" + projectID + ":aud",
			"urn:zitadel:iam:org:project:roles",
		}),
	)

	page, err := c.Instances.List(context.Background(), dbaas.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("instances:", len(page.Items))
	// Output: instances: 0
}

func ExampleInstancesService_Create() {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL, dbaas.WithToken(dbaastest.Token))

	inst, err := c.Instances.Create(context.Background(), dbaas.CreateInstanceInput{
		Name:           "orders",
		Version:        16,
		Storage:        "10Gi",
		Instances:      2,
		Username:       "app",
		IdempotencyKey: "orders-db",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Creation is asynchronous: the instance starts Pending.
	fmt.Println(inst.Database, inst.Version, inst.Storage, inst.Instances, inst.Status)
	// Output: orders 16 10Gi 2 Pending
}

// Retrying a create with the same idempotency key returns the instance the
// first call created, so a create that timed out can be retried safely.
func ExampleInstancesService_Create_retry() {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL, dbaas.WithToken(dbaastest.Token))
	ctx := context.Background()

	input := dbaas.CreateInstanceInput{
		Name:           "orders",
		Version:        16,
		Storage:        "10Gi",
		IdempotencyKey: "orders-db",
	}

	first, err := c.Instances.Create(ctx, input)
	if err != nil {
		log.Fatal(err)
	}
	second, err := c.Instances.Create(ctx, input)
	if err != nil {
		log.Fatal(err)
	}

	page, err := c.Instances.List(ctx, dbaas.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("same instance:", first.ID == second.ID)
	fmt.Println("instances:", len(page.Items))
	// Output:
	// same instance: true
	// instances: 1
}

func ExampleInstancesService_Get() {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL, dbaas.WithToken(dbaastest.Token))
	ctx := context.Background()

	created, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
		Name: "orders", Version: 16, Storage: "10Gi", IdempotencyKey: "orders-db",
	})
	if err != nil {
		log.Fatal(err)
	}

	inst, err := c.Instances.Get(ctx, created.ID)
	if err != nil {
		log.Fatal(err)
	}

	// Database is the name you chose; ID and Name are generated.
	fmt.Println(inst.Database, inst.ID == created.ID)
	// Output: orders true
}

func ExampleInstancesService_WaitUntilRunning() {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL,
		dbaas.WithToken(dbaastest.Token),
		dbaas.WithPollInterval(time.Millisecond),
	)

	// Always bound the wait: provisioning can take minutes, or never end.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
		Name: "orders", Version: 16, Storage: "10Gi", IdempotencyKey: "orders-db",
	})
	if err != nil {
		log.Fatal(err)
	}

	inst, err = c.Instances.WaitUntilRunning(ctx, inst.ID)
	switch {
	case errors.Is(err, dbaas.ErrInstanceFailed):
		log.Fatal("provisioning failed")
	case errors.Is(err, context.DeadlineExceeded):
		log.Fatal("still not running after 10 minutes")
	case err != nil:
		log.Fatal(err)
	}

	fmt.Println(inst.Status, inst.ReadyInstances)
	// Output: Running 1
}

func ExampleInstancesService_List() {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL, dbaas.WithToken(dbaastest.Token))
	ctx := context.Background()

	for _, name := range []string{"orders", "users", "billing"} {
		_, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
			Name: name, Version: 16, Storage: "1Gi", IdempotencyKey: name,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	// Follow NextCursor until it's empty to read every page.
	opts := dbaas.ListOptions{Limit: 2}
	for {
		page, err := c.Instances.List(ctx, opts)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("page with", len(page.Items), "instances")

		if page.NextCursor == "" {
			break
		}
		opts.Cursor = page.NextCursor
	}

	// Output:
	// page with 2 instances
	// page with 1 instances
}

// All does the cursor-following of List for you, as a range-over-func
// iterator.
func ExampleInstancesService_All() {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL, dbaas.WithToken(dbaastest.Token))
	ctx := context.Background()

	for _, name := range []string{"orders", "users", "billing"} {
		_, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
			Name: name, Version: 16, Storage: "1Gi", IdempotencyKey: name,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	for inst, err := range c.Instances.All(ctx, 2) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(inst.Database)
	}

	// Output:
	// orders
	// users
	// billing
}

func ExampleInstancesService_Update() {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL, dbaas.WithToken(dbaastest.Token))
	ctx := context.Background()

	inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
		Name: "orders", Version: 15, Storage: "10Gi", IdempotencyKey: "orders-db",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Fields are pointers: nil means "leave as is".
	version := 16
	inst, err = c.Instances.Update(ctx, inst.ID, dbaas.UpdateInstanceInput{Version: &version})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(inst.Version, inst.Storage, inst.Status)
	// Output: 16 10Gi Pending
}

func ExampleInstancesService_Delete() {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL, dbaas.WithToken(dbaastest.Token))
	ctx := context.Background()

	inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
		Name: "orders", Version: 16, Storage: "10Gi", IdempotencyKey: "orders-db",
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := c.Instances.Delete(ctx, inst.ID); err != nil {
		log.Fatal(err)
	}

	_, err = c.Instances.Get(ctx, inst.ID)
	fmt.Println("gone:", errors.Is(err, dbaas.ErrNotFound))
	// Output: gone: true
}

func ExampleInstancesService_Credentials() {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL,
		dbaas.WithToken(dbaastest.Token),
		dbaas.WithPollInterval(time.Millisecond),
	)
	ctx := context.Background()

	inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
		Name: "orders", Version: 16, Storage: "10Gi", IdempotencyKey: "orders-db",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Credentials need a reachable instance.
	_, err = c.Instances.Credentials(ctx, inst.ID)
	fmt.Println("not ready yet:", errors.Is(err, dbaas.ErrInstanceNotReady))

	if _, err := c.Instances.WaitUntilRunning(ctx, inst.ID); err != nil {
		log.Fatal(err)
	}

	creds, err := c.Instances.Credentials(ctx, inst.ID)
	if err != nil {
		log.Fatal(err)
	}

	// Pass creds.URI to pgx, database/sql or psql.
	fmt.Println(creds.URI)
	// Output:
	// not ready yet: true
	// postgresql://app:secret@db.dbaas.localhost:30001/orders
}

func ExampleInstancesService_Logs() {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL, dbaas.WithToken(dbaastest.Token))
	ctx := context.Background()

	inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
		Name: "orders", Version: 16, Storage: "10Gi", IdempotencyKey: "orders-db",
	})
	if err != nil {
		log.Fatal(err)
	}

	lines, err := c.Instances.Logs(ctx, inst.ID)
	if err != nil {
		log.Fatal(err)
	}

	for _, l := range lines {
		fmt.Println(l.Line)
	}
	// Output: database system is ready to accept connections
}

func ExampleInstancesService_Query() {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL,
		dbaas.WithToken(dbaastest.Token),
		dbaas.WithPollInterval(time.Millisecond),
	)
	ctx := context.Background()

	inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
		Name: "orders", Version: 16, Storage: "10Gi", IdempotencyKey: "orders-db",
	})
	if err != nil {
		log.Fatal(err)
	}
	if _, err := c.Instances.WaitUntilRunning(ctx, inst.ID); err != nil {
		log.Fatal(err)
	}

	res, err := c.Instances.Query(ctx, inst.ID, "SELECT 1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res.Columns, res.Rows, res.RowCount)
	// Output: [?column?] [[1]] 1
}

// A statement PostgreSQL rejects comes back as ErrQueryFailed, with
// PostgreSQL's own error in the details.
func ExampleAPIError_QueryDetails() {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL,
		dbaas.WithToken(dbaastest.Token),
		dbaas.WithPollInterval(time.Millisecond),
	)
	ctx := context.Background()

	inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
		Name: "orders", Version: 16, Storage: "10Gi", IdempotencyKey: "orders-db",
	})
	if err != nil {
		log.Fatal(err)
	}
	if _, err := c.Instances.WaitUntilRunning(ctx, inst.ID); err != nil {
		log.Fatal(err)
	}

	_, err = c.Instances.Query(ctx, inst.ID, "SELEC 1")

	var apiErr *dbaas.APIError
	if errors.As(err, &apiErr) {
		if details, ok := apiErr.QueryDetails(); ok {
			fmt.Println(apiErr.Message)
			fmt.Println("SQLSTATE", details.SQLState, "at position", details.Position)
		}
	}
	// Output:
	// syntax error at or near "SELEC"
	// SQLSTATE 42601 at position 1
}

// Use errors.Is with the sentinel errors to decide what to do, and
// errors.As to read the API's error code and message.
func ExampleAPIError() {
	srv := dbaastest.NewServer()
	defer srv.Close()
	c := dbaas.New(srv.URL, dbaas.WithToken(dbaastest.Token))

	_, err := c.Instances.Get(context.Background(), "instance-does-not-exist")

	fmt.Println(errors.Is(err, dbaas.ErrNotFound))

	var apiErr *dbaas.APIError
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.StatusCode, apiErr.Code)
	}

	fmt.Println(err)
	// Output:
	// true
	// 404 NOT_FOUND
	// get instance: dbaas: 404 NOT_FOUND: resource not found
}

func ExampleClient_AuditLogs() {
	srv := dbaastest.NewServer()
	defer srv.Close()
	ctx := context.Background()

	user := dbaas.New(srv.URL, dbaas.WithToken(dbaastest.Token))
	inst, err := user.Instances.Create(ctx, dbaas.CreateInstanceInput{
		Name: "orders", Version: 16, Storage: "10Gi", IdempotencyKey: "orders-db",
	})
	if err != nil {
		log.Fatal(err)
	}

	// The audit trail is for admins only.
	_, err = user.AuditLogs(ctx, dbaas.AuditLogsOptions{})
	fmt.Println("user forbidden:", errors.Is(err, dbaas.ErrForbidden))

	admin := dbaas.New(srv.URL, dbaas.WithToken(dbaastest.AdminToken))
	page, err := admin.AuditLogs(ctx, dbaas.AuditLogsOptions{ResourceID: inst.ID})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("events:", len(page.Logs))
	// Output:
	// user forbidden: true
	// events: 1
}

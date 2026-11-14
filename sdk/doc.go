// Package dbaas is a Go client for the DBaaS Platform REST API.
//
// Create a Client with New, choose how it authenticates with an Option, then
// call the API through it:
//
//	c := dbaas.New("https://api.example.com",
//		dbaas.WithClientCredentials(tokenURL, clientID, clientSecret, scopes),
//	)
//
//	inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
//		Name:           "orders",
//		Version:        16,
//		Storage:        "10Gi",
//		IdempotencyKey: "orders-db",
//	})
//	inst, err = c.Instances.WaitUntilRunning(ctx, inst.ID)
//	creds, err := c.Instances.Credentials(ctx, inst.ID)
//
// # Contexts
//
// Every method takes a context.Context, so callers control timeouts and
// cancellation.
//
// # Errors
//
// A non-2xx response is returned as an *APIError, wrapped with the name of
// the operation. Match it against the sentinel errors with errors.Is, or get
// at its fields with errors.As:
//
//	if errors.Is(err, dbaas.ErrNotFound) { ... }
//
//	var apiErr *dbaas.APIError
//	if errors.As(err, &apiErr) { log.Println(apiErr.Code) }
//
// # Testing
//
// Package dbaastest runs an in-memory fake of the API, to test code that
// uses this package without a cluster.
package dbaas

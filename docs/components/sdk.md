# Go SDK

[`sdk/`](https://github.com/foyez/dbaas-platform/tree/main/sdk) is a typed Go client for the REST API, for teams that want to manage databases from code or CI instead of the UI. It's a thin wrapper: every method is one API call, plus `WaitUntilRunning` and `All`, which loop over calls.

```sh
go get github.com/foyez/dbaas-platform/sdk
```

```go
import dbaas "github.com/foyez/dbaas-platform/sdk"

c := dbaas.New("https://api.example.com",
    dbaas.WithClientCredentials(tokenURL, clientID, clientSecret, []string{
        "openid",
        "urn:zitadel:iam:org:project:id:" + projectID + ":aud",
    }),
)

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
defer cancel()

inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
    Name:           "orders",      // the database name; the ID is generated
    Version:        16,
    Storage:        "10Gi",
    IdempotencyKey: "orders-db",   // reuse it on every retry of this create
})
inst, err = c.Instances.WaitUntilRunning(ctx, inst.ID)
creds, err := c.Instances.Credentials(ctx, inst.ID) // creds.URI is a postgresql:// URL
```

## Authentication

| Option | For | Token lifetime |
|---|---|---|
| `WithClientCredentials` | Machine users: CI, scripts, services | Fetched on first call, cached, refetched before expiry |
| `WithToken` | A token you already have, e.g. a user's | Fixed; fails with `ErrUnauthorized` once expired |
| neither | `Health` only | — |

`WithClientCredentials` uses the OAuth2 client-credentials grant against ZITADEL. The project audience scope `urn:zitadel:iam:org:project:id:<projectID>:aud` is required, or the API can't validate the token. Add `urn:zitadel:iam:org:project:roles` to put the user's roles in the token.

## Operations

| Method | Endpoint | Notes |
|---|---|---|
| `Health` | `GET /health` | No token needed |
| `Instances.Create` | `POST /api/v1/instances` | `IdempotencyKey` is required; returns `Pending` |
| `Instances.List` | `GET /api/v1/instances` | One page; follow `NextCursor` |
| `Instances.All` | `GET /api/v1/instances` (repeated) | Iterator over every page |
| `Instances.Get` | `GET /api/v1/instances/{id}` | |
| `Instances.WaitUntilRunning` | `GET /api/v1/instances/{id}` (repeated) | Polls every 5s (`WithPollInterval`) |
| `Instances.Update` | `PATCH /api/v1/instances/{id}` | `nil` fields are left unchanged |
| `Instances.Delete` | `DELETE /api/v1/instances/{id}` | Asynchronous |
| `Instances.Credentials` | `GET /api/v1/instances/{id}/credentials` | `ErrInstanceNotReady` until reachable |
| `Instances.Logs` | `GET /api/v1/instances/{id}/logs` | |
| `Instances.Query` | `POST /api/v1/instances/{id}/query` | SQL console; one statement |
| `AuditLogs` | `GET /api/v1/audit-logs` | Admin role only |

## Errors

Every non-2xx response becomes an `*APIError` with the status code, the API's error code, message and details. It matches the sentinel errors with `errors.Is`:

```go
_, err := c.Instances.Get(ctx, id)
switch {
case errors.Is(err, dbaas.ErrNotFound):      // doesn't exist, or isn't yours
case errors.Is(err, dbaas.ErrUnauthorized):  // bad or expired token
case err != nil:
    var apiErr *dbaas.APIError
    if errors.As(err, &apiErr) {
        log.Println(apiErr.StatusCode, apiErr.Code, apiErr.Message)
    }
}
```

| Sentinel | API error code / status |
|---|---|
| `ErrInvalidRequest` | `INVALID_REQUEST` |
| `ErrUnauthorized` | `UNAUTHORIZED`, or 401 |
| `ErrForbidden` | `FORBIDDEN`, or 403 |
| `ErrNotFound` | `NOT_FOUND`, `INSTANCE_NOT_FOUND`, or 404 |
| `ErrAlreadyExists` | `INSTANCE_ALREADY_EXISTS` |
| `ErrInstanceNotReady` | `INSTANCE_NOT_READY` |
| `ErrMissingIdempotencyKey` | `MISSING_IDEMPOTENCY_KEY` |
| `ErrNoUpdateFields` | `NO_UPDATE_FIELDS` |
| `ErrQueryFailed` | `QUERY_FAILED`; `APIError.QueryDetails()` has PostgreSQL's SQLSTATE and position |
| `ErrQueryTimeout` | `QUERY_TIMEOUT` |
| `ErrInstanceFailed` | From `WaitUntilRunning`, when the instance ends up `Failed` |

## Project layout

| Path | What's in it |
|---|---|
| `doc.go` | Package documentation |
| `client.go`, `options.go`, `auth.go` | `New`, functional options, authentication |
| `request.go` | The one place requests are built, authenticated, sent and decoded |
| `errors.go` | `APIError` and the sentinels |
| `types.go` | Response types |
| `health.go`, `instances.go`, `audit.go` | One file per API area |
| `dbaastest/` | In-memory fake API, exported so SDK users can test against it too |
| `example_test.go` | Testable examples |
| `client_test.go` | Unit tests |
| `integration_test.go` | End-to-end test against a real API, behind the `integration` build tag |
| `examples/provision/` | A runnable program that provisions a database for a CI job |

## Testing

There are three layers, from fast and always-on to slow and opt-in:

| Layer | Runs against | Command |
|---|---|---|
| Testable examples (`Example...` functions) | `dbaastest` fake | `make test` |
| Unit tests (`Test...` functions) | `dbaastest` fake, or a one-response `httptest` server | `make test` |
| Integration test | A real API and ZITADEL | `make test-integration` |

The examples are the documentation. `go test` runs each one and fails if what it prints differs from its `// Output:` comment, so the examples on pkg.go.dev can't go stale. Browse them locally with `make docs`.

`make test` needs no cluster and no credentials, so CI ([`.forgejo/workflows/sdk-ci.yaml`](https://github.com/foyez/dbaas-platform/blob/main/.forgejo/workflows/sdk-ci.yaml)) runs it on every change to `sdk/`.

### Test your own code with `dbaastest`

```go
func TestProvision(t *testing.T) {
    srv := dbaastest.NewServer()
    defer srv.Close()

    c := dbaas.New(srv.URL,
        dbaas.WithToken(dbaastest.Token),
        dbaas.WithPollInterval(time.Millisecond),
    )

    // ... call your code with c ...
}
```

The fake implements every endpoint with the real error codes. A new instance is `Pending` and becomes `Running` the next time it's fetched, and `srv.SetStatus(id, dbaas.StatusFailed)` simulates a failure. `srv.TokenURL` with `dbaastest.ClientID` and `dbaastest.ClientSecret` tests the client-credentials flow.

## Try it against the local platform

The local ZITADEL has an SDK machine user (`sdk-test-dev`, [`infra/zitadel/tf/sdk_dev_user.tf`](https://github.com/foyez/dbaas-platform/blob/main/infra/zitadel/tf/sdk_dev_user.tf)). With the platform running (`make dev`):

```sh
cd sdk
cp .env.example .env
# fill in the three values .env.example points to:
terraform -chdir=../hack/dev/terraform output -raw sdk_client_id
terraform -chdir=../hack/dev/terraform output -raw sdk_client_secret
terraform -chdir=../hack/dev/terraform output -raw project_id

make test-integration                       # create, wait, query, delete
set -a; . ./.env; set +a
go run ./examples/provision -name orders -key demo-1   # prints the connection URI
go run ./examples/provision -key demo-1 -delete
```

How the SDK is designed, and what went wrong while building it, is in the learning notes: [Building a Go SDK](../notes/go-sdk.md).

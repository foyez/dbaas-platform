# Go SDK

[`sdk/`](https://github.com/foyez/dbaas-platform/tree/main/sdk) is a typed Go client for the REST API, for teams that want to manage databases from code or CI instead of the UI. It's a thin wrapper: every method is one API call, plus `WaitUntilRunning` and `All`, which loop over calls.

```sh
go get github.com/foyez/dbaas-platform/sdk/dbaas
```

```go
import "github.com/foyez/dbaas-platform/sdk/dbaas"

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

```text
sdk/                        module github.com/foyez/dbaas-platform/sdk
├── dbaas/                  package dbaas: the client
│   ├── doc.go              package documentation
│   ├── client.go           Client and New
│   ├── options.go          functional options
│   ├── auth.go             WithClientCredentials, WithToken
│   ├── request.go          the one place requests are sent and decoded
│   ├── errors.go           APIError and sentinel errors
│   ├── types.go            response types
│   ├── version.go          Version, sent as the User-Agent
│   ├── health.go, instances.go, audit.go   one file per API area
│   ├── *_test.go           unit tests, one per source file
│   ├── example_test.go     testable examples
│   ├── integration_test.go against a real API, behind the integration build tag
│   └── dbaastest/          package dbaastest: in-memory fake API for tests
├── examples/provision/     runnable program: provision a database for CI
├── .golangci.yml           lint configuration
├── CHANGELOG.md
└── Makefile
```

The library lives in `sdk/dbaas/`, not the module root, so the import path ends in the package name. The module root keeps the tooling, examples and docs. The fake is a subpackage of the package it fakes, the way `net/http/httptest` sits under `net/http`.

## Testing

There are four layers, from fast and always-on to slow and opt-in:

| Layer | Runs against | Command |
|---|---|---|
| Testable examples (`example_test.go`) | `dbaastest` fake | `make test` |
| Unit tests (`request_test.go`, `errors_test.go`, ...) | `dbaastest` fake, or a one-response `httptest` server | `make test` |
| Fake contract tests (`dbaastest/server_test.go`) | `dbaastest` fake, in raw HTTP | `make test` |
| Integration test (`integration_test.go`) | A real API and ZITADEL | `make test-integration` |

The examples are the documentation. `go test` runs each one and fails if what it prints differs from its `// Output:` comment, so the examples on pkg.go.dev can't go stale. Browse them locally with `make docs`.

The contract tests pin the API behaviour the fake has to copy (status codes, error codes, the plain-text 401, idempotent create). The integration test checks the same behaviour on the real API.

`make test` and `make lint` need no cluster and no credentials, so CI ([`.forgejo/workflows/sdk-ci.yaml`](https://github.com/foyez/dbaas-platform/blob/main/.forgejo/workflows/sdk-ci.yaml)) runs them on every change to `sdk/`.

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

## Releasing

The SDK is its own Go module inside the repository, so releases are tags prefixed with its directory, like the operator's `operator/vX.Y.Z`:

1. Move the `Unreleased` entries in [`sdk/CHANGELOG.md`](https://github.com/foyez/dbaas-platform/blob/main/sdk/CHANGELOG.md) under the new version.
2. Set `Version` in `sdk/dbaas/version.go`; it's sent as `User-Agent: dbaas-go/<version>`.
3. `git tag sdk/v0.2.0 && git push origin sdk/v0.2.0`.

Users then run `go get github.com/foyez/dbaas-platform/sdk/dbaas@v0.2.0`.

How the SDK is designed, and what went wrong while building it, is in the learning notes: [Building a Go SDK](../notes/go-sdk.md).

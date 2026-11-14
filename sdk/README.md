# DBaaS Platform Go SDK

A typed Go client for the DBaaS Platform REST API: create, scale, query and delete PostgreSQL instances from code or CI.

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

inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
    Name:           "orders",
    Version:        16,
    Storage:        "10Gi",
    IdempotencyKey: "orders-db",
})
inst, err = c.Instances.WaitUntilRunning(ctx, inst.ID)
creds, err := c.Instances.Credentials(ctx, inst.ID)
```

Full documentation, with the reasoning behind the design: [Go SDK](../docs/components/sdk.md) and the learning notes in [Building a Go SDK](../docs/notes/go-sdk.md).

## Layout

| Path | What's in it |
|---|---|
| `doc.go` | Package documentation |
| `client.go`, `options.go`, `auth.go` | `New`, functional options, authentication |
| `request.go` | The single place HTTP requests are built, sent and decoded |
| `errors.go` | `APIError` and the sentinel errors it matches |
| `types.go` | Response types (`Instance`, `InstanceCredentials`, ...) |
| `health.go`, `instances.go`, `audit.go` | One file per API area |
| `dbaastest/` | In-memory fake API, for testing code that uses the SDK |
| `example_test.go` | Testable examples: run by `go test`, shown on pkg.go.dev |
| `client_test.go` | Unit tests for wire format, errors, tokens, waiting |
| `integration_test.go` | End-to-end test against a real API (`integration` build tag) |
| `examples/provision/` | Runnable program: provision a database for a CI job |

## Development

```sh
make test              # unit tests + examples, against the fake API
make vet               # gofmt + go vet
make cover             # coverage per function
make docs              # browse the package docs at http://localhost:6060
make test-integration  # against a real API; needs .env (see .env.example)
```

The integration test and `examples/provision` read the `DBAAS_*` variables from `.env`. For the local platform (`make dev` at the repository root), `.env.example` shows where each value comes from.

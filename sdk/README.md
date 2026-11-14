# DBaaS Platform Go SDK

A typed Go client for the DBaaS Platform REST API: create, scale, query and delete PostgreSQL instances from code or CI.

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

inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
    Name:           "orders",
    Version:        16,
    Storage:        "10Gi",
    IdempotencyKey: "orders-db",
})
inst, err = c.Instances.WaitUntilRunning(ctx, inst.ID)
creds, err := c.Instances.Credentials(ctx, inst.ID)
```

Requires Go 1.25 or later. Full documentation: [Go SDK](../docs/components/sdk.md). Design and lessons learned: [Building a Go SDK](../docs/notes/go-sdk.md).

## Layout

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
│   ├── example_test.go     testable examples (run by go test, shown on pkg.go.dev)
│   ├── integration_test.go against a real API, behind the integration build tag
│   └── dbaastest/          package dbaastest: in-memory fake API for tests
├── examples/provision/     runnable program: provision a database for CI
├── .golangci.yml           lint configuration
├── CHANGELOG.md
└── Makefile
```

The import path ends in the package name (`.../sdk/dbaas` → `package dbaas`), and test helpers live in a `…test` subpackage, like `net/http` and `net/http/httptest`.

## Development

```sh
make test              # unit tests, examples and fake contract tests; no cluster needed
make lint              # golangci-lint (downloaded into ./bin)
make cover             # coverage per function
make docs              # browse the package docs at http://localhost:6060
make test-integration  # against a real API; needs .env (see .env.example)
```

The integration test and `examples/provision` read the `DBAAS_*` variables from `.env`. For the local platform (`make dev` at the repository root), `.env.example` shows where each value comes from.

## Releasing

The SDK is a Go module in a subdirectory of the repository, so its versions are tags with the directory as prefix:

1. Move the `Unreleased` entries in `CHANGELOG.md` under the new version.
2. Set `Version` in `dbaas/version.go` to match.
3. Tag the commit and push the tag: `git tag sdk/v0.2.0 && git push origin sdk/v0.2.0`.

Users then get it with `go get github.com/foyez/dbaas-platform/sdk/dbaas@v0.2.0`. Before v1, minor versions may break the API; from v1 on, breaking changes need a new major version (`sdk/v2`).

# Go SDK

[`sdk/`](https://github.com/foyez/dbaas-platform/tree/main/sdk) is a typed Go client for the REST API, for teams that want to manage databases from code or CI instead of the UI.

```go
import "github.com/foyez/dbaas-platform/sdk"

c := dbaas.New("https://api.example.com",
    dbaas.WithClientCredentials(tokenURL, clientID, clientSecret, []string{
        "openid",
        "urn:zitadel:iam:org:project:id:" + projectID + ":aud",
    }),
)

inst, err := c.Instances.Create(ctx, dbaas.CreateInstanceInput{
    Name:           "orders-db",
    Version:        16,
    Storage:        "10Gi",
    IdempotencyKey: "orders-db-create-1",
})

creds, err := c.Instances.Credentials(ctx, inst.ID)
```

## Authentication

`WithClientCredentials` uses the OAuth2 client-credentials grant against ZITADEL with a machine user. Tokens are fetched lazily and refreshed automatically when they expire. The project audience scope is required so the API can introspect the token.

## Operations

| Method | Endpoint |
|---|---|
| `Health` | `GET /health` |
| `Instances.Create` | `POST /instances` |
| `Instances.List` | `GET /instances` |
| `Instances.Get` | `GET /instances/{id}` |
| `Instances.Update` | `PATCH /instances/{id}` |
| `Instances.Delete` | `DELETE /instances/{id}` |
| `Instances.Credentials` | `GET /instances/{id}/credentials` |
| `Instances.Logs` | `GET /instances/{id}/logs` |

## Try it against the local platform

The local ZITADEL has an SDK test machine user. Its credentials come from the local Terraform state:

```sh
terraform -chdir=hack/dev/terraform output -raw sdk_client_id
terraform -chdir=hack/dev/terraform output -raw sdk_client_secret
terraform -chdir=hack/dev/terraform output -raw project_id
```

Use `http://auth.dbaas.localhost/oauth/v2/token` as the token URL and `http://api.dbaas.localhost` as the API URL. The examples in `sdk/example_*_test.go` read these from the environment variables listed in `sdk/.env.example`.

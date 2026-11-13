# OpenAPI

## API-First + Code Generation

### Workflow

```
Design API
      ↓
Write openapi.yaml
      ↓
Generate Go types/interfaces
      ↓
Implement handlers
      ↓
Run tests
```

Write:

```
openapi.yaml
```

`oapi-codegen`

```sh
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

Generate models:

```sh
oapi-codegen \
    -generate types \
    -package api \
    api/openapi.yaml \
    > internal/api/types.gen.go
```

Generate Gin server interface:

```sh
oapi-codegen \
    -generate gin-server \
    -package api \
    api/openapi.yaml \
    > internal/api/server.gen.go
```

## Interview Point

I used an API-first workflow. I designed the REST API in OpenAPI, generated strongly typed Go models and Gin server interfaces with oapi-codegen, and implemented the handlers against those generated interfaces. This keeps the implementation synchronized with the API contract and avoids drift.
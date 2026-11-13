# Paas Service Project Structure

## Overview

This project follows a layered architecture where responsibilities are separated between HTTP transport, business logic, domain models, and infrastructure concerns.

The main goal of this structure is:

* Keep business logic independent from frameworks and external systems.
* Make code easier to test and maintain.
* Allow new features to be added without creating tightly coupled code.
* Keep dependencies flowing in one direction.

The dependency direction is:

```
API Layer
    |
    v
Service Layer
    |
    v
Domain Layer
    |
    v
Infrastructure Layer
```

The domain layer remains independent and does not know about HTTP, databases, Kubernetes clients, or logging.

---

# Current Project Structure

```
internal/
├── api/
│   ├── handler/
│   │   └── instance.go
│   ├── router/
│   │   └── router.go
│   └── dto.go
│
├── httpx/
│   ├── response.go
│   ├── error_codes.go
│   └── errors.go
│
├── domain/
│   ├── instance.go
│   └── service.go
│
├── service/
│   ├── instance_service.go
│   └── errors.go
│
├── config/
│   └── config.go
│
├── logger/
│   └── logger.go
│
└── infrastructure/
    └── kubernetes/
```

---

# internal/

The `internal` directory contains application-specific code.

Go prevents external projects from importing packages inside `internal`, which protects application internals from being used as public libraries.

Use `internal` for:

* Business logic
* API implementation
* Configuration
* Infrastructure integrations

Do not put reusable libraries here.

---

# api/

The `api` package represents the HTTP transport layer.

Responsibilities:

* Handle HTTP requests.
* Validate incoming requests.
* Convert HTTP DTOs into domain inputs.
* Call application services.
* Convert domain results into HTTP responses.

The API layer should not contain business rules.

Example flow:

```
HTTP Request
      |
      v
Handler
      |
      v
Service
      |
      v
Domain
```

---

# api/dto.go

DTO means Data Transfer Object.

DTOs define the external API contract.

Responsibilities:

* Define request payloads.
* Define response payloads.
* Handle JSON serialization.

Example:

```go
type CreateInstanceRequest struct {
    Name string `json:"name"`
}

type InstanceResponse struct {
    ID string `json:"id"`
}
```

DTOs should not contain:

* Business logic.
* Database models.
* Infrastructure details.

The API contract can evolve independently from internal models.

---

# api/handler/

Handlers contain HTTP request handling logic.

Responsibilities:

* Read request.
* Validate request.
* Map DTO → domain input.
* Call service interface.
* Map domain output → response DTO.
* Return HTTP status.

Example:

```
POST /v1/instances

        |
        v

CreateInstanceHandler()

        |
        v

InstanceService.CreateInstance()
```

Handlers should remain thin.

---

# api/router/

The router defines HTTP routes and middleware.

Responsibilities:

* Register endpoints.
* Apply middleware.
* Connect URLs to handlers.

Example:

```
POST /v1/instances
        |
        v
CreateInstanceHandler
```

Router should not contain:

* Business logic.
* Validation rules.
* Data processing.

---

# httpx/

The `httpx` package contains shared HTTP transport utilities used by API handlers.

It provides a consistent way to format API responses and convert internal application errors into HTTP responses.

The purpose of this package is to avoid duplicating response formatting and error handling logic across multiple handlers.

Example:

```text
Handler
   |
   v
Service returns error
   |
   v
httpx.RespondError()
   |
   v
HTTP response
```

---

# httpx/response.go

Responsible for defining the common API response format and response helper functions.

Responsibilities:

* Define standard success response structure.
* Define standard error response structure.
* Provide helper methods for writing HTTP responses.

Example:

```go
type Response struct {
    Code    ErrorCode `json:"code"`
    Message string    `json:"message"`
    Data    any       `json:"data,omitempty"`
}
```

All API responses should follow this structure to keep the external API contract consistent.

Example success response:

```json
{
  "code": "OK",
  "message": "success",
  "data": {
    "id": "inst-7f3a2b"
  }
}
```

Example error response:

```json
{
  "code": "NOT_FOUND",
  "message": "resource not found"
}
```

---

# httpx/error_codes.go

Contains API-level error codes returned to clients.

Responsibilities:

* Define stable error identifiers.
* Provide a consistent contract for API consumers.

Example:

```go
const (
    CodeOK             = "OK"
    CodeNotFound       = "NOT_FOUND"
    CodeInternalError  = "INTERNAL_ERROR"
)
```

Error codes should represent API behavior, not internal implementation details.

For example:

Good:

```text
INSTANCE_NOT_FOUND
```

Avoid:

```text
POSTGRES_OPERATOR_OBJECT_DOES_NOT_EXIST
```

because the internal implementation may change.

---

# httpx/errors.go

Responsible for mapping application errors into HTTP responses.

The service layer defines business errors:

```text
service.ErrNotFound
service.ErrAlreadyExists
```

The HTTP layer converts them:

```text
service.ErrNotFound
        |
        v
httpx.CodeNotFound
        |
        v
HTTP 404
```

Responsibilities:

* Map service errors to HTTP status codes.
* Return standardized error responses.
* Hide internal errors from API consumers.

Example:

```go
func RespondError(c *gin.Context, err error)
```

Handlers should not contain error mapping logic.

Instead of:

```go
if err == service.ErrNotFound {
    c.JSON(404, ...)
}
```

Handlers should use:

```go
if err != nil {
    httpx.RespondError(c, err)
    return
}
```

---

# Why httpx is separate from api

The `api` package contains feature-specific HTTP logic:

* Handlers.
* Routes.
* DTOs.

The `httpx` package contains reusable HTTP utilities:

* Response formatting.
* Error handling.
* HTTP-level conventions.

This prevents duplication when adding more APIs.

Example:

```text
api/
├── instance_handler.go
├── backup_handler.go
└── user_handler.go


httpx/
├── response.go
├── error_codes.go
└── errors.go
```

All handlers share the same response and error behavior.

---

# Adding New Features

When adding a new feature, only add feature-specific errors and mappings.

Example:

Adding a backup feature:

```go
service.ErrBackupNotFound
```

Add mapping:

```go
{
    err: service.ErrBackupNotFound,
    status: http.StatusNotFound,
    code: httpx.CodeBackupNotFound,
}
```

The handler does not need custom error handling.

The flow remains:

```text
Feature Handler
       |
       v
Feature Service
       |
       v
Service Error
       |
       v
httpx Error Mapping
       |
       v
HTTP Response
```

This keeps all APIs consistent as the application grows.

---

# domain/

The domain package contains the core business concepts.

This is the most important layer.

Responsibilities:

* Define business entities.
* Define business states.
* Define service contracts.

The domain should not import:

* Gin
* JSON packages
* Database libraries
* Kubernetes clients
* Logger packages

---

# domain/instance.go

Contains instance-related business models.

Example:

```go
type Instance struct {
    ID string
    Name string
    Status InstanceStatus
}
```

Contains:

* Entities.
* Value objects.
* Domain enums.
* Domain input models.

---

# domain/service.go

Contains service interfaces.

Example:

```go
type InstanceService interface {
    CreateInstance(
        ctx context.Context,
        input CreateInstanceInput,
    ) (*Instance, error)
}
```

Why keep interfaces here?

Because the domain defines what behavior it needs.

The implementation can change without affecting callers.

Example:

Today:

```
InstanceService
        |
        v
Kubernetes implementation
```

Later:

```
InstanceService
        |
        v
Cloud provider implementation
```

The API does not need to change.

---

# service/

The service package contains business use-case implementations.

Responsibilities:

* Execute business workflows.
* Apply business rules.
* Coordinate infrastructure dependencies.
* Handle application errors.

Example:

```
Create Instance

1. Validate input
2. Check existing instance
3. Create Kubernetes resource
4. Store metadata
5. Return domain object
```

---

# service/instance_service.go

Contains the implementation of:

```
domain.InstanceService
```

Example:

```go
type instanceService struct {
    kubernetesClient Client
}
```

This is where business orchestration happens.

---

# service/errors.go

Contains business-level errors.

Examples:

```go
ErrInstanceNotFound

ErrInstanceAlreadyExists
```

Keeping errors centralized allows:

* Consistent error handling.
* Easier testing.
* Cleaner handlers.

---

# config/

Contains application configuration.

Responsibilities:

* Load environment variables.
* Parse configuration.
* Provide application settings.

Example:

```go
type Config struct {
    Port string
    LogLevel string
}
```

Configuration is created during application startup and injected into dependencies.

---

# logger/

Contains application logging setup.

Responsibilities:

* Create logger instance.
* Configure log level.
* Provide common logging behavior.

Services receive logger dependencies instead of creating their own loggers.

---

# infrastructure/

Contains integrations with external systems.

Examples:

```
infrastructure/
├── kubernetes/
├── postgres/
└── cnpg/
```

Responsibilities:

* Communicate with external APIs.
* Hide implementation details.
* Provide interfaces required by services.

The service layer should not directly know low-level client details.

---

# Adding New Features Later

When adding a new feature, organize code by responsibility.

Example new feature:

```
Backup API
```

Expand structure:

```
internal/
├── api/
│   ├── handler/
│   │   ├── instance.go
│   │   └── backup.go
│   ├── router/
│   └── dto.go
│
├── domain/
│   ├── instance.go
│   ├── backup.go
│   └── service.go
│
├── service/
│   ├── instance_service.go
│   ├── backup_service.go
│   └── errors.go
```

The pattern remains the same:

```
Backup Handler
        |
        v
Backup Service Interface
        |
        v
Backup Service Implementation
        |
        v
Infrastructure
```

---

# When the Project Becomes Larger

If the number of features grows significantly, move toward feature-based organization:

```
internal/

├── instance/
│   ├── api.go
│   ├── service.go
│   ├── domain.go
│   └── repository.go
│
├── backup/
│   ├── api.go
│   ├── service.go
│   ├── domain.go
│   └── repository.go
│
├── user/
│   ├── api.go
│   ├── service.go
│   └── domain.go
```

This avoids large shared folders.

---

# Why This Structure Was Chosen

This architecture provides:

## Maintainability

Each layer has a clear responsibility.

## Testability

Services can be tested without HTTP servers or real infrastructure.

## Flexibility

External systems can change without rewriting business logic.

Example:

Changing from one Kubernetes provider to another only affects infrastructure code.

## Scalability

New features can be added by following the same pattern.

---

# Final Guidelines

1. Keep domain independent.
2. Keep handlers thin.
3. Put business rules in services.
4. Put external integrations in infrastructure.
5. Use DTOs only for external communication.
6. Avoid putting reusable code in `internal`.
7. Create `pkg` only when code is intentionally shared by multiple applications.
8. Expand structure by feature when the application becomes large.

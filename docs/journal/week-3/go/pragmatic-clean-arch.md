# Pragmatic Clean Architecture in Go

**Pragmatic Clean Architecture in Go** is basically:

> Keep your business rules independent from frameworks, databases, HTTP, and other infrastructure—but don't turn a small Go service into 40 interfaces and 80 packages.

It's a practical interpretation of Clean Architecture adapted to Go's strengths: **small interfaces, dependency injection, explicit dependencies, simple packages, and boring code.**

### The core idea

Think in terms of dependency direction:

```text
        HTTP / gRPC / CLI
               │
               ▼
        ┌──────────────┐
        │   Delivery   │
        └──────┬───────┘
               │
               ▼
        ┌──────────────┐
        │ Application  │
        │ / Use Cases  │
        └──────┬───────┘
               │
               ▼
        ┌──────────────┐
        │   Domain     │
        │   (business) │
        └──────────────┘
               ▲
               │
        ┌──────┴───────┐
        │ Infrastructure│
        │ DB, APIs, etc │
        └───────────────┘
```

The important part isn't the number of layers. It's that **business logic doesn't know about infrastructure**.

For example, this is good:

```go
type UserRepository interface {
    FindByID(ctx context.Context, id string) (User, error)
}

type UserService struct {
    users UserRepository
}

func (s *UserService) GetUser(ctx context.Context, id string) (User, error) {
    return s.users.FindByID(ctx, id)
}
```

Your use case doesn't care whether `UserRepository` is implemented with PostgreSQL, MongoDB, an HTTP API, or an in-memory map.

The concrete implementation can live elsewhere:

```go
type PostgresUserRepository struct {
    db *sql.DB
}

func (r *PostgresUserRepository) FindByID(
    ctx context.Context,
    id string,
) (User, error) {
    // SQL here
}
```

And wiring happens at the edge:

```go
repo := NewPostgresUserRepository(db)
service := NewUserService(repo)
handler := NewUserHandler(service)
```

That's dependency inversion without a dependency-injection framework.

---

## What makes it "pragmatic"?

Classic Clean Architecture can easily become something like:

```text
entities/
usecases/
interfaces/
repositories/
controllers/
presenters/
gateways/
adapters/
factories/
```

For a CRUD service with five endpoints, that's often ridiculous in Go.

A pragmatic Go project might instead look like:

```text
internal/
├── user/
│   ├── user.go
│   ├── service.go
│   ├── repository.go
│   ├── postgres.go
│   └── handler.go
│
├── order/
│   ├── order.go
│   ├── service.go
│   ├── repository.go
│   ├── postgres.go
│   └── handler.go
│
└── platform/
    ├── postgres/
    └── http/
```

Or, for a smaller application:

```text
internal/
├── domain/
├── application/
├── infrastructure/
└── transport/
```

**There isn't one correct directory structure.** The architecture is primarily about dependency boundaries, not folders.

---

# A useful Go rule

One of the most important pragmatic principles is:

> **Define interfaces where they're consumed, not where they're implemented.**

For example, don't necessarily do this in your repository package:

```go
// repository package
type UserRepository interface {
    FindByID(...)
    Create(...)
    Delete(...)
    Update(...)
    List(...)
    ...
}
```

Instead, your use case defines exactly what it needs:

```go
// user service
type UserFinder interface {
    FindByID(ctx context.Context, id string) (User, error)
}

type UserService struct {
    users UserFinder
}
```

Now your service depends on a tiny interface.

This is particularly idiomatic Go because **interfaces describe what a consumer needs**.

---

# Domain vs application logic

A useful distinction is:

### Domain

Rules that are fundamentally about your business.

```go
type Order struct {
    Items []Item
    Status Status
}

func (o *Order) Cancel() error {
    if o.Status == Shipped {
        return ErrCannotCancelShippedOrder
    }

    o.Status = Cancelled
    return nil
}
```

This shouldn't require:

```go
*sql.DB
http.Request
gin.Context
redis.Client
```

It should be ordinary Go.

### Application / use-case layer

Coordinates domain objects and external dependencies.

```go
func (s *OrderService) CancelOrder(
    ctx context.Context,
    orderID string,
) error {
    order, err := s.orders.FindByID(ctx, orderID)
    if err != nil {
        return err
    }

    if err := order.Cancel(); err != nil {
        return err
    }

    return s.orders.Save(ctx, order)
}
```

The service orchestrates. The domain object owns the business rule.

---

# Where does the database go?

The database is an **implementation detail**.

Your application might depend on:

```go
type OrderRepository interface {
    FindByID(ctx context.Context, id string) (*Order, error)
    Save(ctx context.Context, order *Order) error
}
```

And infrastructure implements it:

```go
type PostgresOrderRepository struct {
    db *sql.DB
}
```

So:

```text
Application
    ↓
OrderRepository interface
    ↑
PostgresOrderRepository
    ↓
PostgreSQL
```

Notice the interesting bit:

**the interface belongs to the application boundary, while PostgreSQL depends on it.**

That's dependency inversion.

---

# HTTP should be thin

A common mistake is putting business logic into handlers:

```go
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    // validate everything
    // check inventory
    // calculate prices
    // talk to database
    // send email
    // ...
}
```

Instead:

```go
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    var req CreateOrderRequest

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        // respond 400
        return
    }

    order, err := h.orders.Create(r.Context(), req)
    if err != nil {
        // map error to HTTP response
        return
    }

    // respond 201
}
```

The handler's job is primarily:

**HTTP → application → HTTP**

Not:

**HTTP → 700 lines of business logic**

---

# Don't abstract prematurely

This is probably the most important "pragmatic" qualification.

You don't need:

```go
type UserServiceFactory interface {}
type UserRepositoryFactory interface {}
type UserMapper interface {}
type UserDTOFactory interface {}
```

just because Clean Architecture diagrams contain abstractions.

Start with concrete code.

Introduce an interface when you have a reason:

* the application needs to invert a dependency
* you need a test double
* multiple implementations exist
* the boundary is genuinely valuable

For example, this can be perfectly fine:

```go
type Service struct {
    db *sql.DB
}
```

If the service genuinely needs SQL semantics and there's no architectural reason to isolate it, **don't invent an interface just for purity**.

---

# Testing becomes straightforward

Because the application layer depends on small interfaces:

```go
type fakeOrderRepo struct {
    order *Order
}

func (r *fakeOrderRepo) FindByID(
    ctx context.Context,
    id string,
) (*Order, error) {
    return r.order, nil
}

func (r *fakeOrderRepo) Save(
    ctx context.Context,
    order *Order,
) error {
    r.order = order
    return nil
}
```

You can test the use case without PostgreSQL, Docker, HTTP, or mocks generated by a framework.

That's one of the biggest practical benefits.

---

## The mental model I'd use

Rather than memorizing Clean Architecture's circles, ask these five questions:

1. **What are my business rules?**
2. **What are my use cases?**
3. **What external things does each use case need?**
4. **Can those dependencies be expressed as small interfaces?**
5. **Can I keep HTTP/DB/framework code at the edges?**

If yes, you're probably doing pragmatic Clean Architecture.

### In one sentence

**Pragmatic Clean Architecture in Go = business logic in plain Go, use cases coordinating that logic, infrastructure implementing interfaces, thin delivery layers, dependency injection at startup, and as little abstraction as the problem actually warrants.**

And importantly, **Clean Architecture is a tool, not a goal**. For a tiny Go service, a simpler 3-package design can be more maintainable than a textbook Clean Architecture implementation.

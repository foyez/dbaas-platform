# Clean code principles

!!! abstract "Where it's used in this project"
    The Go API in [`platform/internal`](https://github.com/foyez/dbaas-platform/tree/main/platform/internal): small interfaces in [`ports/`](https://github.com/foyez/dbaas-platform/tree/main/platform/internal/ports), the error table in [`httpx/errors.go`](https://github.com/foyez/dbaas-platform/blob/main/platform/internal/httpx/errors.go) and the wiring in [`cmd/server/main.go`](https://github.com/foyez/dbaas-platform/blob/main/platform/cmd/server/main.go). The operator splits its reconcile steps across files in [`operator/internal/controller`](https://github.com/foyez/dbaas-platform/tree/main/operator/internal/controller). How these ideas shape the whole codebase is on [Clean architecture](code-design-clean-architecture.md).

## Key ideas

Clean code is code that is easy to read, test and change. SOLID, DRY, KISS and YAGNI are short names for habits that get you there. None of them is a law. Each one solves a specific problem, and each can be overdone.

The examples are in Go. Go has no classes or inheritance, but the principles still apply: read "class" as "type with methods" and "subclass" as "another type that satisfies the same interface".

### SOLID

| Principle | One-line meaning | Warning sign |
|---|---|---|
| **S**ingle responsibility | A type or package has one reason to change | One file changes for unrelated reasons (HTTP, SQL, pricing) |
| **O**pen/closed | Add new cases without editing stable code | A growing `if`/`switch` chain, one branch per new case |
| **L**iskov substitution | Any implementation works wherever the interface is expected | An implementation panics or returns "not implemented" |
| **I**nterface segregation | Many small interfaces beat one big one | Fakes that stub methods the test never calls |
| **D**ependency inversion | Depend on interfaces, not concrete types | A constructor that builds its own database client |

#### Single responsibility

A unit should have one reason to change. "Reason to change" means a kind of requirement, not one method. A useful test: if three different people (product, infra, security) would each ask to edit the same file, it does too much.

The API splits a request across three packages, each changing for one reason:

| Package | Changes when |
|---|---|
| `api/handler` | the HTTP shape changes (headers, JSON, status codes) |
| `service` | a business rule changes (allowed versions, ownership checks) |
| `infra/cnpg` | the way instances are stored in Kubernetes changes |

The operator does the same at file level. `Reconcile` reads like a table of contents and calls one helper per job: `reconcileCNPGCluster` (`cnpg.go`), `reconcileExternalExposure` (`tcpexpose.go`, `portallocator.go`) and `updateStatus`.

SRP doesn't mean one function per type. Keep things together that change together; split only when the reasons really differ.

#### Open/closed

Open for extension, closed for modification: a new case should mean new code, not edits to code that already works.

Mapping errors to HTTP status codes could be a chain of `if errors.Is(...)` branches that grows with every new error. Instead, `httpx` keeps a table and one loop:

```go
var errorMappings = []errorMapping{
    {err: service.ErrNotFound, status: http.StatusNotFound, code: CodeNotFound},
    {err: service.ErrAlreadyExists, status: http.StatusConflict, code: CodeInstanceAlreadyExists},
    // a new error is a new row
}

for _, mapping := range errorMappings {
    if errors.Is(err, mapping.err) {
        Error(c, mapping.status, mapping.code, mapping.err.Error())
        return
    }
}
```

A new error means a new row. The lookup logic stays untouched. OCP doesn't mean "never edit code"; it means protect the stable parts from being edited for every new feature.

#### Liskov substitution

If code works with an interface, it must keep working with every implementation of it. Matching method signatures isn't enough; the behaviour (the contract) must match too.

`ports.InstanceClient` expects `GetInstance` to return `domain.ErrNotFound` when an instance is missing or belongs to someone else, and `infra/cnpg` does exactly that. Any other implementation, including a test fake, must keep that promise. A fake that returned `nil, nil` instead would let a service test pass against behaviour production never shows. Because Go interfaces only check signatures, this contract lives in doc comments and tests, not in the compiler.

Rules of thumb for an implementation:

- Accept at least what the interface accepts.
- Guarantee at least what it promises, including which errors it returns.
- Don't panic or return "not supported" for a method you claim to implement. If you can't do it, you shouldn't satisfy that interface; split it instead (next section).

#### Interface segregation

Clients shouldn't depend on methods they don't use. In Go this comes naturally: define small interfaces where they're used.

`ports.Querier` has two methods, `ExecuteQuery` and `Close`. The SQL console only needs those, and its mock is about 30 lines. `ports.InstanceClient` has seven, because the instance service really uses all of them. A test that only covers `DeleteInstance` still has to work with a mock for the rest, which is the cost of a bigger interface.

ISP vs SRP: SRP is about what a type *does*; ISP is about what it *exposes* to each client.

#### Dependency inversion

High-level logic shouldn't depend on low-level details. Both should depend on an interface, and the interface belongs to the high-level side.

```go
// service/instance.go: depends on interfaces from ports, not on Kubernetes
type instanceService struct {
    client  ports.InstanceClient
    querier ports.Querier
    ...
}
```

The service never calls `cnpg.NewClient` itself. `main.go` builds the real implementations and passes them in:

```go
instanceClient := cnpg.NewClient(k8sClient, cfg.Server.PublicDBHost)
querier := pgquery.NewPool()
svc := service.NewInstanceService(instanceClient, querier, lokiClient, log)
```

| Term | What it is |
|---|---|
| Dependency inversion (DIP) | The principle: depend on abstractions |
| Dependency injection (DI) | The technique: pass dependencies in from outside, usually through the constructor |

DI is how you apply DIP. In Go you rarely need a DI framework; plain constructor arguments in `main` are enough.

### DRY, KISS and YAGNI

| Principle | Goal | Common mistake |
|---|---|---|
| **DRY** (don't repeat yourself) | One source of truth for each rule | Merging code that only *looks* alike |
| **KISS** (keep it simple) | Remove complexity you don't need | Calling skipped error handling "simple" |
| **YAGNI** (you aren't gonna need it) | Don't build for imagined needs | Ignoring requirements that are already confirmed |

#### DRY: one place per rule

DRY is about repeated *knowledge* (a rule, a decision), not repeated lines. The test: if the rule changes, how many places must you edit?

The supported PostgreSQL versions are a rule. Today it lives in several places:

| Where | How |
|---|---|
| `api/dto.go` (create and update) | `binding:"oneof=14 15 16"` |
| `service/instance.go` (create and update) | range check against `14` and `16` |
| `sdk/dbaas/dbaastest/server.go` | the same range in the fake API |

Adding PostgreSQL 17 means editing all five, and missing one gives an API that accepts a version the service rejects. One constant in `domain` (for example `SupportedVersions`) used by the service, with the DTO tag kept only as a fast first check, would make it one edit.

DRY also applies outside code: one config value instead of the same URL in several manifests, one doc page that others link to instead of copied explanations.

#### The wrong abstraction

Two pieces of code can look identical today and still be different knowledge. `CreateInstance` and `UpdateInstance` both validate a version, but an update might later forbid downgrades while create allows any version. Forcing them into one generic `validate(input any)` couples rules that should change separately.

> Duplication is cheaper than the wrong abstraction. Merge on shared meaning, not on shared shape.

The **rule of three** balances DRY against YAGNI: write it once, notice it the second time, extract it the third time.

#### KISS: clear beats clever

Write the simplest code that is still correct. Go's culture already pushes this way: early returns instead of nested conditions, a plain loop instead of a clever helper.

```go
// Hard to scan
if input.Name != "" && input.Storage != "" && input.Version >= 14 && input.Version <= 16 {
    // ...the real work, indented
}

// Easy to scan: guard clauses, one rule per line
if input.Name == "" {
    return ErrInvalidInput
}
if input.Version < 14 || input.Version > 16 {
    return ErrInvalidInput
}
```

The second form is what `validateCreateInstance` does. Simple doesn't mean naive: validation, error handling and timeouts are part of being correct, not extra complexity.

#### YAGNI: build it when you need it

Don't add features, options or layers "just in case". Each one costs reading time, tests and bugs, and the guess is often wrong.

- **Interfaces.** Add one for a real reason: inverting a dependency, a test double, or a second implementation. `ports.Querier` has one real implementation (`pgquery`) and still earns its interface: it keeps the pgx driver out of the service, and it lets a test swap in `ports/mocks.MockQuerier`.
- **Caching.** Add it after a measurement shows a need. A cache brings stale data and invalidation logic with it.
- **Stub comments.** Both `ports/client.go` and `domain/service.go` still end with a `// Future operations: GetInstance(...)` list. Those methods now exist, so the comment is noise. Write the code when it's needed, and delete plans once they're done.

YAGNI doesn't mean skipping the basics. Logging, validation and graceful shutdown aren't speculative.

## What I learned building it

- **Small interfaces fall out of "define them where they're used".** `ports.Querier` holds only what the service calls, so its mock is tiny. The seven-method `InstanceClient` is the one I'd consider splitting, for example into a reader and a writer.
- **A table is often the simplest form of open/closed.** The error-mapping slice in `httpx` handles every new error with one new row, and it kept the HTTP status logic out of the service.
- **Repeated literals are a DRY smell even when they're tiny.** `14` and `16` look harmless, but they're a product decision copied into five places, including the SDK's fake server.
- **Interfaces only pay off if tests use them.** The mocks in `ports/mocks` and `domain/mocks` exist, but `service/instance_test.go` and `handler/instance_test.go` are fully commented out and still call the old two-argument `NewInstanceService`. The design makes testing easy; it doesn't do the testing.
- **Dead comments rot faster than code.** The "Future operations" lists became wrong the moment the methods were written, and nothing reminded me to delete them.

## Review questions

??? question "SRP says 'one reason to change'. What counts as a reason?"
    A kind of requirement coming from one source: an HTTP contract change, a business rule change, a storage change. It's not "one method". A type can have many methods and still have one reason to change.

??? question "How does the error-mapping table in httpx follow open/closed?"
    New errors are added as new rows in `errorMappings`. The loop that checks `errors.Is` and writes the response never changes, so adding an error can't break the existing mappings.

??? question "Go has no inheritance. How can Liskov substitution still be violated?"
    Through interfaces. Any type that satisfies an interface but breaks its contract violates LSP: returning `nil, nil` instead of `ErrNotFound`, panicking with "not implemented", or ignoring a context deadline the interface implies it honours.

??? question "What's the difference between dependency inversion and dependency injection?"
    Inversion is the principle: high-level code depends on an interface it owns, not on a concrete low-level type. Injection is the technique for supplying the implementation from outside, e.g. passing it to `NewInstanceService` in `main.go`.

??? question "Two functions have identical bodies. Should you always merge them?"
    No. Merge only if they encode the same rule and will change together. If they belong to different features that just happen to look alike today, merging creates the wrong abstraction, which is costlier than duplication.

??? question "When is adding an interface in Go justified, according to YAGNI?"
    When there's a real reason today: the consumer must not depend on the concrete package (dependency inversion), a test needs a fake, or a second implementation exists. "We might swap it one day" isn't enough.

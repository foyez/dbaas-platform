# Pragmatic Clean Architecture

Uncle Bob's original model is four concentric rings.

```
┌───────────────────────────────────────────────────────────────┐
│                 FRAMEWORKS & DRIVERS  (outermost)             │
│   web framework, DB driver, external APIs, UI, devices        │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                 INTERFACE ADAPTERS                      │  │
│  │   controllers, presenters, gateways, repositories       │  │
│  │  ┌───────────────────────────────────────────────────┐  │  │
│  │  │                   USE CASES                       │  │  │
│  │  │      application-specific business rules          │  │  │
│  │  │  ┌─────────────────────────────────────────────┐  │  │  │
│  │  │  │                  ENTITIES                   │  │  │  │
│  │  │  │     enterprise-wide business rules & models │  │  │  │
│  │  │  └─────────────────────────────────────────────┘  │  │  │
│  │  └───────────────────────────────────────────────────┘  │  │
│  └─────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘

        Dependency Rule:  source code dependencies may only
        point INWARD  ( outer ──▶ inner ), never outward.
```

- **Entities (innermost)** — the plain data + rules that would still make sense even if this weren't a web app at all. "A PostgreSQL instance has a name, a version, and a status" is true regardless of whether you expose it over HTTP, gRPC, or a CLI.
- **Use Cases** — the specific things *this application* does with those entities: "create an instance," "reject a request with an invalid version." This layer orchestrates entities to do something, and defines the *interfaces* (ports) it needs from the outside world — e.g. "something that can persist an instance" — without knowing what that something actually is.
- **Interface Adapters** — translators. On one side they speak the use case's language; on the other side they speak someone else's format: HTTP JSON in, use-case input out (a controller); use-case output in, a database CRD out (a gateway/repository implementation). Adapters are where interfaces defined by Use Cases get *implemented*.
- **Frameworks & Drivers (outermost)** — the actual libraries and infrastructure doing I/O: your web framework, your DB driver, the real Kubernetes API. Deliberately kept thin and swappable.

The trick that makes the "inward-only" rule survive contact with reality: when an inner ring (Use Cases) needs something an outer ring provides (like talking to a database), the inner ring **defines an interface for it**, and the outer ring **implements that interface**. Control still flows outward at runtime, but the *import/compile-time dependency* still points inward, because the interface lives with the inner ring. This is Dependency Inversion, and it's the mechanism, not just a principle.


# Components

The platform is a handful of small, separately deployed parts. Each page covers what a part does, how it's built and where its code lives.

| Component | What it does | Code |
|---|---|---|
| [Operator](operator.md) | Turns `PostgreSQL` custom resources into CloudNativePG clusters and exposes them over TCP | [`operator/`](https://github.com/foyez/dbaas-platform/tree/main/operator) |
| [REST API](api.md) | Authenticated CRUD, credentials, logs and the SQL console | [`platform/`](https://github.com/foyez/dbaas-platform/tree/main/platform) |
| [Web UI](frontend.md) | Vue app for the whole instance lifecycle | [`client/`](https://github.com/foyez/dbaas-platform/tree/main/client) |
| [Authentication](auth.md) | ZITADEL login, roles and token checks | [`infra/zitadel/tf`](https://github.com/foyez/dbaas-platform/tree/main/infra/zitadel/tf) |
| [Go SDK](sdk.md) | Typed client for the REST API | [`sdk/`](https://github.com/foyez/dbaas-platform/tree/main/sdk) |
| [Delivery](delivery.md) | CI, GitOps and Argo CD | [`.forgejo/`](https://github.com/foyez/dbaas-platform/tree/main/.forgejo), [`gitops/`](https://github.com/foyez/dbaas-platform/tree/main/gitops) |
| [Observability](observability.md) | Metrics, logs and the audit trail | [`gitops/observability`](https://github.com/foyez/dbaas-platform/tree/main/gitops/observability) |
| [Infrastructure](infrastructure.md) | Clusters on OpenStack and STACKIT SKE | [`infra/`](https://github.com/foyez/dbaas-platform/tree/main/infra) |

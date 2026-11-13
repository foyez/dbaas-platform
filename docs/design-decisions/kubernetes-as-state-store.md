# 1. The Kubernetes API Is the Only State Store

No separate platform database exists. Each instance is a `PostgreSQL` custom resource, and ownership is a label on it (`app.example.com/user-id`).

**Why:** etcd already provides persistence, watches and optimistic concurrency. A second store would be a second source of truth that can drift from the cluster.

**Trade-off:** queries are limited to label selectors. Pagination is cursor-only, because the List API has no offset. Anything relational, such as per-user quotas or billing, would need a real database later.

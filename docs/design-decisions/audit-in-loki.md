# 7. Audit Events Are Structured Logs in Loki

They don't live in a database table.

**Why:** it reuses the observability stack that already exists. Audit becomes one more `component=audit` log stream, queryable by the admin endpoint.

**Trade-off:** audit retention equals log retention (7 days), and the logs are not tamper-evident. A compliance-grade audit trail would need its own append-only storage.

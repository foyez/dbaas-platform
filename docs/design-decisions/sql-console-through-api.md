# 5. The SQL Console Goes Through the API

The browser never connects to PostgreSQL directly.

**Why:** browsers can't speak the PostgreSQL wire protocol. Going through the API reuses the same authentication and ownership checks. Each query is limited to 10 seconds and 500 rows, with at most 3 pooled connections per instance and idle eviction.

**Trade-off:** the API becomes part of the data path and holds per-pod connection state. Queries run as the database owner.

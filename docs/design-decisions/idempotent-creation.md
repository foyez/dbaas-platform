# 4. Creation Is Idempotent Through a Deterministic Resource Name

`sha256(userID + Idempotency-Key)`

**Why:** the API server's `AlreadyExists` check becomes the atomic deduplication step. No idempotency table is needed, and it works across API replicas.

**Trade-off:** clients must send a key on every create. Reusing a key with a different body isn't detected yet.

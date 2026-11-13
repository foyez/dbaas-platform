# 6. Authentication Is Delegated to ZITADEL

Access tokens are checked through introspection.

**Why:** OIDC, roles, social login and service accounts (used by the SDK) come for free and are maintained by a dedicated product.

**Trade-off:** every request makes a network call to an external dependency. Local JWT validation would be faster, but token revocation would then take effect later.

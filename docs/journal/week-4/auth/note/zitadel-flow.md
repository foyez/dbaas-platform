# Authentication Architecture Prep Guide
### Vue → ZITADEL → Google → Go API (OAuth 2.0 / OIDC / PKCE)

---

## 1. The One-Paragraph Answer (memorize this)

> "My PaaS has a Vue frontend and a Go backend API. I use ZITADEL as the identity provider. The Vue app is an OIDC client using the Authorization Code Flow with PKCE, since a browser SPA can't safely store a client secret. When a user isn't authenticated, Vue redirects to ZITADEL, which authenticates the user directly or via Google as an external identity provider. ZITADEL redirects back with an authorization code, which Vue exchanges — along with the PKCE code verifier — for tokens. The ID token identifies the user; the access token is sent to my Go API as a Bearer token. The Go API validates the token's signature, issuer, audience, and expiration, then checks the user's role (admin/user) to authorize the request. The frontend can use roles for UI decisions, but real authorization always happens in the Go API, since the frontend can't be trusted for security."

If you can say this fluently and then explain *any* piece in more depth, you're ready.

---

## 2. The Big Picture (whiteboard diagram)

```
┌──────────┐        ┌─────────────────┐        ┌──────────┐
│  Google  │◄───────┤     ZITADEL     │◄──────►│   Vue    │
│   IdP    │        │  Auth Server /  │  OIDC  │ Frontend │
└──────────┘        │      IdP        │        │ (client) │
                    └────────┬────────┘        └────┬─────┘
                             │                      │ access token
                             │                      ▼
                             │                  ┌──────────┐
                             └─────────────────►│  Go API  │
                                                │ (resource│
                                                │  server) │
                                                └────┬─────┘
                                                     ▼
                                                 Database
```

**Story in 11 steps:**
1. User opens Vue app.
2. Vue checks if user is authenticated.
3. If not, redirect to ZITADEL.
4. ZITADEL authenticates (directly, or via Google).
5. ZITADEL returns an authorization code to Vue.
6. Vue exchanges code + PKCE verifier for tokens.
7. Vue stores tokens.
8. Vue calls Go API with the **access token**.
9. Go validates the token.
10. Go extracts user identity + role.
11. Go allows/denies the operation.

---

## 3. Core Concepts — OAuth vs OIDC

| | OAuth 2.0 | OIDC |
|---|---|---|
| Solves | **Authorization** — "Can this app access this resource?" | **Authentication** — "Who is this user?" |
| Built on | Base framework | Built *on top of* OAuth 2.0 |
| Introduces | Access tokens | ID tokens (identity claims) |
| Key phrase | Delegated authorization | Identity layer over OAuth |

**Interview line:**
> "OAuth 2.0 provides delegated authorization, while OpenID Connect adds an identity layer on top of OAuth 2.0 so the client can authenticate the user."

### The Four Actors

| Role | In your architecture |
|---|---|
| Resource Owner | The user |
| Client | Vue frontend |
| Authorization Server | ZITADEL |
| Resource Server | Go API |

---

## 4. Authorization Code Flow + PKCE

**Why PKCE?** Your Vue app is a **public client** — it runs in the browser, so it can't hide a `client_secret` (anyone can inspect it). PKCE replaces the secret with a one-time, dynamically generated proof. It protects the OAuth/OIDC **authorization code** from being stolen and reused by an attacker.

Without PKCE:

```text
Browser → ZITADEL → authorization code
                         │
                    attacker steals it
                         │
                         ▼
                  exchange for token ❌
```

With PKCE:

```text
1. App creates secret random value
   code_verifier = "random-secret"

2. App sends a hash of it to ZITADEL
   code_challenge = SHA256(code_verifier)

3. ZITADEL returns authorization code

4. App exchanges:
   authorization code
   +
   original code_verifier
        │
        ▼
      ZITADEL
        │
        ▼
   Access token ✅
```

ZITADEL checks:

```text
SHA256(code_verifier) == code_challenge ?
```

If yes → issue token.

If an attacker steals only the authorization code, they **don't have the `code_verifier`**, so they can't exchange the code for a token.

### What problem does it solve?

> **PKCE prevents an intercepted OAuth authorization code from being exchanged for an access token.**

It's especially important for **SPAs and mobile apps**, because they are public clients and cannot safely keep a client secret.

Your Vue + ZITADEL application should use:

```text
Authorization Code Flow + PKCE
```

which is the modern recommended approach.


**PKCE = Proof Key for Code Exchange**

| Step | What happens |
|---|---|
| 1 | Vue generates a random `code_verifier` (kept locally, never sent yet) |
| 2 | Vue derives `code_challenge = BASE64URL(SHA256(code_verifier))` |
| 3 | Vue sends the `code_challenge` with the authorization request |
| 4 | ZITADEL authenticates the user |
| 5 | ZITADEL returns an `authorization_code` |
| 6 | Vue sends `authorization_code` + original `code_verifier` |
| 7 | ZITADEL re-hashes the verifier and compares it to the stored challenge — if it matches, tokens are issued |

**Why it's secure:** if an attacker intercepts only the authorization code, they don't have the `code_verifier`, so they can't complete the exchange.

**Interview line:**
> "PKCE binds the authorization code exchange to the client that initiated the request, by requiring a secret verifier that was never sent in the initial authorization request."

---

## 5. Tokens — Know the Difference

| Token | Purpose | Used by | Example claims |
|---|---|---|---|
| **ID Token** | "Who authenticated?" | Vue (frontend identity) | `sub`, `email`, `name` |
| **Access Token** | "What can be accessed?" | Sent to Go API as `Authorization: Bearer ...` | `aud`, `scope`, roles |
| **Refresh Token** | Get a new access token without re-login | Vue (if `offline_access` scope used) | — |

⚠️ **Common mistake to avoid:** don't say the frontend sends the ID token to the API. The API should receive an **access token** meant for it (validated via `aud`).

### Scopes you're using
- `openid` — required, signals this is an OIDC (not plain OAuth) request
- `profile` — basic profile info
- `email` — user's email
- `offline_access` — allows refresh tokens
- `urn:zitadel:iam:org:project:id:{project_id}:aud` — adds your API as a token audience so the API can verify the token was meant for it

---

## 6. What the Go API Must Validate

When a request comes in with `Authorization: Bearer <token>`, the API checks:

| Check | Question it answers |
|---|---|
| **Signature** | Was this token issued by ZITADEL and untampered? (verified via public key / JWKS) |
| **Issuer (`iss`)** | Did it come from your ZITADEL instance? |
| **Audience (`aud`)** | Was it intended for your API? |
| **Expiration (`exp`)** | Is it still valid? |
| **Roles/claims** | Is this user allowed to do this action? |

```
Request → Bearer token → validate signature/iss/aud/exp
        → extract claims/roles → authorization middleware
        → requireRole("admin")? → allow / 403
```

**401 vs 403:**
- `401 Unauthorized` → not properly authenticated (missing/invalid token)
- `403 Forbidden` → authenticated, but lacks permission

---

## 7. Roles & Authorization

- Roles (`admin`, `user`) come from ZITADEL claims embedded in the access token.
- The **frontend can use roles for UI only** (e.g., hide a delete button) — this is UX, not security.
- The **Go API must independently enforce** authorization, because an attacker can call the API directly, bypassing the Vue UI entirely.

```
Foyez → authenticated → role: admin → can create/delete resources
John  → authenticated → role: user  → blocked from admin actions
```

---

## 8. Google Login (as External IdP)

Google does **not** talk to Vue directly — ZITADEL brokers it.

```
Vue → ZITADEL → user picks "Google" → Google authenticates
    → Google returns identity to ZITADEL
    → ZITADEL creates/links user, issues its own tokens
    → Vue
```

Your app never implements Google's OAuth logic itself — ZITADEL abstracts that away. Depending on config, ZITADEL either creates a new user or links to an existing one by identity.

---

## 9. Why Two ZITADEL Applications?

Be ready to explain the actual reason from your setup — typically:
- One application = the **Vue frontend** (public client, PKCE)
- Other application = represents the **API/project audience** so the Go API can validate tokens meant for it

Also know the hierarchy:

```
ZITADEL Instance
   └── Project
         ├── Application (Vue - public, PKCE)
         ├── Application (API-facing)
         └── Roles (admin, user)
```

---

## 10. Security Topics to Expect

| Topic | Key point |
|---|---|
| **Token storage** | Know exactly what your app does — tradeoffs between `localStorage`, `sessionStorage`, memory, or HttpOnly cookies |
| **Stolen access token** | Attacker can use it until expiry — mitigate with short lifetimes, HTTPS, audience/scope checks |
| **XSS** | Malicious JS runs in your app → can steal tokens accessible to JS |
| **CSRF** | Browser tricked into sending authenticated requests — more relevant with cookie-based auth |
| **CORS** | Browser security mechanism (not authentication) that governs cross-origin requests between `frontend.example.com` and `api.example.com` |
| **JWT** | `header.payload.signature` — payload is **encoded, not encrypted** → never put secrets in claims |
| **Why not build auth yourself** | Password storage, resets, MFA, session mgmt, OAuth/OIDC flows, Google integration, key rotation — huge surface area ZITADEL already solves |

---

## 11. Quick Glossary

**OAuth/OIDC:** OAuth 2.0, OpenID Connect, Authorization Server, Resource Server, Client, Resource Owner
**Flow:** Authorization Code, PKCE, code verifier, code challenge, redirect URI
**Tokens:** access token, ID token, refresh token, JWT, claims, scope, `aud`, `iss`, `sub`, `exp`, JWKS
**Security:** HTTPS, CORS, CSRF, XSS, public vs confidential client
**ZITADEL:** Instance, Organization, Project, Application, External IdP, Role

---

## 12. Practice Questions (say answers out loud, no notes)

**Fundamentals**
1. What is OAuth 2.0? Is it authentication or authorization?
2. What is OIDC and why was it created?
3. Difference between access token, ID token, refresh token?

**PKCE**
4. What is PKCE and why do you need it for a Vue SPA?
5. What are code_verifier and code_challenge?
6. What happens if an attacker steals the authorization code alone?

**Your Architecture**
7. Walk through your full login flow, start to finish.
8. Why two ZITADEL applications?
9. Which piece is the authorization server? The resource server?

**Backend**
10. How does your Go API validate a token?
11. Difference between 401 and 403?
12. Why can't you trust a role sent by the frontend?

**Security**
13. Where are tokens stored, and why?
14. What's the risk if an access token is stolen?
15. What's the difference between XSS, CSRF, and CORS?

---

## 13. Before the Interview — Know Your Actual Config

Don't just know the theory — be ready to point to your real setup:

- Vue: client ID, redirect URI, scopes, PKCE config, token storage
- Go API: middleware, issuer, audience check, role extraction, authorization logic

Interviewers often chain: *"What is PKCE?" → "How did you implement it?" → "Where's the verifier generated?" → "How does your API check roles?"* — being able to follow that whole chain is the real goal.
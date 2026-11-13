# PKCE — Proof Key for Code Exchange

## 1. What is PKCE?

**PKCE (Proof Key for Code Exchange)** is a security mechanism used with **OAuth 2.0 Authorization Code Flow**.

It protects the authorization process from **authorization code interception attacks**.

In simple terms:

> **PKCE ensures that the application requesting an authorization code is the same application that later exchanges that code for an access token.**

PKCE is pronounced **"pixy."**

---

# 2. Why do we need PKCE?

To understand PKCE, first understand the normal OAuth Authorization Code Flow.

Normally, the flow looks like this:

```text
User
  |
  | Login
  v
Application
  |
  | Authorization Request
  v
Authorization Server
  |
  | Authorization Code
  v
Application
  |
  | Exchange Code
  v
Access Token
```

The problem is that the **authorization code travels back to the application**.

If an attacker manages to intercept that authorization code, they may try to exchange it for an access token.

For example:

```text
User
  |
  v
Your App -----> Authorization Server
  |                    |
  |<--- Authorization Code
  |
  X  Attacker intercepts code
  |
  v
Attacker -----> Authorization Server
                    |
                    v
                Access Token
```

The attacker doesn't necessarily need the user's password. If the authorization server accepts the stolen code, the attacker could potentially obtain tokens.

**PKCE solves this problem.**

---

# 3. The basic idea behind PKCE

PKCE introduces two values:

* **Code Verifier**
* **Code Challenge**

The application generates a random secret called the **code verifier**.

It then creates a derived value called the **code challenge** and sends the challenge to the authorization server.

Later, when exchanging the authorization code for a token, the application sends the original verifier.

The authorization server checks:

```text
Does SHA256(code_verifier)
equal
the code_challenge I received earlier?
```

If yes, the application is allowed to exchange the code for tokens.

If not, the request is rejected.

So, even if an attacker steals the authorization code, **the authorization code alone isn't sufficient**.

---

# 4. How PKCE works step by step

## Step 1 — Generate a Code Verifier

The client application generates a cryptographically random string.

For example:

```text
code_verifier =
dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk
```

This value should be kept secret by the client during the authorization flow.

---

## Step 2 — Generate the Code Challenge

The application derives the code challenge from the verifier.

With the recommended `S256` method:

```text
code_challenge =
BASE64URL(
    SHA256(code_verifier)
)
```

So we have:

```text
code_verifier
      |
      | SHA-256
      v
code_challenge
```

The important point is:

> The code challenge can be sent to the authorization server, but the original code verifier is kept by the client.

---

# 5. Send the Authorization Request

The application sends the user to the authorization server.

The request contains the code challenge:

```http
GET /authorize?
    response_type=code
    &client_id=my-client
    &redirect_uri=https://myapp.com/callback
    &code_challenge=...
    &code_challenge_method=S256
```

Notice that the **code verifier is not sent** here.

Only the challenge is sent.

---

# 6. User authenticates

The authorization server asks the user to authenticate.

For example:

```text
Username
Password
    ↓
MFA
    ↓
User approves access
```

After successful authentication and consent, the authorization server sends an **authorization code** back to the application.

For example:

```text
https://myapp.com/callback?code=ABC123
```

---

# 7. Exchange the Authorization Code

Now the application sends the authorization code to the token endpoint.

But this time it also sends the **original code verifier**.

```http
POST /token

grant_type=authorization_code
code=ABC123
client_id=my-client
redirect_uri=https://myapp.com/callback
code_verifier=dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk
```

---

# 8. Authorization Server Verifies PKCE

The authorization server takes the received verifier and calculates:

```text
SHA256(code_verifier)
```

It then compares the result with the challenge that was stored earlier.

Conceptually:

```text
SHA256(received_code_verifier)
              |
              v
       Calculated Challenge
              |
              | compare
              v
Stored Code Challenge
```

If they match:

```text
✅ PKCE verification successful
       ↓
Authorization Code accepted
       ↓
Access Token issued
```

If they don't match:

```text
❌ PKCE verification failed
       ↓
Request rejected
```

---

# 9. Complete PKCE Flow

The complete process can be explained like this:

```text
                CLIENT / APPLICATION
                       |
                       |
             1. Generate verifier
                       |
                       v
                code_verifier
                       |
                       | SHA-256
                       v
                code_challenge
                       |
                       |
                       | 2. Authorization Request
                       |    + code_challenge
                       v
              AUTHORIZATION SERVER
                       |
                       | 3. User Login
                       | 4. User Consent
                       |
                       | 5. Authorization Code
                       v
                CLIENT / APPLICATION
                       |
                       | 6. Send:
                       |    authorization_code
                       |    code_verifier
                       v
              AUTHORIZATION SERVER
                       |
                       | 7. Verify:
                       |    SHA256(verifier)
                       |       ==
                       |    stored challenge
                       |
                       v
                  ACCESS TOKEN
```

---

# 10. What happens if an attacker steals the authorization code?

This is the key reason PKCE exists.

Suppose an attacker somehow obtains:

```text
authorization_code = ABC123
```

They try:

```text
Attacker → /token
           code=ABC123
```

But the authorization server requires the correct:

```text
code_verifier
```

The attacker doesn't have it.

Therefore:

```text
Authorization Code
       +
Correct Code Verifier
       ↓
    Required
       ↓
Access Token
```

Having only:

```text
Authorization Code
```

is not enough.

---

# 11. Why not just use a Client Secret?

A common interview question is:

**"If we already have a client secret, why do we need PKCE?"**

The answer depends on the type of application.

A traditional backend application can securely keep a client secret because the secret stays on the server.

But applications such as:

* Mobile apps
* Single-page applications
* Desktop applications

cannot reliably keep a client secret secret.

For example, if you put a client secret inside a mobile application:

```text
Mobile App
   |
   +-- client_secret
```

Someone can potentially reverse-engineer the application and extract it.

Therefore, a client secret isn't a reliable way to identify a particular installation of a public client.

PKCE provides protection without requiring the client to permanently store a secret.

---

# 12. PKCE vs Client Secret

| Client Secret                               | PKCE                                                         |
| ------------------------------------------- | ------------------------------------------------------------ |
| Long-term credential                        | Short-lived per-authorization secret                         |
| Used to authenticate a client               | Used to bind the authorization request to the token exchange |
| Can be securely stored by a backend         | Suitable for public clients                                  |
| Not suitable as a secret in mobile/SPA apps | Designed to protect these flows                              |
| Usually static                              | Generated dynamically for each authorization request         |

An important distinction:

> **PKCE does not necessarily replace client authentication.**

A confidential backend application may use **both a client secret and PKCE**.

---

# 13. What is `S256`?

PKCE supports a `code_challenge_method`.

The recommended method is:

```text
S256
```

It means:

```text
code_challenge = BASE64URL(SHA256(code_verifier))
```

There was also a `plain` method, where the challenge is essentially the verifier itself.

Conceptually:

```text
plain:
code_challenge = code_verifier

S256:
code_challenge = BASE64URL(SHA256(code_verifier))
```

**S256 should be preferred** because the verifier itself is not exposed during the authorization request.

---

# 14. PKCE does NOT encrypt the authorization code

This is an important point to mention in an interview.

PKCE does **not** encrypt the authorization code.

Instead, it creates a relationship between:

```text
Authorization Request
        |
        | code_challenge
        v
Authorization Server
        |
        | authorization code
        v
Token Request
        |
        | code_verifier
        v
Verification
```

The authorization server uses this relationship to prove that the party exchanging the code possesses the original verifier.

---

# 15. PKCE vs OAuth — what's the difference?

Another common interview question:

**"Is PKCE an authentication protocol?"**

No.

OAuth 2.0 is an **authorization framework**.

PKCE is an **extension to the OAuth Authorization Code Flow** that improves its security.

Think of it this way:

```text
OAuth 2.0
   |
   +--- Authorization Code Flow
           |
           +--- PKCE
```

And if you're doing user authentication with OAuth, you may encounter **OpenID Connect (OIDC)** on top of OAuth.

---

# 16. What security problem does PKCE solve?

The main problem is:

### Authorization Code Interception

Without PKCE:

```text
Authorization Code
        |
        +----> Legitimate App
        |
        +----> Attacker
```

If the attacker gets the code, they may attempt to redeem it.

With PKCE:

```text
Authorization Code
        |
        +----> Legitimate App
        |          +
        |       Code Verifier
        |
        +----> Attacker
                   |
                   X
              No verifier
```

Therefore, the attacker cannot successfully redeem the code.

---

# 17. PKCE in a real-world example

Imagine you're using a mobile application and choose:

```text
Login with Google
```

The mobile app starts an OAuth authorization flow.

Before doing so, it generates:

```text
code_verifier = random-secret
```

Then:

```text
code_challenge = SHA256(code_verifier)
```

It sends the challenge to the authorization server.

After you log in, the authorization server returns an authorization code.

The mobile app then sends:

```text
authorization_code
+
code_verifier
```

The authorization server verifies the verifier and returns the tokens.

If another malicious application intercepted the authorization code, it would not have the original verifier.

That's the protection PKCE provides.

---

# 18. Interview-ready explanation

If an interviewer asks:

### "What is PKCE?"

You can answer:

> **PKCE stands for Proof Key for Code Exchange. It is a security extension to OAuth 2.0 Authorization Code Flow that protects against authorization code interception. The client generates a random code verifier and derives a code challenge from it, usually using SHA-256. The challenge is sent during the authorization request, while the original verifier is sent later when exchanging the authorization code for tokens. The authorization server verifies that the verifier matches the original challenge before issuing tokens. This ensures that simply stealing the authorization code is not enough to obtain an access token. PKCE is especially important for public clients such as mobile apps, SPAs, and desktop applications.**"

That's a strong **30–45 second interview answer**.

---

# 19. One-minute presentation explanation

You can present it like this:

> "PKCE, or Proof Key for Code Exchange, is a security mechanism used with OAuth 2.0 Authorization Code Flow. Its main purpose is to prevent an attacker from using a stolen authorization code to obtain an access token.
>
> The client first generates a random code verifier and creates a code challenge from it using SHA-256. The client sends the challenge when starting the authorization flow. After the user authenticates, the authorization server returns an authorization code. When the client exchanges that code for tokens, it sends the original code verifier. The authorization server calculates the challenge again and compares it with the original one. If they match, the token is issued; otherwise, the request is rejected.
>
> The key advantage is that an attacker who only intercepts the authorization code cannot redeem it because they don't have the code verifier. That's why PKCE is particularly important for mobile, SPA, and desktop applications."

---

# 20. Key points to remember

For an interview, remember these **7 points**:

1. **PKCE = Proof Key for Code Exchange**
2. It is an **OAuth 2.0 security extension**
3. It protects the **Authorization Code Flow**
4. It prevents **authorization code interception/replay**
5. It uses **code_verifier + code_challenge**
6. `S256` is the recommended challenge method
7. It is especially important for **mobile apps, SPAs, and other public clients**

### The easiest way to remember it

Think:

```text
Code Challenge
      ↓
"Whoever starts this flow must prove
they have the original secret later."
      ↓
Code Verifier
      ↓
Token
```

**One-line summary:**

> **PKCE makes an authorization code useless to an attacker who doesn't possess the original code verifier.**

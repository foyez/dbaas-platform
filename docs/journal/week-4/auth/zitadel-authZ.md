# ZITADEL Authentication & Two-Role Authorization

This API uses ZITADEL for authentication and simple role-based authorization.

There are two application roles:

* `admin` — administrator
* `user` — normal authenticated user

ZITADEL is the source of truth for users and application roles. The API does not need a separate users/roles database.

## 1. Create the project roles

In the ZITADEL Console:

1. Open your **Project**.
2. Go to **Roles**.
3. Click **Add Role**.
4. Create:

| Key     | Display Name  | Group |
| ------- | ------------- | ----- |
| `admin` | Administrator | empty |
| `user`  | User          | empty |

The **Key** is what the API uses when checking authorization.

For example:

```go
auth.RequireRole("admin")
```

## 2. Assign roles to users

Go to the project's **Role Assignments** and create an assignment.

For your own account:

```text
User: your ZITADEL account
Role: admin
```

For normal users:

```text
User: another user
Role: user
```

You do not need to assign both roles to a user.

The intended model is:

```text
You             → admin
Normal users    → user
```

## 3. Return roles during authentication

Creating and assigning roles is not enough. The role information must be returned to your application during authentication.

In the ZITADEL Console, open your application and configure the authentication/token settings so that **project roles are asserted/returned**.

Enable:

**Assert Roles on Authentication**

ZITADEL can provide role information during authentication. For OpenID Connect, roles can be requested from the UserInfo endpoint or included in the ID Token or Access Token, depending on the application's configuration.

For this API, the important point is that the access token/introspection result used by the API must contain the user's project roles.

After changing role assignments or authentication settings, obtain a **new access token** before testing. Existing tokens may not contain the newly assigned roles.

## 4. Configure the Go API

Set:

```env
ZITADEL_DOMAIN=your-instance.zitadel.cloud
ZITADEL_KEY_PATH=./zitadel-key.json
```

The key file is the JWT key generated for the API application in ZITADEL.

Validate these values during application startup.

## 5. Initialize ZITADEL in Go

Initialize authentication during application startup rather than inside the router:

```go
auth, err := authmw.New(ctx, cfg.Zitadel)
if err != nil {
	return fmt.Errorf("initialize auth: %w", err)
}
```

The `authmw` package should hide the ZITADEL SDK from the rest of the application.

## 6. Protect authenticated endpoints

Use `RequireAuth()` for endpoints available to any authenticated user:

```go
v1 := r.Group("/api/v1")

protected := v1.Group("")
protected.Use(authmw.Gin(auth.RequireAuth()))

protected.GET("/instances", h.ListInstances)
protected.GET("/instances/:id", h.GetInstance)
protected.POST("/instances", h.CreateInstance)
protected.PATCH("/instances/:id", h.UpdateInstance)
```

`RequireAuth()` verifies the user's ZITADEL access token.

## 7. Protect admin endpoints

Add the `admin` role only where administrator privileges are required:

```go
admin := protected.Group("")
admin.Use(authmw.Gin(auth.RequireRole("admin")))

admin.DELETE("/instances/:id", h.DeleteInstance)
```

The authorization flow is therefore:

```text
Request
   ↓
RequireAuth()
   ↓
Valid ZITADEL identity?
   │
   ├── No  → 401 Unauthorized
   │
   └── Yes
         ↓
   RequireRole("admin")
         ↓
   Has admin role?
         │
         ├── No  → 403 Forbidden
         │
         └── Yes
                ↓
             Handler
```

## 8. Get the authenticated user

After `RequireAuth()` succeeds, the API can retrieve the ZITADEL user ID:

```go
userID := auth.UserID(c.Request.Context())
```

Use the ZITADEL user ID as the stable identity in your application.

For example, if instances belong to users:

```text
instances
---------
id
owner_id
name
created_at
```

where `owner_id` contains the ZITADEL user ID.

This separates two authorization concepts:

```text
ZITADEL
├── Authentication → Who is the user?
└── Roles           → Is the user an admin?

Application DB
└── Ownership       → Does this user own this resource?
```

## 9. Recommended model

Keep the authorization model simple:

```text
ZITADEL Project
│
├── Roles
│   ├── admin
│   └── user
│
└── Assignments
    ├── You            → admin
    └── Other users    → user
```

Then in the API:

```go
// Any authenticated user.
protected.Use(authmw.Gin(auth.RequireAuth()))

// Administrator only.
admin.Use(authmw.Gin(auth.RequireRole("admin")))
```

There is normally no need to explicitly check `RequireRole("user")` on every normal endpoint. `RequireAuth()` is sufficient for operations available to all authenticated users.

If the application later needs teams, organizations, custom permissions, or resource-specific roles, more advanced authorization can be introduced without changing the basic authentication model.

# Safety and Idempotency in REST APIs – Interview Notes

## Why do these concepts exist?

REST APIs communicate over HTTP, which operates across an unreliable network. Requests can be delayed, lost, duplicated, retried, cached, or even sent automatically by browsers and intermediaries. To make APIs predictable and reliable in such an environment, HTTP defines the concepts of **Safety** and **Idempotency**.

These are not just theoretical terms—they solve real-world problems that occur every day in distributed systems.

---

# 1. Safety

## Definition

A request is **safe** if it **does not modify the state of the server**.

It is intended only for retrieving information.

The client can send the request any number of times without worrying about changing data.

### Think of it as

> "Read only."

---

## Why is Safety Needed?

Many systems send HTTP requests automatically without direct user interaction.

Examples include:

* Browsers prefetching links
* Search engine crawlers indexing websites
* Monitoring tools checking application health
* Caching proxies validating cached responses

If reading a resource changed server data, these automated requests could cause unintended consequences.

---

## Example

### Correct Design

```http
GET /users/101
```

Returns user information.

Nothing changes in the database.

---

### Incorrect Design

```http
GET /deleteUser/101
```

Imagine Google or another search engine crawls every URL on your website.

It could accidentally delete every user account simply by visiting those links.

This is exactly why **GET must be safe.**

---

## Benefits of Safety

* Browsers can freely refresh pages.
* Search engines can crawl websites safely.
* Caches can store responses.
* Monitoring tools can poll endpoints.
* Developers know that read operations never modify data.

---

# 2. Idempotency

## Definition

A request is **idempotent** if sending it multiple times has the **same final effect as sending it once**.

The first request may change the resource.

Subsequent identical requests should not change it further.

### Think of it as

> "Repeat without additional side effects."

---

## Why is Idempotency Needed?

Networks are unreliable.

Sometimes:

* the request reaches the server,
* but the response never reaches the client.

The client now has no idea whether the operation succeeded.

Should it retry?

Without idempotency, retrying could perform the operation twice.

Idempotency allows clients to safely retry requests after failures.

---

## Example

Client sends

```http
PUT /users/101

{
   "name":"Alice"
}
```

The server updates the user's name.

The response is lost because of a network timeout.

The client retries the exact same request.

Result:

The user's name is still

```
Alice
```

The resource reaches the same final state.

No duplicate updates occur.

---

## Another Example

Deleting a resource

```http
DELETE /users/101
```

First request

```
User deleted
```

Second request

```
User already deleted
```

The response may differ (for example, 404 Not Found), but the **server state remains unchanged** after the first successful deletion.

---

## Benefits of Idempotency

* Safe retries after network failures
* Prevents duplicate updates
* Prevents duplicate deletions
* Makes distributed systems more reliable
* Simplifies client-side error handling

---

# Why Both Are Needed

Safety and idempotency solve different problems.

| Concept     | Purpose                                                              |
| ----------- | -------------------------------------------------------------------- |
| Safety      | Prevent accidental modification of data during read operations.      |
| Idempotency | Allow requests to be retried without causing duplicate side effects. |

---

## Real-world Example

### ATM Balance Inquiry

Checking your balance

```
Check Balance
Check Balance
Check Balance
```

Your account never changes.

This is **safe**.

---

### Setting Your Profile Name

```
Set Name = Alice
Set Name = Alice
Set Name = Alice
```

The final result is still

```
Alice
```

This is **idempotent**.

---

### Depositing Money

```
Deposit ₹100
Deposit ₹100
Deposit ₹100
```

Balance keeps increasing.

Each request changes the result.

This is **not idempotent**.

---

# How HTTP Uses These Concepts

| HTTP Method | Safe | Idempotent | Reason                                                                                                             |
| ----------- | ---- | ---------- | ------------------------------------------------------------------------------------------------------------------ |
| GET         | ✅    | ✅          | Reads data only; repeated requests do not change server state.                                                     |
| HEAD        | ✅    | ✅          | Retrieves headers only; no state change.                                                                           |
| OPTIONS     | ✅    | ✅          | Returns communication options; no state change.                                                                    |
| PUT         | ❌    | ✅          | Replaces a resource; repeating the same replacement leaves the same final state.                                   |
| DELETE      | ❌    | ✅          | Deletes a resource; once deleted, repeating the request doesn't change the final state.                            |
| POST        | ❌    | ❌          | Usually creates a new resource or triggers an action; repeating it may create duplicates.                          |
| PATCH       | ❌    | Depends    | Some PATCH operations are idempotent (e.g., setting a value), while others (e.g., incrementing a counter) are not. |

---

# Common Interview Question

**Q: Why isn't POST idempotent?**

Suppose you create an order.

```http
POST /orders
```

Sending it once

```
Order #1001 created
```

Sending it again

```
Order #1002 created
```

Each request creates a new resource.

Therefore, POST is **not idempotent**.

---

# Common Interview Question

**Q: Why is PUT idempotent?**

```http
PUT /users/101

{
   "city":"Delhi"
}
```

Run it once.

The city becomes Delhi.

Run it ten more times.

The city is still Delhi.

The final state remains identical.

---

# Key Takeaways

* **Safety** is about **whether a request changes server state**.
* **Idempotency** is about **whether repeating the same request changes the final outcome**.
* Safe operations are designed for reading data.
* Idempotent operations make retries safe in unreliable networks.
* These properties help browsers, caches, proxies, load balancers, and client applications work together reliably.

---

# Interview Closing Statement (Impressive Answer)

> "Safety and idempotency are fundamental HTTP semantics that make REST APIs predictable in a distributed environment. Safety ensures that read operations never change server state, allowing browsers, crawlers, and caches to interact with resources without unintended side effects. Idempotency ensures that if a request is retried because of a timeout or network failure, the server reaches the same final state, preventing duplicate operations. Together, these concepts make REST APIs reliable, scalable, and resilient to real-world network conditions."

This explanation shows not only *what* these concepts mean, but *why* they exist and *how* they improve the reliability of RESTful systems—an understanding interviewers often look for beyond memorized definitions.

# Cardinality

**Cardinality** means **how many unique values a label/dimension can have**.

This is especially important in **Prometheus**.

### Simple example

Suppose you have:

```text
http_requests_total{method="GET"}
```

`method` might have only:

```text
GET
POST
PUT
DELETE
```

So the cardinality of `method` is **4**.

Now imagine:

```text
http_requests_total{user_id="12345"}
```

If you have **1 million users**, `user_id` could have **1 million unique values**.

That's **high cardinality**.

### Why is high cardinality a problem?

Prometheus creates a separate time series for each unique combination of labels.

For example:

```text
http_requests_total{
    method="GET",
    status="200"
}
```

is one time series.

But:

```text
http_requests_total{
    method="GET",
    status="200",
    user_id="12345"
}
```

and

```text
http_requests_total{
    method="GET",
    status="200",
    user_id="67890"
}
```

are **different time series**.

With millions of users, you can end up with enormous numbers of time series → more **memory, storage, and CPU usage**.

### Easy rule

**Low cardinality — generally good:**

```text
method = GET, POST
status = 200, 404, 500
environment = prod, staging
region = eu, us
```

**High cardinality — be careful:**

```text
user_id
request_id
transaction_id
email
IP address
```

Think of it as:

> **Cardinality = number of unique combinations you're asking your monitoring system to keep track of.**

This concept becomes **very important when designing Prometheus metrics**.

# Logs

**Logs** are records of things that happened inside a system, usually written as text or structured data for **debugging, troubleshooting, monitoring, and auditing**.

For example:

```text
2026-08-17 09:31:42 INFO  User login successful user_id=1234
2026-08-17 09:31:45 WARN  Database response slow duration=2.4s
2026-08-17 09:31:47 ERROR Payment failed order_id=5678
```

### Logs vs. events vs. metrics

A simple way to remember them:

* **Event → What happened?**

  * `user_signed_up`
* **Metric → How much/how often?**

  * `1,240 sign-ups today`
* **Log → What happened, with details/context?**

  * `INFO User 1234 signed up using Google at 09:31`

Logs can contain much more contextual information, such as error messages, IDs, timestamps, stack traces, and details about what the application was doing.

### Example

Suppose an application crashes.

**Metric:**

> Crash rate = 0.8%

**Event:**

> `app_crashed`

**Log:**

> `ERROR NullPointerException in PaymentService.processPayment()`
> `order_id=48291`
> `timestamp=09:42:17`

The **metric tells you there's a problem**, the **event tells you that a crash occurred**, and the **log helps you figure out why**.

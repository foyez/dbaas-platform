# Events

In telemetry, an **event** is a record that **something happened** in a system or application.

For example:

* User **opened** the app
* User **clicked** the “Buy” button
* A **login succeeded**
* A **login failed**
* A file was **uploaded**
* A server **crashed**
* A payment was **completed**

An event typically contains information such as:

```text
event_name: "button_clicked"
timestamp: 2026-08-17 09:28:15
button: "Buy now"
page: "Product page"
```

### Events vs. metrics

A useful distinction:

**Event:** *What happened?*
→ `user_signed_up`

**Metric:** *How much/how often?*
→ `1,247 sign-ups today`

So telemetry might collect both:

> At 09:28, user X clicked "Buy now."
> The system recorded 1,247 purchases today.

Events are particularly useful for understanding **user behavior, application activity, and specific actions or occurrences**.

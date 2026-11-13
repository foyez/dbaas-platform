# Metrics

**Metrics** are numerical measurements that tell you **how a system is performing or what is happening over time**.

For example, in software telemetry:

* **CPU usage:** `72%`
* **Memory usage:** `4.2 GB`
* **Response time:** `180 ms`
* **Requests per second:** `850`
* **Error rate:** `2.1%`
* **Number of active users:** `12,400`
* **App crashes:** `37`

### Events vs. metrics

Think of it this way:

|             | Event                  | Metric                        |
| ----------- | ---------------------- | ----------------------------- |
| Answers     | **What happened?**     | **How much/how often?**       |
| Example     | User clicked "Buy"     | 1,200 purchases               |
| Nature      | Individual occurrence  | Numerical measurement         |
| Typical use | Investigating behavior | Monitoring trends/performance |

For example, a website might generate an event every time someone submits a form:

> `form_submitted`

A metric could then count those events:

> **Forms submitted:** 5,430 today

So, **events are individual occurrences; metrics are measurements derived from or describing system activity.**

# Telemetry Data

Telemetry data is **information that a device, application, or system automatically collects and sends to another system for monitoring, analysis, or improvement**. The word comes from *tele* ("remote") and *metry* ("measurement")—literally, measurements sent from afar.

### Examples of telemetry data

The exact data depends on the product, but common examples include:

* **Device information**

  * Operating system version
  * Hardware model
  * Battery level
  * Memory and CPU usage

* **Performance data**

  * App startup time
  * Response times
  * Network latency
  * Resource utilization

* **Error and crash reports**

  * Crash logs
  * Error messages
  * Stack traces
  * Diagnostic information

* **Usage data**

  * Which features are used
  * How often users open the app
  * Session duration
  * Button clicks or navigation paths

* **Location or sensor data** (if enabled)

  * GPS location
  * Temperature
  * Speed
  * Accelerometer readings

### Real-world examples

* **Cars:** Modern vehicles send telemetry such as engine temperature, tire pressure, fuel consumption, and diagnostic trouble codes to manufacturers or fleet managers.
* **Smartphones:** Your phone may send anonymous crash reports and performance metrics to Apple or Google.
* **Websites:** A website may collect page load times, browser type, and which pages visitors view.
* **Cloud servers:** Servers continuously report CPU usage, memory consumption, disk health, and network traffic.

### Why companies collect telemetry

Telemetry helps organizations:

* Detect crashes and bugs
* Monitor system health
* Improve performance
* Understand which features are popular
* Predict hardware failures
* Enhance security by spotting unusual activity

### Privacy considerations

Telemetry can range from **anonymous technical metrics** to **personally identifiable information (PII)**, depending on what's collected.

Good telemetry practices include:

* Collecting only what's necessary
* Being transparent about data collection
* Giving users the option to opt out where appropriate
* Encrypting data in transit and at rest
* Removing or anonymizing personal information whenever possible

### Telemetry vs. logs

They're related but not the same:

| Telemetry                             | Logs                                   |
| ------------------------------------- | -------------------------------------- |
| Structured metrics sent automatically | Records of events or messages          |
| Used for monitoring and analytics     | Used mainly for debugging and auditing |
| Often aggregated over time            | Usually examined event by event        |

In modern software, telemetry often includes **metrics, logs, traces, and events** together to provide a complete picture of how a system is performing and being used.

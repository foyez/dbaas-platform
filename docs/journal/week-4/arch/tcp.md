# TCP (Transport Control Protocol)

## 1. TCP 3-Way Handshake

```mermaid
sequenceDiagram
    autonumber

    participant C as 🖥️ Client
    participant S as 🖥️ Server:443

    C->>S: SYN<br/>seq = x
    Note right of S: "I want to establish<br/>a TCP connection."

    S->>C: SYN-ACK<br/>seq = y<br/>ack = x + 1
    Note left of C: "I received your SYN.<br/>Here's my sequence number."

    C->>S: ACK<br/>ack = y + 1
    Note right of S: "I received your SYN-ACK."

    Note over C,S: ✅ TCP connection established
```

### Think of it as:

```text
CLIENT                                      SERVER
  │                                            │
  │──────────── SYN (seq=x) ─────────────────►│
  │          "Can we talk?"                    │
  │                                            │
  │◄────── SYN-ACK (seq=y, ack=x+1) ─────────│
  │          "Yes. I heard you."               │
  │                                            │
  │──────────── ACK (ack=y+1) ───────────────►│
  │          "Great, I heard you too."         │
  │                                            │
  │           CONNECTION ESTABLISHED           │
  │                                            │
  │════════════ DATA TRANSFER ════════════════►│
```

---

# 2. What are `seq` and `ack` actually doing?

This is one of the most important TCP concepts.

Imagine the client wants to send:

```text
HELLO
```

TCP numbers bytes so it can detect missing or out-of-order data.

```mermaid
sequenceDiagram
    participant C as 🖥️ Client
    participant S as 🖥️ Server

    C->>S: SYN seq=100
    S->>C: SYN-ACK seq=500, ack=101
    C->>S: ACK ack=501

    C->>S: DATA seq=101<br/>"HELLO" (5 bytes)
    S->>C: ACK ack=106

    Note over S: "I received bytes<br/>101 through 105.<br/>Next I expect byte 106."
```

So:

```text
Client sends:
    seq = 101
    "HELLO" = 5 bytes

Server replies:
    ack = 106

Meaning:

"I successfully received bytes
101, 102, 103, 104, 105.

The next byte I want is 106."
```

That's how TCP provides **ordered and reliable delivery**.

---

# 3. What happens after the handshake?

The handshake isn't the actual application data.

It establishes the connection first:

```mermaid
flowchart LR
    A["🖥️ Client"] --> B["1️⃣ SYN"]
    B --> C["2️⃣ SYN-ACK"]
    C --> D["3️⃣ ACK"]
    D --> E["🔗 TCP Connection"]

    E --> F["📦 Application Data"]
    F --> G["📦 ACKs"]
    G --> F

    style E fill:#c8e6c9,stroke:#388e3c
    style F fill:#e3f2fd,stroke:#1976d2
```

For HTTPS, the overall flow is roughly:

```text
DNS
 │
 ▼
IP address
 │
 ▼
TCP 3-way handshake
 │
 ▼
TLS handshake
 │
 ▼
HTTP request
 │
 ▼
HTTP response
```

So when you type:

```text
https://google.com
```

you can mentally visualize:

```mermaid
sequenceDiagram
    participant B as 🖥️ Browser
    participant S as 🌐 Google

    B->>S: TCP SYN
    S->>B: TCP SYN-ACK
    B->>S: TCP ACK

    Note over B,S: 🔗 TCP established

    B->>S: TLS ClientHello
    S->>B: TLS ServerHello + certificate + ...
    B->>S: TLS Finished

    Note over B,S: 🔐 TLS established

    B->>S: HTTP GET /
    S->>B: HTTP 200 OK + HTML

    Note over B,S: 🌐 HTTPS communication
```

---

# 4. TCP ports

The connection isn't simply:

```text
Client → Server
```

It's actually identified by a **4-tuple**:

```text
Source IP
Source Port
Destination IP
Destination Port
```

For example:

```text
Client
192.168.1.20:54321
       │
       │ TCP
       ▼
Google
142.250.x.x:443
```

So the connection is:

```text
192.168.1.20 : 54321
        ↓
142.250.x.x : 443
```

Where:

```text
54321 = ephemeral client port
443   = HTTPS server port
```

A useful mental model:

```mermaid
flowchart LR
    C["🖥️ Client<br/>192.168.1.20:54321"] 
    S["🌐 Server<br/>142.250.x.x:443"]

    C <-->|"TCP connection"| S

    style C fill:#e3f2fd,stroke:#1976d2
    style S fill:#e8f5e9,stroke:#388e3c
```

The server can therefore have thousands of TCP connections to port `443` simultaneously because each connection has a different client-side tuple.

---

# 5. MSS and Window Size

There are two different concepts worth separating.

### MSS — "How big can each TCP payload be?"

For a typical Ethernet MTU:

```text
Ethernet MTU
    1500 bytes
       │
       ├── IP header      ~20 bytes
       ├── TCP header     ~20 bytes
       │
       └── TCP payload   ~1460 bytes
```

Therefore:

```text
1500 - 20 - 20 = 1460 bytes
```

So TCP might split:

```text
Large application data
        │
        ▼
┌──────────────┐
│ TCP segment  │ 1460 bytes
├──────────────┤
│ TCP segment  │ 1460 bytes
├──────────────┤
│ TCP segment  │ 1460 bytes
├──────────────┤
│ TCP segment  │ ...
└──────────────┘
```

### Window — "How much can I send before waiting?"

Imagine the receiver says:

```text
Receive Window = 10,000 bytes
```

The sender can have roughly that much unacknowledged data in flight.

```mermaid
flowchart LR
    A["Sender"] -->|"DATA 1"| B["Receiver"]
    A -->|"DATA 2"| B
    A -->|"DATA 3"| B
    A -->|"DATA 4"| B

    B -->|"ACK + Window"| A

    style A fill:#e3f2fd,stroke:#1976d2
    style B fill:#e8f5e9,stroke:#388e3c
```

This is part of **TCP flow control**.

A larger effective window generally allows more data to be in flight, which matters especially on high-latency connections.

---

# 6. TCP failures — visualize the difference

This is particularly useful in interviews.

## Connection refused

```mermaid
sequenceDiagram
    participant C as 🖥️ Client
    participant S as 🖥️ Server

    C->>S: SYN
    S-->>C: RST ❌

    Note over C,S: Connection refused
```

The important distinction:

> **Something answered you and explicitly rejected the connection.**

Typical symptom:

```text
connect() → ECONNREFUSED
```

Usually happens when nothing is listening on that port, or something actively sends a reset.

---

## Connection timeout

```mermaid
sequenceDiagram
    participant C as 🖥️ Client
    participant F as 🧱 Firewall
    participant S as 🖥️ Server

    C->>F: SYN
    F--xS: DROP ❌

    Note over C: No SYN-ACK
    C->>F: SYN (retry)
    F--xS: DROP ❌

    Note over C: ⏱️ Eventually timeout
```

The key difference:

```text
REFUSED:

Client ── SYN ──► Server
Client ◄─ RST ─── Server
              ↑
          "NO"


TIMEOUT:

Client ── SYN ──► 🧱
                  X
              silently dropped

              ...waiting...

              ⏱️ TIMEOUT
```

**Interview shortcut:**

> `RST` → somebody actively rejected you.
> No response → something is probably dropping or preventing the traffic.

---

# 7. SYN flood

Normally:

```text
SYN
 │
 ▼
Server allocates state
 │
 ▼
SYN-ACK
 │
 ▼
ACK
 │
 ▼
Connection
```

An attacker can instead do:

```mermaid
flowchart TD
    A["☠️ Attacker"] -->|SYN| S["🖥️ Server"]
    A -->|SYN| S
    A -->|SYN| S
    A -->|SYN| S
    A -->|SYN| S
    A -->|Thousands of SYNs| S

    S --> Q["📋 Half-open connection state"]
    Q --> Q2["💥 Resource exhaustion"]

    style A fill:#ffcdd2,stroke:#c62828
    style Q2 fill:#ff8a80,stroke:#b71c1c
```

The attacker doesn't complete the handshake.

The server has lots of:

```text
SYN received
    ↓
SYN-ACK sent
    ↓
Waiting for ACK...
    ↓
Waiting...
    ↓
Waiting...
```

**SYN cookies** help by avoiding allocation of full connection state until the final ACK arrives.

---

# 8. The complete mental model

Put DNS and TCP together:

```mermaid
flowchart TD
    A["🖥️ Browser<br/>https://google.com"] 
    --> B["🔍 DNS Resolution"]

    B --> C["🔄 Recursive Resolver"]
    C --> D["🌍 Root"]
    D --> E["🏛️ .com TLD"]
    E --> F["👑 Authoritative DNS"]
    F --> C
    C --> G["142.250.x.x"]

    G --> H["🤝 TCP 3-Way Handshake"]

    H --> I["SYN"]
    I --> J["SYN-ACK"]
    J --> K["ACK"]

    K --> L["🔗 TCP Connection"]

    L --> M["🔐 TLS Handshake"]
    M --> N["🌐 HTTP Request"]
    N --> O["📦 HTTP Response"]

    style B fill:#fff3e0,stroke:#ef6c00
    style H fill:#e3f2fd,stroke:#1976d2
    style L fill:#c8e6c9,stroke:#388e3c
    style M fill:#f3e5f5,stroke:#7b1fa2
    style N fill:#e0f7fa,stroke:#00838f
    style O fill:#e0f7fa,stroke:#00838f
```

### The interview mental model

```text
1. DNS
   "What IP is google.com?"

          ↓

2. TCP
   "Can we establish a reliable connection?"

   SYN
   SYN-ACK
   ACK

          ↓

3. TLS
   "Can we establish encrypted communication?"

          ↓

4. HTTP
   "GET /"

          ↓

5. Server
   "200 OK + response"
```

The **most important distinction** is:

> **DNS finds the destination. TCP establishes the reliable connection. TLS encrypts it. HTTP carries the application request/response.**

# DNS (Domain Name Server)

### 1. Normal DNS resolution

```mermaid
sequenceDiagram
    autonumber

    participant C as 🖥️ Your Machine
    participant R as 🔄 Recursive Resolver<br/>(ISP / 8.8.8.8)
    participant Root as 🌍 Root DNS<br/>(.)
    participant TLD as 🏛️ .com TLD DNS
    participant Auth as 👑 Authoritative DNS<br/>(google.com)
    participant G as 🌐 Google Server

    C->>R: DNS query: google.com?
    
    Note over R: Check cache first
    alt Cached
        R-->>C: Cached IP: 142.250.x.x
    else Not cached
        R->>Root: Who handles .com?
        Root-->>R: .com TLD nameservers

        R->>TLD: Who handles google.com?
        TLD-->>R: Authoritative NS for google.com

        R->>Auth: What is the IP for google.com?
        Auth-->>R: 142.250.x.x + TTL

        Note over R: Cache result for TTL
        R-->>C: 142.250.x.x

        C->>G: HTTPS request to 142.250.x.x
        G-->>C: HTTP response
    end
```

### 2. The key idea

Think of DNS as **asking a chain of directories for increasingly specific information**:

```mermaid
flowchart TD
    A["🖥️ Your Browser<br/>google.com"] --> B["🔄 Recursive Resolver"]

    B --> C{"Is google.com<br/>already cached?"}

    C -->|YES| D["Return cached IP"]
    C -->|NO| E["🌍 Root DNS"]

    E -->|"Who handles .com?"| F["🏛️ .com TLD DNS"]

    F -->|"Who handles google.com?"| G["👑 Authoritative DNS"]

    G -->|"Here is the IP<br/>142.250.x.x"| B

    B --> H["💾 Cache IP + TTL"]
    H --> D

    D --> I["🖥️ Your Machine"]
    I --> J["🌐 Connect to Google"]

    style A fill:#e3f2fd,stroke:#1976d2
    style B fill:#fff3e0,stroke:#ef6c00
    style E fill:#f3e5f5,stroke:#7b1fa2
    style F fill:#e8f5e9,stroke:#388e3c
    style G fill:#ffebee,stroke:#c62828
    style H fill:#fffde7,stroke:#f9a825
    style J fill:#e0f7fa,stroke:#00838f
```

### 3. What each server actually knows

This is the part that is **really important for interviews**:

```mermaid
flowchart LR
    R["🌍 Root DNS"] 
    T["🏛️ .com TLD DNS"]
    A["👑 google.com<br/>Authoritative DNS"]

    R -->|"I don't know Google's IP.<br/>But .com is handled here →"| T
    T -->|"I don't know the IP.<br/>But google.com is handled here →"| A
    A -->|"I know the actual record:<br/>google.com → 142.250.x.x"| R

    style R fill:#f3e5f5,stroke:#7b1fa2
    style T fill:#e8f5e9,stroke:#388e3c
    style A fill:#ffebee,stroke:#c62828
```

In other words:

**Root:**

> "I don't know `google.com`, but I know who manages `.com`."

**`.com` TLD:**

> "I don't know Google's IP, but I know which authoritative nameservers manage `google.com`."

**Authoritative server:**

> "I am responsible for `google.com`. Here is its DNS record."

---

## 4. Where caching fits in

The recursive resolver is usually the most important component to understand:

```mermaid
flowchart TD
    C["🖥️ Client"] --> R["🔄 Recursive Resolver"]

    R --> CACHE{"💾 Cache<br/>google.com?"}

    CACHE -->|"HIT"| IP["142.250.x.x"]
    CACHE -->|"MISS"| ROOT["🌍 Root"]
    
    ROOT --> TLD["🏛️ .com TLD"]
    TLD --> AUTH["👑 Authoritative DNS"]
    AUTH --> RESULT["142.250.x.x<br/>TTL = 300 seconds"]

    RESULT --> R
    R --> SAVE["💾 Store in cache<br/>for 300 seconds"]
    SAVE --> IP

    IP --> C

    style CACHE fill:#fff3e0,stroke:#ef6c00
    style SAVE fill:#e8f5e9,stroke:#388e3c
```

Suppose the authoritative server says:

```text
google.com → 142.250.x.x
TTL → 300 seconds
```

The recursive resolver can answer subsequent clients from its cache for those **300 seconds**.

So:

```text
                 First request
                     │
                     ▼
Client ──────► Recursive Resolver
                     │
                  Cache MISS
                     │
                     ▼
              Root → TLD → Auth
                     │
                     ▼
                142.250.x.x
                     │
                     ▼
                   Cache
                     │
                     ▼
                   Client


                 Next request
                     │
                     ▼
Client ──────► Recursive Resolver
                     │
                  Cache HIT
                     │
                     ▼
                142.250.x.x
```

That is why the recursive resolver can make DNS look extremely fast.

---

## 5. What happens when DNS fails?

Now you can place your failure modes directly onto the resolution path:

```mermaid
flowchart TD
    A["🖥️ Client asks:<br/>google.com"] --> B["🔄 Recursive Resolver"]

    B --> C{"Cache hit?"}
    C -->|Yes| SUCCESS["✅ Return IP"]
    C -->|No| D["🌍 Root DNS"]

    D --> E{"Response?"}

    E -->|No response| TIMEOUT["⏱️ TIMEOUT<br/>Check network / UDP 53 / resolver"]
    E -->|Referral to .com| F["🏛️ .com TLD"]

    F --> G{"Response?"}
    G -->|No response| SERVFAIL["❌ SERVFAIL<br/>Check upstream / authoritative DNS"]
    G -->|Referral| H["👑 Authoritative DNS"]

    H --> I{"Domain exists?"}

    I -->|No| NX["❌ NXDOMAIN<br/>Domain does not exist"]
    I -->|Yes| J["📄 DNS Record<br/>142.250.x.x"]

    J --> K["💾 Cache according to TTL"]
    K --> SUCCESS

    style SUCCESS fill:#dcedc8,stroke:#558b2f
    style TIMEOUT fill:#ffecb3,stroke:#ff8f00
    style SERVFAIL fill:#ffcdd2,stroke:#c62828
    style NX fill:#ffcdd2,stroke:#c62828
```

### Remember the three common errors

| Error        | Meaning                                        | Think                                  |
| ------------ | ---------------------------------------------- | -------------------------------------- |
| **NXDOMAIN** | Domain definitively doesn't exist              | 🚫 **"No such domain"**                |
| **SERVFAIL** | Resolver couldn't successfully complete lookup | 💥 **"Something went wrong upstream"** |
| **Timeout**  | Didn't get a response in time                  | ⏱️ **"Nobody answered"**               |

---

## 6. And in Kubernetes / microservices

The exact same concept happens internally:

```mermaid
flowchart LR
    P["📦 Payment Pod"] --> DNS["🔄 CoreDNS"]
    
    DNS --> SVC["☸️ Kubernetes DNS"]
    
    SVC --> RECORD["payment-service<br/>→ 10.96.42.17"]

    RECORD --> DNS
    DNS --> P

    P --> SERVICE["🎯 Payment Service<br/>10.96.42.17"]

    style P fill:#e3f2fd,stroke:#1976d2
    style DNS fill:#fff3e0,stroke:#ef6c00
    style SVC fill:#e8f5e9,stroke:#388e3c
    style SERVICE fill:#f3e5f5,stroke:#7b1fa2
```

If CoreDNS is broken:

```text
Payment Pod
    │
    │ lookup payment-service
    ▼
 CoreDNS
    │
    X  💥 CoreDNS unavailable
    │
    ▼
"lookup payment-service: no such host"
```

Notice something subtle:

> **The Payment Service itself can be perfectly healthy while DNS makes it unreachable by name.**

That's why DNS is such an important production troubleshooting skill.

### The mental model I'd memorize

```text
                 DNS LOOKUP

Client
  │
  ▼
Recursive Resolver
  │
  ├── Cache HIT ──────────────► IP
  │
  └── Cache MISS
        │
        ▼
      Root
   "Who handles .com?"
        │
        ▼
      TLD
   "Who handles google.com?"
        │
        ▼
  Authoritative
   "Here is the IP"
        │
        ▼
 Recursive Resolver
   "Cache it + TTL"
        │
        ▼
      Client
        │
        ▼
     Connect
```

**One sentence for an interview:**

> “The client asks a recursive resolver, which first checks its cache; on a miss, it walks the DNS hierarchy from root to TLD to authoritative nameserver, gets the record, caches it according to its TTL, and returns the IP to the client.”

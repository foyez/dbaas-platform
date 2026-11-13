## Architecture

### Traffic Flow

```mermaid
graph LR
    Client((Client)) -->| HTTPS:443 | LB[STACKIT NLB]
    LB -->| HTTP | ING[NGINX Ingress Controller ]

    subgraph Kubernetes Cluster
        ING -->|app.foyez...cloud| FESVC[Service<br/>dbaas-frontend-svc]
        ING -->|api.foyez...cloud| APISVC[Service<br/>dbaas-svc]

        FESVC -->|HTTP:3000| FEPOD[Frontend Pods<br/>dbaas-frontend-84...]
        APISVC -->|HTTP:8080| APIPOD[API Pods<br/>dbaas-deploy-43...]
    end

    Client -.->|OIDC Login| Z[ZITADEL]
    Client -->|HTTPS + Bearer Token| LB

    classDef external fill:#f8fafc,stroke:#64748b,color:#0f172a
    classDef network fill:#fef3c7,stroke:#d97706,color:#78350f
    classDef kubernetes fill:#dcfce7,stroke:#16a34a,color:#14532d
    classDef auth fill:#ede9fe,stroke:#7c3aed,color:#4c1d95

    class Client external
    class LB,ING network
    class FESVC,APISVC,FEPOD,APIPOD kubernetes
    class Z auth
```

The platform uses a **STACKIT Network Load Balancer** as the public entry point, forwarding traffic to the **NGINX Ingress Controller** inside the Kubernetes cluster.

* **Frontend:** `app.foyez.runs.onstackit.cloud` → Frontend Service → Frontend Pods
* **Backend API:** `api.foyez.runs.onstackit.cloud` → API Service → API Pods
* **Authentication:** Users authenticate directly with **ZITADEL** using OIDC.
* **Authorization:** The frontend sends the ZITADEL access token as a `Bearer` token with API requests.
* **Services:** Kubernetes `ClusterIP` Services provide internal communication between the Ingress and Pods.
* **Scalability:** Both frontend and backend run with multiple replicas for availability.

### Instance Creation Sequential Flow 

```mermaid
sequenceDiagram
    autonumber

    participant Client
    participant Router as API Router<br/>(Gin)
    participant Handler as HTTP Handler<br/>(CreateInstance)
    participant Service as Service<br/>(CreateInstance)
    participant K8sClient as K8s Client Wrapper
    participant K8sAPI as K8s API Server
    participant CR as PostgreSQL CR
    participant Operator as Meta Operator<br/>(Kubebuilder)
    participant CNPG as CNPG Operator

    Client->>Router: POST /instances<br/>Idempotency-Key: 550e8400...

    Router->>Handler: Route request

    Handler->>Handler: Bind JSON request<br/>Validate input

    Handler->>Service: CreateInstance(request)

    Service->>Service: Generate instance ID<br/>Hash Idempotency-Key<br/>Build resource name

    Service->>K8sClient: CreateInstanceCR(instance spec)

    K8sClient->>K8sAPI: Create PostgreSQL CR

    alt CR Created
        K8sAPI->>CR: Store PostgreSQL CR
        K8sAPI-->>K8sClient: Created
        K8sClient-->>Service: Success

        Service-->>Handler: Instance Created<br/>Status: Pending
        Handler-->>Router: HTTP 202 Accepted
        Router-->>Client: Instance state

    else CR Already Exists
        K8sAPI-->>K8sClient: AlreadyExists

        K8sClient->>K8sAPI: Get PostgreSQL CR
        K8sAPI-->>K8sClient: Existing resource

        K8sClient-->>Service: Existing instance state

        Service-->>Handler: Current state
        Handler-->>Router: HTTP 200
        Router-->>Client: Existing instance state
    end


    Note over K8sAPI,Operator: Async reconciliation

    Operator->>K8sAPI: Watch PostgreSQL CR

    Operator->>K8sAPI: Create CNPG Cluster CR

    CNPG->>K8sAPI: Watch CNPG Cluster CR

    CNPG->>K8sAPI: Create PostgreSQL resources

    Note over CNPG: Creates Pods, PVCs,<br/>Services, Secrets

    CNPG->>K8sAPI: Update Cluster status

    Operator->>K8sAPI: Update PostgreSQL status = Ready
```

The API follows a **declarative, asynchronous provisioning model**:

1. Client sends `POST /instances` with an **Idempotency-Key**.
2. Gin router validates and forwards the request to the service layer.
3. The service creates a **PostgreSQL Custom Resource (CR)** through the Kubernetes API.
4. API immediately returns **`202 Accepted`** with the instance in `Pending` state.
5. The Meta Operator watches the PostgreSQL CR and creates the corresponding **CNPG Cluster CR**.
6. The **CloudNativePG (CNPG) Operator** provisions the PostgreSQL resources such as Pods, PVCs, Services, and Secrets.
7. Operators update the resource status until the instance reaches **`Ready`**.
8. If the same idempotency key is submitted again, the existing instance is returned instead of creating a duplicate.

**Key concept:**

> The API handles the request synchronously, while Kubernetes operators handle PostgreSQL provisioning asynchronously through the desired state defined in the Custom Resource.

### OpenStack Arch. Diagram

```mermaid
graph TD
    %% Define Styles
    classDef control fill:#D0E1FD,stroke:#1565C0,color:#000
    classDef compute fill:#D1E7DD,stroke:#0F5132,color:#000
    classDef network fill:#F8D7DA,stroke:#842029,color:#000
    classDef storage fill:#E2D9F3,stroke:#59359A,color:#000
    classDef deps fill:#FFF3CD,stroke:#664D03,color:#000
    classDef external fill:#FFF3CD,stroke:#664D03,color:#000

    %% External Entry
    Users([Users]) -->|Login / Create VM| Horizon[Horizon Dashboard / CLI]
    Horizon -->|REST API| ControlPlane

    %% CONTROL PLANE
    subgraph ControlPlane ["Control Plane"]
        Keystone[Keystone Identity]
        Glance[Glance API]
        NovaAPI[Nova API Compute]
        NovaScheduler[Nova Scheduler]
        Placement[Placement API]

        Keystone <-->|Auth & Issue Token| NovaAPI
        NovaAPI --> Glance
        NovaAPI --> NovaScheduler
        NovaScheduler -->|Select Node| Placement
    end
    class Keystone,Glance,NovaAPI,NovaScheduler,Placement control

    %% COMPUTE PLANE
    subgraph ComputePlane ["Compute Plane"]
        NovaCompute[Nova Compute Execute VM Build]
        Libvirt[libvirt VM Management]
        QEMU[Hypervisor QEMU]
        VM[VM]

        NovaCompute -->|Configure VM| Libvirt
        Libvirt --> QEMU
        QEMU -->|Boot from disk| VM
    end
    class NovaCompute,Libvirt,QEMU,VM compute

    %% STORAGE PLANE
    subgraph StoragePlane ["Storage"]
        Cinder[Cinder]
        LocalFS[Local Filesystem /opt/stack/data/glance/images/]
    end
    class Cinder,LocalFS storage

    %% DEPENDENCIES
    subgraph DepsPlane ["Dependencies"]
        MySQL[(MySQL Store VM Req & Metadata)]
        RabbitMQ[RabbitMQ Message Queue]
    end
    class MySQL,RabbitMQ deps

    %% NETWORK PLANE
    subgraph NetworkPlane ["Network Plane"]
        OVN[OVN Controller + agent]
        BrInt[br-int Integration Bridge]
        Geneve[Geneve overlay UDP 6081 tunnel]
        BrEx[br-ex External Bridge]

        OVN --> BrInt
        BrInt --> Geneve
        Geneve --> BrEx
    end
    class OVN,BrInt,Geneve,BrEx network

    %% CROSS-PLANE CONNECTIONS
    NovaAPI -->|Configure Network| NeutronAPI[Neutron API]
    NeutronAPI -->|1. Create port 2. Assign IP...| OVN
    
    NovaCompute -->|Send Network Info| Keystone
    Glance -->|Write / Read| LocalFS
    Cinder -->|Create / Attach Volume| VM
    
    Placement --> RabbitMQ
    NovaAPI --> MySQL

    %% EXTERNAL NETWORK
    BrEx --> FloatingIP[Floating IPs 172.24.4.228]
    FloatingIP --> PublicNet[Public Network 172.24.4.0/24]
    PublicNet --> Internet([Internet])

    class FloatingIP,PublicNet,Internet external
```

---

### PKCE Auth Flow

```mermaid
sequenceDiagram
    participant U as User
    participant A as App (Client)
    participant G as Auth Server

    A->>A: Generate random code_verifier
    A->>A: code_challenge = SHA256(code_verifier)

    U->>A: Click "Sign in"
    A->>G: Authorization request + code_challenge
    Note right of A: verifier stays in the app

    U->>G: Login + approve

    G-->>A: Authorization code

    A->>G: Exchange code + code_verifier
    G->>G: Verify SHA256(verifier) == code_challenge

    alt Match
        G-->>A: Access token (+ refresh token)
    else No match
        G-->>A: ❌ Reject
    end

    A->>G: API request + Bearer access token
    G-->>A: Protected resource
```

The key idea:

```text
Before login:

App
├── code_verifier      🔒 stays here
└── SHA256(verifier)
        │
        └──→ code_challenge → Auth Server


After login:

Auth Server
    │
    └──→ authorization_code → App

App
    │
    ├── authorization_code
    └── code_verifier 🔒
             │
             ▼
        Auth Server
             │
       SHA256(verifier)
             │
             ▼
     matches challenge?
        │          │
       YES         NO
        │           │
      Token       Reject
```

---

### Username & Password Login Flow

```mermaid
sequenceDiagram
    participant U as User
    participant C as Client (Vue)
    participant Z as ZITADEL
    participant API as Go API

    Note over API: API Application registered in ZITADEL<br/>Domain = ZITADEL instance<br/>KeyPath = API application's private key

    C->>C: Generate random code_verifier
    C->>C: code_challenge = SHA256(code_verifier)

    U->>C: Click "Sign in"

    C->>Z: OIDC Authorization Request<br/>+ PKCE code_challenge
    Z-->>U: Login Page
    U->>Z: Username + Password

    Z->>Z: Authenticate User
    Z-->>C: Authorization Code

    C->>Z: Code + code_verifier
    Z->>Z: Verify PKCE<br/>SHA256(verifier) == challenge
    
    Z-->>C: ID Token + Access Token
    C->>API: HTTPS + Bearer Access Token

    Note over API,Z: Token validation via introspection

    API->>Z: Introspection Request<br/>Bearer token + API credentials
    Note right of API: API authenticates to ZITADEL<br/>using API application's private key

    Z->>Z: Validate token<br/>active? expiry? audience? scopes?

    alt Token valid
        Z-->>API: Active + claims / permissions
        API->>API: Check required authorization
        API-->>C: Protected Resource
    else Token invalid
        Z-->>API: Inactive / invalid
        API-->>C: 401 Unauthorized
    end
```

**What is happening:**

* **OIDC** → authenticates the user and provides identity information.
* **OAuth 2.0** → provides authorization to access your Go API.
* **PKCE** → protects the authorization-code exchange.
* **ID Token** → tells the Vue app who the user is.
* **Access Token** → is sent to the Go API to access protected resources.

---

### Google Login Flow

```mermaid
sequenceDiagram
    participant U as User
    participant C as Client (Vue)
    participant Z as ZITADEL
    participant G as Google
    participant API as Go API

    C->>C: Generate random code_verifier
    C->>C: code_challenge = SHA256(code_verifier)

    U->>C: Click "Sign in"
    C->>Z: OIDC Authorization Request<br/>+ code_challenge
    Z-->>C: Login page

    U->>Z: Choose "Sign in with Google"
    Z->>G: OAuth/OIDC Authorization Request
    G-->>U: Google Login
    U->>G: Enter credentials
    G-->>Z: Authentication result

    Z->>Z: Identify / authenticate user
    Z-->>C: Authorization Code

    C->>Z: Code + code_verifier
    Z->>Z: Verify PKCE<br/>SHA256(verifier) == challenge

    alt PKCE valid
        Z-->>C: ID Token + Access Token
        C->>API: Request + Bearer Access Token
        API->>API: Validate Access Token
        API-->>C: Protected Resource
    else PKCE invalid
        Z-->>C: ❌ Reject
    end
```

### The simple picture to remember

```text
                    OIDC + OAuth 2.0
                 +       PKCE
                 ↓
Vue Client ────────────> ZITADEL
                            │
                ┌───────────┴───────────┐
                │                       │
         Username/Password           Google
                │                       │
                └───────────┬───────────┘
                            ↓
                    Authorization Code
                            ↓
                    PKCE Code Exchange
                            ↓
                   ID Token + Access Token
                            │
                            ↓
                       Vue Client
                            │
                  Bearer Access Token
                            ↓
                         Go API
```

**One sentence for the interview:**

> "My Vue application uses OIDC on top of OAuth 2.0 with Authorization Code Flow and PKCE. ZITADEL handles authentication, either with username/password or through Google as an external identity provider. After authentication, ZITADEL returns an authorization code, which Vue exchanges using the PKCE verifier for tokens. The access token is then sent as a Bearer token to my Go API, where it is validated and used for authorization."

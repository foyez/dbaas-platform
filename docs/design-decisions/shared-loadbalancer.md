# 3. One Shared LoadBalancer Exposes Databases Through Allocated TCP Ports

The alternatives were a LoadBalancer per instance, or hostname-based routing.

**Why:** a cloud LoadBalancer per tenant costs money and public IPs. Hostname (SNI) routing doesn't work with generic TCP proxies. A PostgreSQL client first sends a plaintext `SSLRequest` and only starts TLS after the server replies. So the proxy (e.g. `ssl_preread`) never sees a TLS ClientHello it could route on. PostgreSQL 17's `sslnegotiation=direct` fixes this, but the platform supports versions 14–16.

**Trade-off:** the port range is finite (30100–30999). The allocator is a ConfigMap guarded by optimistic locking. The operator also has to patch objects owned by ingress-nginx, which GitOps or Helm can overwrite.

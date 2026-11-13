# Ingress

## Run these commands to check already isntalled ingress-controller

```sh
kubectl get ingressclass
kubectl get svc -A
```

## Install Nginx Ingress controller

```sh
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx

helm repo update

helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace
```

Check ingress controller and

```sh
kubectl get pods -n ingress-nginx

# ingress-nginx-controller   Running

kubectl get ingressclass

# NAME    CONTROLLER
# nginx   k8s.io/ingress-nginx
```

## After installing ingress-nginx

`ingress.yaml`

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: dbaas-api
  namespace: dbaas-api

spec:
  ingressClassName: nginx

  rules:
    - host: api.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: dbaas-svc
                port:
                  number: 8080
```

Apply and check:

```sh
kubectl apply -f ingress.yaml

kubectl get ingress -n dbaas-api

# NAME        HOSTS             ADDRESS
# dbaas-api   api.example.com   <public-ip>
```

Then configure DNS:

```
api.example.com -> <public-ip>
```

```
Client
  |
  | api.example.com
  |
DNS record
  |
  v
ingress-nginx LoadBalancer
  |
  v
Ingress resource
  |
  v
Service (dbaas-svc)
  |
  v
Pods
```

## Get the ingress controller external address

```sh
kubectl get svc -n ingress-nginx

# NAME                                 TYPE           CLUSTER-IP      EXTERNAL-IP
# ingress-nginx-controller             LoadBalancer   100.xx.xx.xx   123.45.67.89

# OR

# EXTERNAL-IP
# my-load-balancer.example.stackit.cloud
```

## Create a DNS record

Create a DNS record in your DNS provider (STACKIT DNS, Cloudflare, Route53, etc.).

### Add an A record (if you control the DNS zone)

Create:

```sh
kubectl get svc -n ingress-nginx

# NAME                                 TYPE           CLUSTER-IP      EXTERNAL-IP
# ingress-nginx-controller             LoadBalancer   100.xx.xx.xx    203.0.113.10
```

```
Type: A
Name: foyez
Value: 203.0.113.10
TTL: 300
```

After saving:

```bash
nslookup foyez.runs.onstackit.cloud
```

should return:

```
Name: foyez.runs.onstackit.cloud
Address: 203.0.113.10
```

Then:

```bash
curl http://foyez.runs.onstackit.cloud
```

## Configure your Ingress host

Your Ingress must match the DNS name.

```yaml
rules:
    - host: api.yourdomain.com
```

Apply and Check:

```sh
kubectl apply -f ingress.yaml

kubectl get ingress -n dbaas-api

# NAME        HOSTS                 ADDRESS
# dbaas-api   api.yourdomain.com    123.45.67.89
```

## Add HTTPS (recommended)

Once HTTP works, install cert-manager and let Kubernetes manage certificates automatically:

```
Client
  |
 HTTPS
  |
Ingress
  |
cert-manager
  |
Let's Encrypt certificate
```

Then your Ingress gets:

```yaml
tls:
  - hosts:
      - api.yourdomain.com
    secretName: api-tls
```

#!/usr/bin/env bash

# Prereqs (one-time, on macOS with OrbStack installed):
#   brew install mkcert kustomize
#   orb start k8s   # or: enable Kubernetes in OrbStack Settings
#
# Usage:
#   export ZITADEL_KEY_JSON=/path/to/your/zitadel-service-account-key.json
#   ./scripts/local-up.sh

set -euo pipefail
cd "$(dirname "$0")/.."
 
CTX="orbstack"
kubectl config use-context "$CTX"
 
echo "==> 1/8  cert-manager"
if ! kubectl get ns cert-manager >/dev/null 2>&1; then
  kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml
  kubectl wait --for=condition=Available --timeout=180s -n cert-manager deployment --all
fi
 
echo "==> 2/8  ingress-nginx"
if ! kubectl get ns ingress-nginx >/dev/null 2>&1; then
  kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.11.3/deploy/static/provider/cloud/deploy.yaml
  kubectl wait --for=condition=Available --timeout=180s -n ingress-nginx deployment --all
fi
 
echo "==> 3/8  local CA (mkcert) -> cert-manager ClusterIssuer"
mkcert -install
CAROOT="$(mkcert -CAROOT)"
kubectl create secret tls dev-ca \
  --cert="$CAROOT/rootCA.pem" \
  --key="$CAROOT/rootCA-key.pem" \
  -n cert-manager \
  --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -k gitops/cert-manager/dev
 
echo "==> 4/8  CNPG operator + dbaas-operator (custom CRD/controller)"
kubectl apply --server-side -k gitops/cnpg/overlays/dev
docker build -t postgres-platform-operator:dev ./operator
kubectl apply -k gitops/operator/overlays/dev
kubectl wait --for=condition=Available --timeout=180s -n dbaas-operator deployment --all
 
echo "==> 5/8  database-system namespace + sample PostgreSQL instance"
kubectl apply -k gitops/database/overlays/dev
 
echo "==> 6/8  ZITADEL credentials (the one non-scriptable secret)"
: "${ZITADEL_KEY_JSON:?Set ZITADEL_KEY_JSON to a local zitadel service-account key file first}"
kubectl create ns dbaas-api --dry-run=client -o yaml | kubectl apply -f -
kubectl create secret generic dbaas-zitadel-key \
  --from-file=zitadel-key.json="$ZITADEL_KEY_JSON" \
  -n dbaas-api \
  --dry-run=client -o yaml | kubectl apply -f -

echo "==> 7/8 dbaas-api"
docker build -f platform/Containerfile -t dbaas-platform-api:dev ./platform
kubectl apply -k gitops/api/overlays/dev
kubectl wait --for=condition=Available --timeout=180s -n dbaas-api deployment --all

echo "==> 8/8  dbaas-frontend"
: "${ZITADEL_TF_DIR:=infra/zitadel/tf}"
: "${VITE_ZITADEL_AUTHORITY:?Export VITE_ZITADEL_AUTHORITY, e.g. https://your-instance.zitadel.cloud (same value as prod)}"
: "${VITE_ZITADEL_CLIENT_ID:?Export VITE_ZITADEL_CLIENT_ID}"
: "${VITE_ZITADEL_PROJECT_ID:?Export VITE_ZITADEL_PROJECT_ID}"
# VITE_ZITADEL_CLIENT_ID="$(terraform -chdir="$ZITADEL_TF_DIR" output -raw frontend_client_id)"
# VITE_ZITADEL_PROJECT_ID="$(terraform -chdir="$ZITADEL_TF_DIR" output -raw paas_project_id)"
 
docker build -f client/Containerfile \
  --target runtime \
  --build-arg VITE_ZITADEL_AUTHORITY="$VITE_ZITADEL_AUTHORITY" \
  --build-arg VITE_ZITADEL_CLIENT_ID="$VITE_ZITADEL_CLIENT_ID" \
  --build-arg VITE_ZITADEL_REDIRECT_URI="https://dbaas.k8s.orb.local/auth/callback" \
  --build-arg VITE_ZITADEL_POST_LOGOUT_URI="https://dbaas.k8s.orb.local" \
  --build-arg VITE_ZITADEL_PROJECT_ID="$VITE_ZITADEL_PROJECT_ID" \
  --build-arg VITE_API_URL="https://api.dbaas.k8s.orb.local/api/v1" \
  -t dbaas-frontend:dev \
  -f client/Containerfile client
kubectl apply -k gitops/client/overlays/dev
kubectl wait --for=condition=Available --timeout=180s -n dbaas-frontend deployment --all

cat <<EOF

Ready:
  UI  -> https://dbaas.k8s.orb.local
  API -> https://api.dbaas.k8s.orb.local

Both are trusted automatically (mkcert root CA), no browser warnings.

EOF

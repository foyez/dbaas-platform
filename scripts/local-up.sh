#!/usr/bin/env bash

# Prereqs (one-time, on macOS with OrbStack installed):
#   brew install kustomize
#   orb start k8s   # or: enable Kubernetes in OrbStack Settings
#
# Usage:
#   export ZITADEL_KEY_JSON=/path/to/your/zitadel-service-account-key.json
#   export VITE_ZITADEL_AUTHORITY=https://your-instance.zitadel.cloud
#   ./scripts/local-up.sh

set -euo pipefail
cd "$(dirname "$0")/.."
 
CTX="orbstack"
kubectl config use-context "$CTX"
 
echo "==> 1/6  CNPG operator + dbaas-operator (custom CRD/controller)"
kubectl apply --server-side -k gitops/cnpg/overlays/dev
docker build -t postgres-platform-operator:dev ./operator
kubectl apply -k gitops/operator/overlays/dev
kubectl wait --for=condition=Available --timeout=180s -n dbaas-operator deployment --all
 
echo "==> 2/6  database-system namespace + sample PostgreSQL instance"
kubectl apply -k gitops/database/overlays/dev
 
echo "==> 3/6  ZITADEL credentials (the one non-scriptable secret)"
: "${ZITADEL_KEY_JSON:?Set ZITADEL_KEY_JSON to a local zitadel service-account key file first}"
kubectl create ns dbaas-api --dry-run=client -o yaml | kubectl apply -f -
kubectl create secret generic dbaas-zitadel-key \
  --from-file=zitadel-key.json="$ZITADEL_KEY_JSON" \
  -n dbaas-api \
  --dry-run=client -o yaml | kubectl apply -f -

echo "==> 4/6 dbaas-api"
docker build -f platform/Containerfile -t dbaas-platform-api:dev ./platform
kubectl apply -k gitops/api/overlays/dev
kubectl wait --for=condition=Available --timeout=180s -n dbaas-api deployment --all

echo "==> 5/6  dbaas-frontend"
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
  --build-arg VITE_ZITADEL_REDIRECT_URI="http://localhost:5173/auth/callback" \
  --build-arg VITE_ZITADEL_POST_LOGOUT_URI="http://localhost:5173" \
  --build-arg VITE_ZITADEL_PROJECT_ID="$VITE_ZITADEL_PROJECT_ID" \
  --build-arg VITE_API_URL="http://localhost:8080/api/v1" \
  -t dbaas-frontend:dev \
  -f client/Containerfile client
kubectl apply -k gitops/client/overlays/dev
kubectl wait --for=condition=Available --timeout=180s -n dbaas-frontend deployment --all

echo "==> 6/6  starting port-forwards"
./scripts/local-forward.sh

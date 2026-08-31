#!/usr/bin/env bash

set -euo pipefail
cd "$(dirname "$0")/.."

./scripts/local-forward-stop.sh 2>/dev/null || true

kubectl delete -k gitops/database/overlays/dev --ignore-not-found
kubectl delete -k gitops/client/overlays/dev --ignore-not-found
kubectl delete -k gitops/api/overlays/dev --ignore-not-found
kubectl delete -k gitops/operator/overlays/dev --ignore-not-found
kubectl delete ns dbaas-api dbaas-frontend dbaas-operator database-system --ignore-not-found

echo "App namespaces removed. Cluster-wide infra (cert-manager, ingress-nginx, CNPG) left in place"

#!/usr/bin/env bash

set -euo pipefail

NAMESPACE="argocd"

echo "Starting Argo CD port-forward..."

kubectl port-forward -n "$NAMESPACE" svc/argocd-server 8081:443 \
  > /tmp/argocd-forward.log 2>&1 &

echo $! > /tmp/argocd-forward.pid

sleep 2

cat <<EOF

Ready:
  Argo CD -> https://localhost:8081

Stop with:
  ./scripts/argocd-forward-stop.sh

EOF

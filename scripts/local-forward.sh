#!/usr/bin/env bash

set -euo pipefail

echo "Starting port-forwards..."

kubectl port-forward -n dbaas-api svc/dbaas-svc 8080:8080 \
  > /tmp/dbaas-api-forward.log 2>&1 &
echo $! > /tmp/dbaas-api-forward.pid

kubectl port-forward -n dbaas-frontend svc/dbaas-frontend-svc 5173:3000 \
  > /tmp/dbaas-frontend-forward.log 2>&1 &
echo $! > /tmp/dbaas-frontend-forward.pid

sleep 2

cat <<EOF
 
Ready:
  UI  -> http://localhost:5173
  API -> http://localhost:8080/api/v1
 
Stop with: ./scripts/local-forward-stop.sh

EOF

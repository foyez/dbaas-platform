#!/usr/bin/env bash

set -euo pipefail

NAMESPACE="observability"

echo "Starting observability port-forwards..."

kubectl port-forward -n "$NAMESPACE" svc/dbaas-grafana 3000:80 \
  > /tmp/dbaas-grafana-forward.log 2>&1 &
echo $! > /tmp/dbaas-grafana-forward.pid

kubectl port-forward -n "$NAMESPACE" svc/dbaas-loki 3100:3100 \
  > /tmp/dbaas-loki-forward.log 2>&1 &
echo $! > /tmp/dbaas-loki-forward.pid

kubectl port-forward -n "$NAMESPACE" svc/dbaas-prometheus-kube-prom-prometheus 9090:9090 \
  > /tmp/dbaas-prometheus-forward.log 2>&1 &
echo $! > /tmp/dbaas-prometheus-forward.pid

sleep 2

cat <<EOF

Ready:
  Grafana    -> http://localhost:3000
  Loki       -> http://localhost:3100
  Prometheus -> http://localhost:9090

Stop with: ./scripts/observability-forward-stop.sh

EOF

#!/usr/bin/env bash

set -euo pipefail

for name in dbaas-grafana dbaas-loki dbaas-prometheus; do
  pidfile="/tmp/${name}-forward.pid"

  if [ -f "$pidfile" ]; then
    pid="$(cat "$pidfile")"

    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      echo "Stopped $name port-forward."
    else
      echo "$name port-forward is not running."
    fi

    rm -f "$pidfile"
  fi
done

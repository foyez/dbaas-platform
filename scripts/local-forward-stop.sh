#!/usr/bin/env bash
set -euo pipefail

for name in dbaas-api dbaas-frontend; do
  pidfile="/tmp/${name}-forward.pid"
  if [ -f "$pidfile" ]; then
    kill "$(cat "$pidfile")" 2>/dev/null || true
    rm -f "$pidfile"
    echo "Stopped $name port-forward."
  fi
done

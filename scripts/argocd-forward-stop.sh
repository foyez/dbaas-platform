#!/usr/bin/env bash

set -euo pipefail

PIDFILE="/tmp/argocd-forward.pid"

if [ -f "$PIDFILE" ]; then
  PID="$(cat "$PIDFILE")"

  if kill -0 "$PID" 2>/dev/null; then
    kill "$PID" 2>/dev/null || true
    echo "Stopped Argo CD port-forward."
  else
    echo "Argo CD port-forward is not running."
  fi

  rm -f "$PIDFILE"
else
  echo "Argo CD port-forward is not running."
fi

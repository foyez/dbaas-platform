#!/usr/bin/env bash
# Deletes the local kind cluster and the state make dev generated.
source "$(dirname "$0")/lib.sh"

if cluster_exists; then
  step "deleting kind cluster '$CLUSTER'"
  "$KIND" delete cluster --name "$CLUSTER"
else
  info "kind cluster '$CLUSTER' does not exist"
fi

rm -rf "$STATE_DIR"

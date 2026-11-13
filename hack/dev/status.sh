#!/usr/bin/env bash
# Shows the state of the local platform.
source "$(dirname "$0")/lib.sh"

if ! cluster_exists; then
  echo "kind cluster '$CLUSTER' is not running. Start it with: make dev"
  exit 0
fi

step "pods"
k get pods -A -o wide --field-selector=metadata.namespace!=kube-system

step "ingresses"
k get ingress -A

step "PostgreSQL instances"
k get postgresql -A 2>/dev/null || info "CRD not installed yet"

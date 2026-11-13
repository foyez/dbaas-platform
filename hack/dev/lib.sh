# Shared settings for hack/dev scripts. Sourced, not executed.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DEV_DIR="$ROOT/hack/dev"
STATE_DIR="$DEV_DIR/.state"

CLUSTER="dbaas-dev"
CTX="kind-$CLUSTER"

# Everything is served from *.dbaas.localhost on port 80. Browsers resolve
# *.localhost to 127.0.0.1 and treat it as a secure context (needed for PKCE);
# inside the cluster CoreDNS rewrites the same names to ingress-nginx.
DOMAIN="dbaas.localhost"
APP_HOST="app.$DOMAIN"
API_HOST="api.$DOMAIN"
AUTH_HOST="auth.$DOMAIN"

# The operator hands out database ports from 30100 up; kind-config.yaml maps
# the first 20 to 127.0.0.1, enough for local use.
DB_PORT_FIRST=30100
DB_PORT_LAST=30119

ZITADEL_CHART_VERSION="10.5.0"

: "${KIND:?run via make}" "${KUBECTL:?run via make}" "${HELM:?run via make}" "${TERRAFORM:?run via make}"

# Always pin the context so a stray current-context can't hit a real cluster.
k() { "$KUBECTL" --context "$CTX" "$@"; }
h() { "$HELM" --kube-context "$CTX" "$@"; }

step() { printf '\n\033[1;36m==> %s\033[0m\n' "$*"; }
info() { printf '    %s\n' "$*"; }

cluster_exists() { "$KIND" get clusters 2>/dev/null | grep -qx "$CLUSTER"; }

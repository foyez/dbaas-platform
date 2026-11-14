#!/usr/bin/env bash
# Brings up the full platform on a local kind cluster. Safe to re-run.
source "$(dirname "$0")/lib.sh"

create_cluster() {
  step "kind cluster '$CLUSTER'"
  if cluster_exists; then
    info "already exists"
  else
    "$KIND" create cluster --config "$DEV_DIR/kind-config.yaml" --wait 120s
  fi
}

install_ingress() {
  step "ingress-nginx (HTTP on :80, database TCP on :$DB_PORT_FIRST-$DB_PORT_LAST)"
  k apply -k "$DEV_DIR/ingress-nginx"
  k rollout status -n ingress-nginx deployment/ingress-nginx-controller --timeout=180s
  k wait -n ingress-nginx --for=condition=Available deployment/ingress-nginx-controller --timeout=180s
  # The admission webhook can refuse connections for a few seconds after the
  # deployment reports Available; wait until it actually serves.
  info "waiting for the admission webhook"
  until k apply --dry-run=server -f "$DEV_DIR/manifests/ingress-probe.yaml" >/dev/null 2>&1; do sleep 2; done
}

configure_dns() {
  step "CoreDNS: *.$DOMAIN -> ingress-nginx"
  local corefile rule
  rule="rewrite name regex (.+)\\.${DOMAIN//./\\.}\\.\$ ingress-nginx-controller.ingress-nginx.svc.cluster.local. answer auto"
  corefile="$(k get configmap coredns -n kube-system -o jsonpath='{.data.Corefile}')"
  if grep -qF "$DOMAIN" <<<"$corefile"; then
    info "already configured"
    return
  fi
  corefile="$(awk -v rule="$rule" '/^ *kubernetes / { print "    " rule } { print }' <<<"$corefile")"
  k create configmap coredns -n kube-system --from-literal=Corefile="$corefile" --dry-run=client -o yaml | k apply -f -
  k rollout restart -n kube-system deployment/coredns
  k rollout status -n kube-system deployment/coredns --timeout=120s
}

install_cnpg() {
  step "CloudNativePG operator"
  k apply --server-side --force-conflicts -k "$ROOT/gitops/cnpg/overlays/dev" >/dev/null
  k wait -n cnpg-system --for=condition=Available deployment --all --timeout=180s
}

install_zitadel() {
  step "ZITADEL ($ZITADEL_CHART_VERSION) at http://$AUTH_HOST"
  k apply -f "$DEV_DIR/manifests/zitadel-db.yaml"
  k wait -n zitadel --for=condition=Ready cluster.postgresql.cnpg.io/zitadel-db --timeout=300s

  if ! k get secret -n zitadel zitadel-masterkey >/dev/null 2>&1; then
    k create secret generic -n zitadel zitadel-masterkey --from-literal=masterkey="$(openssl rand -hex 16)"
  fi

  h upgrade --install zitadel zitadel --repo https://charts.zitadel.com \
    --version "$ZITADEL_CHART_VERSION" -n zitadel \
    -f "$DEV_DIR/zitadel-values.yaml" --wait --timeout 10m
  k wait -n zitadel --for=create secret/terraform --timeout=120s
}

configure_zitadel() {
  step "ZITADEL project, roles and apps (infra/zitadel/tf)"
  local tf_key="$STATE_DIR/zitadel-terraform.json" port=18080
  k get secret -n zitadel terraform -o jsonpath='{.data.terraform\.json}' | base64 -d >"$tf_key"
  chmod 600 "$tf_key"

  # Run the binary directly (not via k) so $! is kubectl itself.
  "$KUBECTL" --context "$CTX" port-forward -n zitadel svc/zitadel "$port:8080" >/dev/null 2>&1 &
  PF_PID=$!
  until curl -fs "http://localhost:$port/debug/healthz" >/dev/null; do
    kill -0 "$PF_PID" 2>/dev/null || { echo "port-forward failed; is port $port in use?" >&2; exit 1; }
    sleep 1
  done

  local tf=("$TERRAFORM" -chdir="$DEV_DIR/terraform")
  "${tf[@]}" init -input=false >/dev/null
  "${tf[@]}" apply -input=false -auto-approve \
    -var "zitadel_port=$port" -var "jwt_profile_file=$tf_key"

  "${tf[@]}" output -raw api_key_details >"$STATE_DIR/api-key.json"
  chmod 600 "$STATE_DIR/api-key.json"
  FRONTEND_CLIENT_ID="$("${tf[@]}" output -raw frontend_client_id)"
  PROJECT_ID="$("${tf[@]}" output -raw project_id)"

  cleanup
  # Builds an image and loads it into the kind node, so no registry is needed.
build_image() {
  local tag="$1"; shift
  info "building $tag"
  docker build -q -t "$tag" "$@" >/dev/null
  "$KIND" load docker-image "$tag" --name "$CLUSTER" >/dev/null 2>&1
}

deploy_operator() {
  step "dbaas operator"
  build_image postgres-platform-operator:dev "$ROOT/operator"
  k apply -k "$ROOT/gitops/operator/overlays/dev"
  k rollout restart -n dbaas-operator deployment >/dev/null
  k rollout status -n dbaas-operator deployment --timeout=180s

  # Namespace for platform-managed instances, plus a sample PostgreSQL CR.
  k apply -k "$ROOT/gitops/database/overlays/dev"
}

deploy_api() {
  step "dbaas API at http://$API_HOST"
  build_image dbaas-platform-api:dev -f "$ROOT/platform/Containerfile" --target prod "$ROOT/platform"
  k apply -k "$ROOT/gitops/api/overlays/dev"
  k create secret generic dbaas-zitadel-key -n dbaas-api \
    --from-file=zitadel-key.json="$STATE_DIR/api-key.json" \
    --dry-run=client -o yaml | k apply -f -
  k rollout restart -n dbaas-api deployment >/dev/null
  k rollout status -n dbaas-api deployment --timeout=180s
}

deploy_frontend() {
  step "dbaas frontend at http://$APP_HOST"
  build_image dbaas-frontend:dev -f "$ROOT/client/Containerfile" --target runtime \
    --build-arg VITE_ZITADEL_AUTHORITY="http://$AUTH_HOST" \
    --build-arg VITE_ZITADEL_CLIENT_ID="$FRONTEND_CLIENT_ID" \
    --build-arg VITE_ZITADEL_PROJECT_ID="$PROJECT_ID" \
    --build-arg VITE_ZITADEL_REDIRECT_URI="http://$APP_HOST/auth/callback" \
    --build-arg VITE_ZITADEL_POST_LOGOUT_URI="http://$APP_HOST" \
    --build-arg VITE_API_URL="http://$API_HOST/api/v1" \
    "$ROOT/client"
  k apply -k "$ROOT/gitops/client/overlays/dev"
  k rollout restart -n dbaas-frontend deployment >/dev/null
  k rollout status -n dbaas-frontend deployment --timeout=180s
}

PF_PID=""
}

# Builds an image and loads it into the kind node, so no registry is needed.
build_image() {
  local tag="$1"; shift
  info "building $tag"
  docker build -q -t "$tag" "$@" >/dev/null
  "$KIND" load docker-image "$tag" --name "$CLUSTER" >/dev/null 2>&1
}

deploy_operator() {
  step "dbaas operator"
  build_image postgres-platform-operator:dev "$ROOT/operator"
  k apply -k "$ROOT/gitops/operator/overlays/dev"
  # The operator wires instances into ingress-nginx's tcp-services ConfigMap.
  k apply -k "$ROOT/gitops/ingress-nginx-tcp-access/base"
  k rollout restart -n dbaas-operator deployment >/dev/null
  k rollout status -n dbaas-operator deployment --timeout=180s

  # Namespace for platform-managed instances, plus a sample PostgreSQL CR.
  k apply -k "$ROOT/gitops/database/overlays/dev"
}

deploy_api() {
  step "dbaas API at http://$API_HOST"
  build_image dbaas-platform-api:dev -f "$ROOT/platform/Containerfile" --target prod "$ROOT/platform"
  k apply -k "$ROOT/gitops/api/overlays/dev"
  k create secret generic dbaas-zitadel-key -n dbaas-api \
    --from-file=zitadel-key.json="$STATE_DIR/api-key.json" \
    --dry-run=client -o yaml | k apply -f -
  k rollout restart -n dbaas-api deployment >/dev/null
  k rollout status -n dbaas-api deployment --timeout=180s
}

deploy_frontend() {
  step "dbaas frontend at http://$APP_HOST"
  build_image dbaas-frontend:dev -f "$ROOT/client/Containerfile" --target runtime \
    --build-arg VITE_ZITADEL_AUTHORITY="http://$AUTH_HOST" \
    --build-arg VITE_ZITADEL_CLIENT_ID="$FRONTEND_CLIENT_ID" \
    --build-arg VITE_ZITADEL_PROJECT_ID="$PROJECT_ID" \
    --build-arg VITE_ZITADEL_REDIRECT_URI="http://$APP_HOST/auth/callback" \
    --build-arg VITE_ZITADEL_POST_LOGOUT_URI="http://$APP_HOST" \
    --build-arg VITE_API_URL="http://$API_HOST/api/v1" \
    "$ROOT/client"
  k apply -k "$ROOT/gitops/client/overlays/dev"
  k rollout restart -n dbaas-frontend deployment >/dev/null
  k rollout status -n dbaas-frontend deployment --timeout=180s
}

print_summary() {
  cat <<EOF

Platform is up.

  App       http://$APP_HOST
  API       http://$API_HOST/api/v1
  API docs  http://$API_HOST/docs
  ZITADEL   http://$AUTH_HOST/ui/console

  Login     admin@dbaas.$AUTH_HOST / Password1!   (dbaas.admin role)

  Status:   make dev-status
  Teardown: make dev-down
EOF
}

PF_PID=""
cleanup() {
  [[ -n "$PF_PID" ]] || return 0
  kill "$PF_PID" 2>/dev/null || true
  wait "$PF_PID" 2>/dev/null || true
}

main() {
  trap cleanup EXIT
  mkdir -p "$STATE_DIR"
  create_cluster
  install_ingress
  configure_dns
  install_cnpg
  install_zitadel
  configure_zitadel
  deploy_operator
  deploy_api
  deploy_frontend
  print_summary
}

main "$@"

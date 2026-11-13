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
  step "ingress-nginx ($INGRESS_NGINX_VERSION)"
  k apply -f "https://raw.githubusercontent.com/kubernetes/ingress-nginx/$INGRESS_NGINX_VERSION/deploy/static/provider/kind/deploy.yaml"
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

main() {
  mkdir -p "$STATE_DIR"
  create_cluster
  install_ingress
  configure_dns
  install_cnpg
  install_zitadel
}

main "$@"

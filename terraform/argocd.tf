resource "helm_release" "argocd" {
  name = "argcd"
  repository = "https://argoproj.github.io/argo-helm"
  chart = "argo-cd"
  namespace = "argocd"
  create_namespace = true
  version = "7.7.7"
}

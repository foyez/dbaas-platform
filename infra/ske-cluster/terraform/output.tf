output "kubeconfig" {
  value     = stackit_ske_kubeconfig.ske_kubeconfig.kube_config
  sensitive = true
}

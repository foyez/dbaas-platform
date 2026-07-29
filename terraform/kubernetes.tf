data "stackit_ske_kubeconfig" "cluster" {
  project_id = var.project_id
  cluster_name = stack_ske_cluster.cluster.name
}

provider "kubernetes" {
  host = data.stackit_ske_kubeconfig.cluster.host

  client_certificate = base64decode(
    data.stackit_ske_kubeconfig.cluster.client_certificate
  )

  client_key = base64decode(
    data.stackit_ske_kubeconfig.cluster.client_key
  )

  cluster_ca_certificate = base64decode(
    data.stackit_ske_kubeconfig.cluster.cluster_ca_certificate
  )
}

# Create the SKE cluster
resource "stackit_ske_cluster" "cluster" {
  project_id = var.project_id
  name       = var.cluster_name

  node_pools = [
    {
      name               = "node-pool-1"
      machine_type       = "g1a.2d"
      minimum            = 1
      maximum            = 3
      availability_zones = ["eu01-1"]

      os_version_min = "4593.2.2"
      os_name        = "flatcar"
      volume_size    = 32
      volume_type    = "storage_premium_perf6"
    }
  ]
}

# You can use 2 g1.2, 2 CPU 8 GB RAM nodes

resource "stackit_ske_kubeconfig" "ske_kubeconfig" {
  project_id   = var.project_id
  cluster_name = stackit_ske_cluster.cluster.name

  refresh        = true
  expiration     = 2592000  # 30 days
  refresh_before = 86400    # 1 day before expiration
}

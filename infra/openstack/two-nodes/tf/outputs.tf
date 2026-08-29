output "control_plane_floating_ip" {
  value       = openstack_networking_floatingip_v2.fip.address
  description = "The floating IP of the control plane"
}

# output "worker_private_ips" {
#   value = openstack_compute_instance_v2.workers[*].access_ip_v4
# }

output "control_plane_private_ip" {
  value = [
    for name, node in openstack_compute_instance_v2.nodes :
    node.access_ip_v4
    if var.nodes[name].role == "control-plane"
  ]
}

output "worker_private_ips" {
  value = [
    for name, node in openstack_compute_instance_v2.nodes :
    node.access_ip_v4
    if var.nodes[name].role == "worker"
  ]
}

output "private_network_id" {
  value       = openstack_networking_network_v2.private_net.id
  description = "The ID of the private network"
}

output "instance_floating_ip" {
  value = openstack_networking_floatingip_v2.fip.address
}

output "instance_private_ip" {
  value = openstack_compute_instance_v2.ubuntu_vm.access_ip_v4
}

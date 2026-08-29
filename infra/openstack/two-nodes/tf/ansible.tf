# ------------------------------------------------------------
# Generate Ansible Inventory
# Creates the Ansible inventory file after infrastructure
# deployment by inserting the VM's floating IP into the
# inventory template.
# ------------------------------------------------------------
resource "local_file" "ansible_inventory" {
  filename = "../ansible/inventory.ini"

  content = templatefile("${path.module}/inventory.tpl", {
    control_plane = { for k, v in var.nodes : k => v.private_ip if v.role == "control-plane" }
    workers       = { for k, v in var.nodes : k => v.private_ip if v.role == "worker" }
    bastion_ip    = openstack_networking_floatingip_v2.fip.address
  })
}

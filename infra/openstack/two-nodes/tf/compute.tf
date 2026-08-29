# ------------------------------------------------------------
# Create SSH Key Pair
# Imports an SSH public key into OpenStack so instances can
# be accessed securely using SSH authentication.
# ------------------------------------------------------------
resource "openstack_compute_keypair_v2" "mykey" {
  name       = var.key_pair_name
  public_key = file(var.public_key_path)
}

# ------------------------------------------------------------
# Security group - split by actual need instead of one bucket.
# HTTP/HTTPS rules removed: nothing on bare k8s nodes listens
# on 80/443, so that was attack surface with no purpose.
# ------------------------------------------------------------
resource "openstack_networking_secgroup_v2" "k8s_sg" {
  name        = "k8s-cluster-sg"
  description = "Kubernetes cluster nodes"
}

# ------------------------------------------------------------
# Allow SSH Access
# Allows inbound SSH connections on TCP port 22.
# ------------------------------------------------------------
resource "openstack_networking_secgroup_rule_v2" "allow_ssh" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = var.admin_cidr
  security_group_id = openstack_networking_secgroup_v2.k8s_sg.id
}

# ------------------------------------------------------------
# Allow ICMP Traffic
# Allows ICMP traffic for network connectivity checks such as
# ping.
# ------------------------------------------------------------
resource "openstack_networking_secgroup_rule_v2" "allow_icmp" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "icmp"
  security_group_id = openstack_networking_secgroup_v2.k8s_sg.id
}

resource "openstack_networking_secgroup_rule_v2" "allow_apiserver" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 6443
  port_range_max    = 6443
  remote_ip_prefix  = var.private_network_cidr
  security_group_id = openstack_networking_secgroup_v2.k8s_sg.id
}

resource "openstack_networking_secgroup_rule_v2" "allow_kubelet" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 10250
  port_range_max    = 10250
  remote_ip_prefix  = var.private_network_cidr
  security_group_id = openstack_networking_secgroup_v2.k8s_sg.id
}

resource "openstack_networking_secgroup_rule_v2" "allow_flannel_vxlan" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "udp"
  port_range_min    = 8472
  port_range_max    = 8472
  remote_ip_prefix  = var.private_network_cidr
  security_group_id = openstack_networking_secgroup_v2.k8s_sg.id
}

resource "openstack_networking_secgroup_rule_v2" "allow_nodeport" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 30000
  port_range_max    = 32767
  remote_ip_prefix  = var.private_network_cidr
  security_group_id = openstack_networking_secgroup_v2.k8s_sg.id
}

# ------------------------------------------------------------
# Fixed-IP ports - so ansible_host never drifts on rebuild,
# and so the full cluster map is known BEFORE any node boots
# (which is what lets cloud-init render a correct /etc/hosts
# and hostname on first boot, with no chicken-and-egg problem).
# ------------------------------------------------------------
resource "openstack_networking_port_v2" "node_port" {
  for_each   = var.nodes
  name       = "${each.key}-port"
  network_id = openstack_networking_network_v2.private_net.id

  fixed_ip {
    subnet_id  = openstack_networking_subnet_v2.private_subnet.id
    ip_address = each.value.private_ip
  }

  security_group_ids = [openstack_networking_secgroup_v2.k8s_sg.id]

  depends_on = [openstack_networking_subnet_v2.private_subnet]
}

# resource "openstack_compute_instance_v2" "control_plane" {
#   name = "control-plane"
#
#   image_name = var.image_name
#   flavor_name = var.flavor_name
#   key_pair = openstack_compute_keypair_v2.mykey.name
#
#   security_groups = [
#     openstack_networking_secgroup_v2.k8s_sg.name
#   ]
#
#   network = [
#     uuid = openstack_networking_network_v2.private_net.id
#   ]
# }
#
#
# resource "openstack_compute_instance_v2" "workers" {
#   count = var.worker_count
#
#   name = format("k8s-worker-%02d", count.index + 1)
#
#   image_name = var.image_name
#   flavor_name = var.flavor_name
#   key_pair   = openstack_compute_keypair_v2.mykey.name
#
#   security_groups = [
#     openstack_networking_secgroup_v2.k8s_sg.name
#   ]
#
#   network {
#     uuid = openstack_networking_network_v2.private_net.id
#   }
# }

resource "openstack_compute_instance_v2" "nodes" {
  for_each = var.nodes

  name = each.key
  # flavor_id = data.openstack_compute_flavor_v2.flavors[each.key].id
  flavor_name = each.value.flavor

  image_name = var.image_name
  key_pair   = openstack_compute_keypair_v2.mykey.name

  security_groups = [
    openstack_networking_secgroup_v2.k8s_sg.name
  ]

  network {
    port = openstack_networking_port_v2.node_port[each.key].id
  }

  # user_data = templatefile("${path.module}/cloud-init.yaml.tpl", {
  #   hostname      = each.key
  #   cluster_hosts = { for k, v in var.nodes : k => v.private_ip }
  # })
}

# ------------------------------------------------------------
# Floating IP - only the control plane gets one. Workers stay
# private-only and are reached by Ansible through the control
# plane as a jump host (see inventory.tpl) - avoids burning
# floating IPs per worker as the cluster grows.
# ------------------------------------------------------------
locals {
  control_plane_key = [for k, v in var.nodes : k if v.role == "control-plane"][0]
}

# ------------------------------------------------------------
# Create Floating IP
# Allocates a public IP address from the external network pool
# (external network with available floating IP addresses)
# so the private VM can be accessed from outside OpenStack.
# ------------------------------------------------------------
resource "openstack_networking_floatingip_v2" "fip" {
  pool = var.external_network_name
}

# ------------------------------------------------------------
# Associate Floating IP with VM
# Assigns the floating IP address to the Ubuntu VM, enabling
# external access to the instance.
# ------------------------------------------------------------
resource "openstack_compute_floatingip_associate_v2" "fip_assoc" {
  floating_ip = openstack_networking_floatingip_v2.fip.address
  instance_id = openstack_compute_instance_v2.nodes[local.control_plane_key].id
}


# ------------------------------------------------------------
# Data Source: Existing External/Public Network
# Used to link our router to the internet via the provider.
# ------------------------------------------------------------
data "openstack_networking_network_v2" "external" {
  name = var.external_network_name
}

# ------------------------------------------------------------
# Resource: Create a Private Network
# Creates an isolated internal network where instances can
# communicate with each other.
# ------------------------------------------------------------
resource "openstack_networking_network_v2" "private_net" {
  name           = var.private_network_name
  admin_state_up = true
}

# ------------------------------------------------------------
# Resource: Create a Private Subnet
# Creates an IPv4 subnet within the private network and
# assigns the CIDR range defined in the variables file.
# ------------------------------------------------------------
resource "openstack_networking_subnet_v2" "private_subnet" {
  name            = "private-subnet-2"
  network_id      = openstack_networking_network_v2.private_net.id
  cidr            = var.private_network_cidr
  ip_version      = 4
  dns_nameservers = var.dns_nameservers # pushed to every node via DHCP - no per-node DNS config needed
}

# ------------------------------------------------------------
# Resource: Create a Router
# Creates a virtual router that connects the private subnet
# to the external/public network, enabling internet access.
# ------------------------------------------------------------
resource "openstack_networking_router_v2" "router2" {
  name                = "router2"
  admin_state_up      = true
  external_network_id = data.openstack_networking_network_v2.external.id
}

# ------------------------------------------------------------
# Resource: Connect Router to Private Network
# Attaches the private subnet to the router so traffic from
# the subnet can be routed to the external network.
# ------------------------------------------------------------
resource "openstack_networking_router_interface_v2" "router_iface" {
  router_id = openstack_networking_router_v2.router2.id
  subnet_id = openstack_networking_subnet_v2.private_subnet.id
}


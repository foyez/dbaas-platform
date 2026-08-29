# Networking

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
  name           = "private-net"
  admin_state_up = true
}

# ------------------------------------------------------------
# Resource: Create a Private Subnet
# Creates an IPv4 subnet within the private network and
# assigns the CIDR range defined in the variables file.
# ------------------------------------------------------------
resource "openstack_networking_subnet_v2" "private_subnet" {
  name       = "private-subnet-2"
  network_id = openstack_networking_network_v2.private_net.id
  cidr       = var.private_network_cidr
  ip_version = 4
}

# ------------------------------------------------------------
# Resource: Create a Router
# Creates a virtual router that connects the private subnet
# to the external/public network, enabling internet access.
# ------------------------------------------------------------
resource "openstack_networking_router_v2" "router1" {
  name                = "router1"
  admin_state_up      = true
  external_network_id = data.openstack_networking_network_v2.external.id
}

# ------------------------------------------------------------
# Resource: Connect Router to Private Network
# Attaches the private subnet to the router so traffic from
# the subnet can be routed to the external network.
# ------------------------------------------------------------
resource "openstack_networking_router_interface_v2" "router_iface" {
  router_id = openstack_networking_router_v2.router1.id
  subnet_id = openstack_networking_subnet_v2.private_subnet.id
}

# ------------------------------------------------------------
# Upload Ubuntu Image (rarely done via Terraform)
# Uploads an Ubuntu cloud image to OpenStack Glance so it can
# be used as the base image when creating virtual machines.
# ------------------------------------------------------------
resource "openstack_images_image_v2" "ubuntu" {
  name             = var.image_name
  local_file_path  = "/opt/stack/jammy-server-cloudimg-amd64.img"
  disk_format      = "qcow2"
  container_format = "bare"
  visibility       = "public"
}

# ------------------------------------------------------------
# Create a Flavor
# Creates a compute flavor that defines the VM hardware
# resources such as RAM, CPU count, and disk size.
# ------------------------------------------------------------
# resource "openstack_compute_flavor_v2" "small" {
#   name  = var.flavor_name
#   ram   = 2048
#   vcpus = 2
#   disk  = 20
# }

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
# Create Security Group
# Creates firewall rules for the VM to allow required inbound
# traffic such as SSH, ICMP, HTTP, and HTTPS.
# ------------------------------------------------------------
resource "openstack_networking_secgroup_v2" "web_secgroup" {
  name        = "web-secgroup"
  description = "Allow SSH, ICMP, HTTP, HTTPS"
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
  security_group_id = openstack_networking_secgroup_v2.web_secgroup.id
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
  security_group_id = openstack_networking_secgroup_v2.web_secgroup.id
}

# ------------------------------------------------------------
# Allow HTTP Access
# Allows inbound HTTP web traffic on TCP port 80.
# ------------------------------------------------------------
resource "openstack_networking_secgroup_rule_v2" "allow_http" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 80
  port_range_max    = 80
  security_group_id = openstack_networking_secgroup_v2.web_secgroup.id
}

# ------------------------------------------------------------
# Allow HTTPS Access
# Allows inbound HTTPS web traffic on TCP port 443.
# ------------------------------------------------------------
resource "openstack_networking_secgroup_rule_v2" "allow_https" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 443
  port_range_max    = 443
  security_group_id = openstack_networking_secgroup_v2.web_secgroup.id
}

# ------------------------------------------------------------
# Launch Ubuntu VM
# Creates an Ubuntu virtual machine using the uploaded image,
# flavor, SSH key pair, and security group configuration.
# Connects the VM to the private network.
# ------------------------------------------------------------
resource "openstack_compute_instance_v2" "ubuntu_vm" {
  name = var.vm_name
  image_id        = openstack_images_image_v2.ubuntu.id
  # flavor_id       = openstack_compute_flavor_v2.small.id
  # image_name      = var.image_name
  flavor_name     = var.flavor_name
  key_pair        = openstack_compute_keypair_v2.mykey.name
  security_groups = [openstack_networking_secgroup_v2.web_secgroup.name]

  network {
    uuid = openstack_networking_network_v2.private_net.id
  }

  depends_on = [openstack_networking_router_interface_v2.router_iface]
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
  instance_id = openstack_compute_instance_v2.ubuntu_vm.id
}

# ------------------------------------------------------------
# Generate Ansible Inventory
# Creates the Ansible inventory file after infrastructure
# deployment by inserting the VM's floating IP into the
# inventory template.
# ------------------------------------------------------------
resource "local_file" "ansible_inventory" {
  filename = "../ansible/inventory.ini"

  content = templatefile("${path.module}/inventory.tpl", {
    floating_ip = openstack_networking_floatingip_v2.fip.address
  })
}

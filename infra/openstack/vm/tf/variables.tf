variable "external_network_name" {
  type        = string
  description = "The name of the existing external/public network in OpenStack"
  default     = "public"
}

variable "private_network_cidr" {
  default = "192.168.10.0/24"
}


variable "image_name" {
  default = "cirros-0.6.3-x86_64-disk"
}

variable "flavor_name" {
  default = "small"
}

variable "key_pair_name" {
  default = "mykey"
}

variable "public_key_path" {
  default = "~/.ssh/id_ed25519.pub"
}

variable "vm_name" {
  default = "ubuntu-vm"
}

variable "dns_nameservers" {
  type    = list(string)
  default = ["8.8.8.8", "8.8.4.4"]
}

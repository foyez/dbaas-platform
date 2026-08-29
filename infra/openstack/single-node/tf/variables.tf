variable "external_network_name" {
  type        = string
  description = "The name of the existing external/public network in OpenStack"
  default     = "public"
}

variable "private_network_cidr" {
  default = "192.168.10.0/24"
}


variable "image_name" {
  default = "ubuntu-24"
}

variable "vm_name" {
  default = "ubuntu-vm"
}

variable "flavor_name" {
  default = "m1.nano"
}

variable "key_pair_name" {
  default = "mykey"
}

variable "public_key_path" {
  default = "~/.ssh/id_ed25519.pub"
}

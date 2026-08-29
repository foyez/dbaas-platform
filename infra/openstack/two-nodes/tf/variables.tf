# variable "worker_count" {
#   description = "Number of Kubernetes worker nodes"
#   type = number
#   default = 1
# }

variable "nodes" {
  description = "Cluster membership. Add a map entry here to add a node"
  type = map(object({
    role       = string # "control-plane", "worker"
    flavor     = string
    private_ip = string
  }))

  default = {
    control-plane = {
      role       = "control-plane"
      flavor     = "ds2G"
      private_ip = "192.168.10.10"
    }

    worker-01 = {
      role       = "worker"
      flavor     = "ds2G"
      private_ip = "192.168.10.11"
    }
  }
}

variable "external_network_name" {
  type        = string
  description = "The name of the existing external/public network in OpenStack"
  default     = "public"
}

variable "private_network_name" {
  type        = string
  description = "The name of the private network"
  default     = "private-net"
}

variable "private_network_cidr" {
  type    = string
  default = "192.168.10.0/24"
}

variable "dns_nameservers" {
  type    = list(string)
  default = ["8.8.8.8", "8.8.4.4"]
}

variable "admin_cidr" {
  description = "CIDR allowed to SSH in. Override in terraform.tfvars - do not leave as 0.0.0.0/0."
  type        = string
  default     = "0.0.0.0/0"
}


variable "image_name" {
  type    = string
  default = "ubuntu-22"
}

variable "key_pair_name" {
  default = "mykey"
}

variable "public_key_path" {
  default = "~/.ssh/id_ed25519.pub"
}

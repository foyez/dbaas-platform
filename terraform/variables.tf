variable "project_id" {
  description = "Stackit project id"
  type        = string
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "service_account_key_path" {
  description = "Stackit service accoutn key path"
  type        = string
  default     = "./stackit-sa-key.json"
}

# variable "private_key_path" {
#   description = "SSH private key path"
#   type        = string
#   default     = "/Users/KaziFoyezAhmed/.ssh/id_rsa"
# }


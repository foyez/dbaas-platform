variable "project_id" {
  description = "Stackit project id"
  type        = string
}

variable "service_account_key" {
  description = "Stackit service account JSON"
  type = string
  sensitive = true
}

# variable "service_account_key_path" {
#   description = "Stackit service accoutn key path"
#   type        = string
#   default     = "./stackit-sa-key.json"
# }

variable "cluster_name" {
  type    = string
  default = "dbaas"
}

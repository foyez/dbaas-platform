variable "project_id" {
  description = "Stackit project id"
  type        = string
}

variable "service_account_key" {
  description = "Stackit service account JSON"
  type = string
  sensitive = true
}

variable "cluster_name" {
  type    = string
  default = "dbaas"
}

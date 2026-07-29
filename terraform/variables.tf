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

variable "service_account_key" {
  description = "Stackit service account JSON"
  type = string
  sensitive = true
}

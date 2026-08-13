variable "zitadel_domain" {
  type        = string
  description = "ZITADEL instance domain"
}

variable "zitadel_org_id" {
  type        = string
  description = "ZITADEL organization ID"
}

variable "zitadel_jwt_profile_file" {
  type        = string
  sensitive   = true
  description = "Terraform service-account JWT profile"
}

variable "admin_user_id" {
  type        = string
  description = "ZITADEL user ID of the initial administrator"
}

variable "frontend_url" {
  type        = string
  description = "Public URL of the Vue frontend"
}

variable "dev_mode" {
  type    = bool
  default = false
}

variable "google_client_id" {
  type      = string
  sensitive = true
}

variable "google_client_secret" {
  type      = string
  sensitive = true
}

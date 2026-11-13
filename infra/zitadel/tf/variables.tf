variable "zitadel_domain" {
  type        = string
  description = "ZITADEL instance domain"
}

variable "zitadel_insecure" {
  type        = bool
  default     = false
  description = "Connect to ZITADEL over plain HTTP (local dev only)"
}

variable "zitadel_port" {
  type        = string
  default     = "443"
  description = "ZITADEL API port"
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
  type        = string
  sensitive   = true
  default     = ""
  description = "Google OAuth client ID; leave empty to disable Google login"
}

variable "google_client_secret" {
  type      = string
  sensitive = true
  default   = ""
}

variable "create_api_key" {
  type        = bool
  default     = false
  description = "Create the API's introspection key in Terraform (stores it in state)"
}

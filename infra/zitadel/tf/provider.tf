terraform {
  required_version = ">= 1.6.0"

  required_providers {
    zitadel = {
      source  = "zitadel/zitadel"
      version = "~> 3.3"
    }
  }
}

provider "zitadel" {
  domain   = var.zitadel_domain
  insecure = var.zitadel_insecure
  port     = var.zitadel_port

  jwt_profile_file = var.zitadel_jwt_profile_file
}

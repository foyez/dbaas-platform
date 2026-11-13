# Applies infra/zitadel/tf to the local ZITADEL started by `make dev`.
# hack/dev/up.sh runs this through a port-forward, because the provider speaks
# gRPC and ingress-nginx can't proxy gRPC over plain HTTP.

terraform {
  required_version = ">= 1.6.0"

  backend "local" {
    path = "../.state/terraform.tfstate"
  }

  required_providers {
    zitadel = {
      source  = "zitadel/zitadel"
      version = "~> 3.3"
    }
  }
}

variable "zitadel_port" {
  type = string
}

variable "jwt_profile_file" {
  type = string
}

locals {
  domain       = "auth.dbaas.localhost"
  frontend_url = "http://app.dbaas.localhost"
}

provider "zitadel" {
  domain           = local.domain
  insecure         = true
  port             = var.zitadel_port
  jwt_profile_file = var.jwt_profile_file
}

# Org and admin user are created by ZITADEL's FirstInstance setup
# (hack/dev/zitadel-values.yaml).
data "zitadel_orgs" "dbaas" {
  name        = "dbaas"
  name_method = "TEXT_QUERY_METHOD_EQUALS"
}

data "zitadel_human_users" "admin" {
  org_id       = data.zitadel_orgs.dbaas.ids[0]
  email        = "admin@dbaas.localhost"
  email_method = "TEXT_QUERY_METHOD_EQUALS"
}

module "zitadel" {
  source = "../../../infra/zitadel/tf"

  zitadel_domain           = local.domain
  zitadel_insecure         = true
  zitadel_port             = var.zitadel_port
  zitadel_jwt_profile_file = var.jwt_profile_file
  zitadel_org_id           = data.zitadel_orgs.dbaas.ids[0]
  admin_user_id            = data.zitadel_human_users.admin.user_ids[0]
  frontend_url             = local.frontend_url
  dev_mode                 = true
  create_api_key           = true
}

output "project_id" {
  value = module.zitadel.paas_project_id
}

output "frontend_client_id" {
  value     = module.zitadel.frontend_client_id
  sensitive = true
}

output "api_key_details" {
  value     = module.zitadel.api_key_details
  sensitive = true
}

# Client-credentials login for the Go SDK examples (sdk/).
output "sdk_client_id" {
  value     = module.zitadel.sdk_test_client_id
  sensitive = true
}

output "sdk_client_secret" {
  value     = module.zitadel.sdk_test_client_secret
  sensitive = true
}

resource "zitadel_machine_user" "sdk_test" {
  org_id = var.zitadel_org_id

  user_name         = "sdk-test-dev"
  name              = "SDK Test Developer"
  description       = "Machine user for testing SDK client-credentials auth"
  access_token_type = "ACCESS_TOKEN_TYPE_JWT"
  with_secret       = true
}

resource "zitadel_user_grant" "sdk_test_user_role" {
  org_id     = var.zitadel_org_id
  project_id = zitadel_project.paas.id
  user_id    = zitadel_machine_user.sdk_test.id

  role_keys = [
    zitadel_project_role.user.role_key
  ]
}

output "sdk_test_client_id" {
  value     = zitadel_machine_user.sdk_test.client_id
  sensitive = true
}

output "sdk_test_client_secret" {
  value     = zitadel_machine_user.sdk_test.client_secret
  sensitive = true
}
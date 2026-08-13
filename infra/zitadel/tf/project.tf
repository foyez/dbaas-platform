resource "zitadel_project" "paas" {
  org_id = var.zitadel_org_id

  name = "DBAAS Platform"

  # true - zitadel can assert the project's roles in tokens
  project_role_assertion = true
  project_role_check     = true
}

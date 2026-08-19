resource "zitadel_project" "paas" {
  org_id = var.zitadel_org_id

  name = "DBAAS Platform"

  project_role_assertion = true # adds roles claim to the token
  project_role_check     = true
  has_project_check      = true
}

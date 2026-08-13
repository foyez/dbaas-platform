resource "zitadel_project_role" "admin" {
  org_id     = var.zitadel_org_id
  project_id = zitadel_project.paas.id

  role_key     = "dbaas.admin"
  display_name = "Administration"
  group        = "Application"
}

resource "zitadel_project_role" "user" {
  org_id     = var.zitadel_org_id
  project_id = zitadel_project.paas.id

  role_key     = "dbaas.user"
  display_name = "User"
  group        = "Application"
}

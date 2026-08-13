resource "zitadel_user_grant" "admin" {
  org_id     = var.zitadel_org_id
  project_id = zitadel_project.paas.id
  user_id    = var.admin_user_id

  role_keys = [
    zitadel_project_role.admin.role_key
  ]
}

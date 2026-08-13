resource "zitadel_action" "assign_default_user_role" {
  org_id = var.zitadel_org_id

  name            = "assign-paas-user-role"
  timeout         = "10s"
  allowed_to_fail = false

  script = <<-EOT
    function addDefaultUserRole(ctx, api) {
      api.userGrants.push({
        projectID: "${zitadel_project.paas.id}"
        roles: ["${zitadel_project_role.user.role_key}"]
      })
    }
  EOT
}

resource "zitadel_trigger_actions" "user_registration" {
  org_id = var.zitadel_org_id

  flow_type    = "FLOW_TYPE_INTERNAL_AUTHENTICATION"
  trigger_type = "TRIGGER_TYPE_POST_CREATION"

  action_ids = [
    zitadel_action.assign_default_user_role.id
  ]
}


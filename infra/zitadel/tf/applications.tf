resource "zitadel_application_oidc" "frontend" {
  org_id     = var.zitadel_org_id
  project_id = zitadel_project.paas.id

  name = "dbaas-web"

  # Browser SPA
  app_type         = "OIDC_APP_TYPE_USER_AGENT"
  auth_method_type = "OIDC_AUTH_METHOD_TYPE_NONE"

  # Authorization Code + PKCE
  response_types = [
    "OIDC_RESPONSE_TYPE_CODE"
  ]

  grant_types = [
    "OIDC_GRANT_TYPE_AUTHORIZATION_CODE",
    "OIDC_GRANT_TYPE_REFRESH_TOKEN"
  ]

  redirect_uris = [
    "${var.frontend_url}/auth/callback"
  ]

  post_logout_redirect_uris = [
    var.frontend_url
  ]

  access_token_type = "OIDC_TOKEN_TYPE_JWT"

  # Put project roles into the access token.
  access_token_role_assertion = true

  # If vue app needs roles from the ID token
  id_token_role_assertion = true

  dev_mode = var.dev_mode
}

resource "zitadel_application_api" "backend" {
  org_id     = var.zitadel_org_id
  project_id = zitadel_project.paas.id

  name = "dbaas-api"

  auth_method_type = "API_AUTH_METHOD_TYPE_PRIVATE_KEY_JWT"
}

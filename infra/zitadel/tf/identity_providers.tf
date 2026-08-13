resource "zitadel_org_idp_google" "google" {
  org_id = var.zitadel_org_id

  name          = "Google"
  client_id     = var.google_client_id
  client_secret = var.google_client_secret

  scopes = [
    "openid",
    "profile",
    "email",
  ]

  is_creation_allowed = true
  is_linking_allowed  = true
  is_auto_creation    = true
  is_auto_update      = true
}

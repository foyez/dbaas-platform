# Google login is optional, e.g. the local dev instance runs without it.
resource "zitadel_org_idp_google" "google" {
  count = nonsensitive(var.google_client_id != "") ? 1 : 0

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

moved {
  from = zitadel_org_idp_google.google
  to   = zitadel_org_idp_google.google[0]
}

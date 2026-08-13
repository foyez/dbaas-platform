resource "zitadel_login_policy" "paas" {
  org_id = var.zitadel_org_id

  user_login         = true
  allow_register     = true
  allow_external_idp = true

  idps = [
    zitadel_org_idp_google.google.id
  ]

  force_mfa            = false
  force_mfa_local_only = false

  passwordless_type = "PASSWORDLESS_TYPE_ALLOWED"

  hide_password_reset = false

  password_check_lifetime       = "240h0m0s"
  external_login_check_lifetime = "240h0m0s"
  multi_factor_check_lifetime   = "24h0m0s"
  mfa_init_skip_lifetime        = "720h0m0s"
  second_factor_check_lifetime  = "24h0m0s"

  ignore_unknown_usernames = true

  default_redirect_uri = var.frontend_url
}

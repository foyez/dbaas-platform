output "paas_project_id" {
  value = zitadel_project.paas.id
}

output "frontend_client_id" {
  value     = zitadel_application_oidc.frontend.client_id
  sensitive = true
}

output "api_client_id" {
  value     = zitadel_application_api.backend.client_id
  sensitive = true
}

output "api_audience_scope" {
  value = "urn:zitadel:iam:org:project:id:${zitadel_project.paas.id}:aud"
}

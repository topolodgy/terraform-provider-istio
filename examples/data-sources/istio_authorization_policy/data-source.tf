data "istio_authorization_policy" "existing" {
  name      = "allow-frontend"
  namespace = "default"
}

output "authorization_policy_manifest" {
  value = data.istio_authorization_policy.existing.manifest
}

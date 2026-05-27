data "istio_request_authentication" "existing" {
  name      = "jwt-auth"
  namespace = "default"
}

output "request_authentication_manifest" {
  value = data.istio_request_authentication.existing.manifest
}

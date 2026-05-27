data "istio_peer_authentication" "existing" {
  name      = "default"
  namespace = "istio-system"
}

output "peer_authentication_manifest" {
  value = data.istio_peer_authentication.existing.manifest
}

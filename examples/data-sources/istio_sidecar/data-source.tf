data "istio_sidecar" "existing" {
  name      = "default"
  namespace = "my-namespace"
}

output "sidecar_manifest" {
  value = data.istio_sidecar.existing.manifest
}

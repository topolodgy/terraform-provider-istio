data "istio_virtual_service" "existing" {
  name      = "my-app"
  namespace = "default"
}

output "virtual_service_manifest" {
  value = data.istio_virtual_service.existing.manifest
}

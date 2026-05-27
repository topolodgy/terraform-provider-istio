data "istio_service_entry" "existing" {
  name      = "external-api"
  namespace = "default"
}

output "service_entry_manifest" {
  value = data.istio_service_entry.existing.manifest
}

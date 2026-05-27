data "istio_telemetry" "existing" {
  name      = "default"
  namespace = "istio-system"
}

output "telemetry_manifest" {
  value = data.istio_telemetry.existing.manifest
}

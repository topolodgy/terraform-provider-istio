data "istio_gateway" "existing" {
  name      = "my-gateway"
  namespace = "istio-system"
}

output "gateway_manifest" {
  value = data.istio_gateway.existing.manifest
}

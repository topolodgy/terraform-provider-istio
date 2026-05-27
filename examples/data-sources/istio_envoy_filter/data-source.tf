data "istio_envoy_filter" "existing" {
  name      = "custom-response-header"
  namespace = "istio-system"
}

output "envoy_filter_manifest" {
  value = data.istio_envoy_filter.existing.manifest
}

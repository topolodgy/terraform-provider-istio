data "istio_proxy_config" "existing" {
  name      = "high-concurrency"
  namespace = "default"
}

output "proxy_config_manifest" {
  value = data.istio_proxy_config.existing.manifest
}

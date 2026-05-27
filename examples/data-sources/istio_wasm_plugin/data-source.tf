data "istio_wasm_plugin" "existing" {
  name      = "custom-auth"
  namespace = "istio-system"
}

output "wasm_plugin_manifest" {
  value = data.istio_wasm_plugin.existing.manifest
}

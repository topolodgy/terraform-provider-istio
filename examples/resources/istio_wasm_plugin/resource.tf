resource "istio_wasm_plugin" "custom_auth" {
  manifest = jsonencode({
    apiVersion = "extensions.istio.io/v1alpha1"
    kind       = "WasmPlugin"
    metadata = {
      name      = "custom-auth"
      namespace = "istio-system"
    }
    spec = {
      selector = {
        matchLabels = {
          istio = "ingressgateway"
        }
      }
      url   = "oci://ghcr.io/my-org/wasm-auth:v1.0.0"
      phase = "AUTHN"
      pluginConfig = {
        auth_endpoint = "https://auth.example.com/verify"
      }
    }
  })
}

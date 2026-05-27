resource "istio_proxy_config" "concurrency" {
  manifest = jsonencode({
    apiVersion = "networking.istio.io/v1beta1"
    kind       = "ProxyConfig"
    metadata = {
      name      = "high-concurrency"
      namespace = "default"
    }
    spec = {
      selector = {
        matchLabels = {
          app = "my-app"
        }
      }
      concurrency = 4
      environmentVariables = {
        ISTIO_META_DNS_CAPTURE = "true"
      }
    }
  })
}

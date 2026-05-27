resource "istio_sidecar" "default" {
  manifest = jsonencode({
    apiVersion = "networking.istio.io/v1"
    kind       = "Sidecar"
    metadata = {
      name      = "default"
      namespace = "my-namespace"
    }
    spec = {
      workloadSelector = {
        labels = {
          app = "my-app"
        }
      }
      egress = [
        {
          hosts = [
            "istio-system/*",
            "./my-dependency.my-namespace.svc.cluster.local"
          ]
        }
      ]
      outboundTrafficPolicy = {
        mode = "REGISTRY_ONLY"
      }
    }
  })
}

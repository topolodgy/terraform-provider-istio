resource "istio_service_entry" "external_api" {
  manifest = jsonencode({
    apiVersion = "networking.istio.io/v1"
    kind       = "ServiceEntry"
    metadata = {
      name      = "external-api"
      namespace = "default"
    }
    spec = {
      hosts    = ["api.external-service.com"]
      location = "MESH_EXTERNAL"
      ports = [
        {
          number   = 443
          name     = "https"
          protocol = "TLS"
        }
      ]
      resolution = "DNS"
    }
  })
}

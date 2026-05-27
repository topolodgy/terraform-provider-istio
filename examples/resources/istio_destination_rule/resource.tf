resource "istio_destination_rule" "app" {
  manifest = jsonencode({
    apiVersion = "networking.istio.io/v1"
    kind       = "DestinationRule"
    metadata = {
      name      = "my-app"
      namespace = "default"
    }
    spec = {
      host = "my-app"
      trafficPolicy = {
        connectionPool = {
          tcp = {
            maxConnections = 100
          }
          http = {
            h2UpgradePolicy       = "DEFAULT"
            http1MaxPendingRequests = 100
            http2MaxRequests       = 1000
          }
        }
        outlierDetection = {
          consecutive5xxErrors = 5
          interval             = "30s"
          baseEjectionTime     = "30s"
        }
      }
      subsets = [
        {
          name   = "v1"
          labels = { version = "v1" }
        },
        {
          name   = "v2"
          labels = { version = "v2" }
        }
      ]
    }
  })
}

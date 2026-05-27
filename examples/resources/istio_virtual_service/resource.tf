resource "istio_virtual_service" "app" {
  manifest = jsonencode({
    apiVersion = "networking.istio.io/v1"
    kind       = "VirtualService"
    metadata = {
      name      = "my-app"
      namespace = "default"
    }
    spec = {
      hosts    = ["my-app.example.com"]
      gateways = ["istio-system/my-gateway"]
      http = [
        {
          match = [
            {
              uri = { prefix = "/api" }
            }
          ]
          route = [
            {
              destination = {
                host = "my-app-api"
                port = { number = 8080 }
              }
            }
          ]
        },
        {
          route = [
            {
              destination = {
                host = "my-app-web"
                port = { number = 80 }
              }
            }
          ]
        }
      ]
    }
  })
}

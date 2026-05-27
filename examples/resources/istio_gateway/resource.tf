resource "istio_gateway" "ingress" {
  manifest = jsonencode({
    apiVersion = "networking.istio.io/v1"
    kind       = "Gateway"
    metadata = {
      name      = "my-gateway"
      namespace = "istio-system"
    }
    spec = {
      selector = {
        istio = "ingressgateway"
      }
      servers = [
        {
          port = {
            number   = 443
            name     = "https"
            protocol = "HTTPS"
          }
          hosts = ["app.example.com"]
          tls = {
            mode           = "SIMPLE"
            credentialName = "app-tls-cert"
          }
        },
        {
          port = {
            number   = 80
            name     = "http"
            protocol = "HTTP"
          }
          hosts = ["app.example.com"]
          tls = {
            httpsRedirect = true
          }
        }
      ]
    }
  })
}

resource "istio_request_authentication" "jwt" {
  manifest = jsonencode({
    apiVersion = "security.istio.io/v1"
    kind       = "RequestAuthentication"
    metadata = {
      name      = "jwt-auth"
      namespace = "default"
    }
    spec = {
      selector = {
        matchLabels = {
          app = "my-app"
        }
      }
      jwtRules = [
        {
          issuer  = "https://accounts.example.com"
          jwksUri = "https://accounts.example.com/.well-known/jwks.json"
          forwardOriginalToken = true
        }
      ]
    }
  })
}

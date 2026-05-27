resource "istio_authorization_policy" "allow_frontend" {
  manifest = jsonencode({
    apiVersion = "security.istio.io/v1"
    kind       = "AuthorizationPolicy"
    metadata = {
      name      = "allow-frontend"
      namespace = "default"
    }
    spec = {
      selector = {
        matchLabels = {
          app = "backend-api"
        }
      }
      action = "ALLOW"
      rules = [
        {
          from = [
            {
              source = {
                principals = [
                  "cluster.local/ns/default/sa/frontend"
                ]
              }
            }
          ]
          to = [
            {
              operation = {
                methods = ["GET", "POST"]
                paths   = ["/api/*"]
              }
            }
          ]
        }
      ]
    }
  })
}

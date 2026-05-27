resource "istio_envoy_filter" "custom_header" {
  manifest = jsonencode({
    apiVersion = "networking.istio.io/v1alpha3"
    kind       = "EnvoyFilter"
    metadata = {
      name      = "custom-response-header"
      namespace = "istio-system"
    }
    spec = {
      configPatches = [
        {
          applyTo = "HTTP_FILTER"
          match = {
            context = "SIDECAR_INBOUND"
            listener = {
              filterChain = {
                filter = {
                  name = "envoy.filters.network.http_connection_manager"
                }
              }
            }
          }
          patch = {
            operation = "INSERT_BEFORE"
            value = {
              name = "envoy.filters.http.lua"
              typed_config = {
                "@type"    = "type.googleapis.com/envoy.extensions.filters.http.lua.v3.Lua"
                inlineCode = "function envoy_on_response(response_handle)\n  response_handle:headers():add(\"X-Custom-Header\", \"true\")\nend"
              }
            }
          }
        }
      ]
    }
  })
}

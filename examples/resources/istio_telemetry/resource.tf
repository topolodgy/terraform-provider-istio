resource "istio_telemetry" "default" {
  manifest = jsonencode({
    apiVersion = "telemetry.istio.io/v1"
    kind       = "Telemetry"
    metadata = {
      name      = "default"
      namespace = "istio-system"
    }
    spec = {
      tracing = [
        {
          randomSamplingPercentage = 10.0
          providers = [
            { name = "otel" }
          ]
        }
      ]
      metrics = [
        {
          providers = [
            { name = "prometheus" }
          ]
        }
      ]
      accessLogging = [
        {
          providers = [
            { name = "envoy" }
          ]
        }
      ]
    }
  })
}

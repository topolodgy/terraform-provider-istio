resource "istio_workload_entry" "vm" {
  manifest = jsonencode({
    apiVersion = "networking.istio.io/v1"
    kind       = "WorkloadEntry"
    metadata = {
      name      = "my-vm"
      namespace = "default"
    }
    spec = {
      address        = "10.0.0.50"
      serviceAccount = "my-app-sa"
      labels = {
        app     = "my-app"
        version = "v1"
      }
    }
  })
}

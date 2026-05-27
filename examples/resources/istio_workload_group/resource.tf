resource "istio_workload_group" "vm_group" {
  manifest = jsonencode({
    apiVersion = "networking.istio.io/v1"
    kind       = "WorkloadGroup"
    metadata = {
      name      = "my-vm-group"
      namespace = "default"
    }
    spec = {
      metadata = {
        labels = {
          app     = "my-app"
          version = "v1"
        }
      }
      template = {
        serviceAccount = "my-app-sa"
      }
    }
  })
}

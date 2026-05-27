data "istio_workload_entry" "existing" {
  name      = "my-vm"
  namespace = "default"
}

output "workload_entry_manifest" {
  value = data.istio_workload_entry.existing.manifest
}

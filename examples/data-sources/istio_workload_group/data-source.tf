data "istio_workload_group" "existing" {
  name      = "my-vm-group"
  namespace = "default"
}

output "workload_group_manifest" {
  value = data.istio_workload_group.existing.manifest
}

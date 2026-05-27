data "istio_destination_rule" "existing" {
  name      = "my-app"
  namespace = "default"
}

output "destination_rule_manifest" {
  value = data.istio_destination_rule.existing.manifest
}

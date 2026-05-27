resource "istio_remote_cluster_secret" "remote" {
  cluster_name  = "remote-cluster"
  server        = "https://remote-cluster.example.com:443"
  ca_cert_base64 = var.remote_ca_cert
  token         = var.remote_cluster_token
  namespace     = "istio-system"
}

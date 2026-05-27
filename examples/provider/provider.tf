# Configure the Istio provider using the default kubeconfig.
provider "istio" {}

# Using a specific kubeconfig path and context.
provider "istio" {
  config_path    = "~/.kube/config"
  config_context = "my-cluster"
}

# Using explicit credentials (e.g. for CI/CD).
provider "istio" {
  host                   = "https://my-cluster.example.com:443"
  token                  = var.kube_token
  cluster_ca_certificate = var.kube_ca_cert
}

# Using in-cluster config (when running inside a Kubernetes pod).
provider "istio" {
  in_cluster = true
}

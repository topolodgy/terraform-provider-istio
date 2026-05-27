# Terraform Provider for Istio

A Terraform provider for managing [Istio](https://istio.io/) service mesh resources declaratively. Covers all Istio CRDs across networking, security, telemetry, and extensions APIs, plus a dedicated resource for multi-cluster remote secrets (replacing `istioctl create-remote-secret`).

Built with the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework) (protocol v6) and the Kubernetes dynamic client.

## Requirements

- **Terraform** >= 1.0
- **Go** >= 1.22 (see `go` version in `go.mod`; test dependencies may require a newer toolchain)
- **Kubernetes** cluster with Istio CRDs installed (this provider manages CRDs, not the Istio control plane itself)

## Compatibility

Versions exercised in CI (weekly Kind + Istio job). Other combinations may work but are not guaranteed.

| Component | Tested versions |
|-----------|-----------------|
| Terraform | >= 1.0 |
| Kubernetes (Kind) | 1.30+ (via `kind-action` default node image) |
| Istio | 1.24.2 (`minimal` profile) |
| Go (build) | >= 1.22 |

Pin Istio and this provider in production. See [UPGRADING.md](UPGRADING.md) when upgrading either.

## Production use

Suitable for production when you:

- Pin provider and Istio versions
- Manage manifests in code review (JSON plans are verbose by design)
- Run `terraform plan` after import to refresh `manifest` ([IMPORT.md](IMPORT.md))
- Use `validate_on_plan = true` in CI if the API server is available at plan time
- Avoid sharing object ownership with GitOps on the same resources without a clear boundary

Start with non-critical namespaces, then expand. See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues.

## Installation

### From the Terraform Registry

```hcl
terraform {
  required_providers {
    istio = {
      source  = "terraform-provider-istio/istio"
      version = "~> 0.1"
    }
  }
}
```

### From source

```bash
git clone https://github.com/terraform-provider-istio/terraform-provider-istio.git
cd terraform-provider-istio
go build -o terraform-provider-istio .
```

Then use [dev overrides](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "terraform-provider-istio/istio" = "/path/to/terraform-provider-istio"
  }
  direct {}
}
```

## Provider configuration

```hcl
# Default: uses KUBECONFIG or ~/.kube/config
provider "istio" {}

# Explicit kubeconfig path and context
provider "istio" {
  config_path    = "~/.kube/config"
  config_context = "my-cluster"
}

# Direct credentials (CI/CD, service accounts)
provider "istio" {
  host                   = "https://my-cluster.example.com:443"
  token                  = var.kube_token
  cluster_ca_certificate = var.kube_ca_cert
}

# In-cluster (running inside a pod)
provider "istio" {
  in_cluster = true
}

# Optional: server-side dry-run validation during plan (requires API at plan time)
provider "istio" {
  validate_on_plan = true
}
```

CRD resources support optional `timeouts` blocks (default 10 minutes per operation if omitted).

## Resources

### Istio CRD resources

All CRD resources use a single `manifest` attribute (JSON string). The `apiVersion` field in the manifest controls which API version is used, so the same resource type works across Istio versions (v1, v1beta1, v1alpha3).

| Resource | API Group | Default Version | Description |
|----------|-----------|-----------------|-------------|
| `istio_gateway` | networking.istio.io | v1 | Ingress/egress load balancer configuration |
| `istio_virtual_service` | networking.istio.io | v1 | Traffic routing rules |
| `istio_destination_rule` | networking.istio.io | v1 | Traffic policies (load balancing, connection pools, TLS) |
| `istio_service_entry` | networking.istio.io | v1 | External service registration |
| `istio_sidecar` | networking.istio.io | v1 | Per-workload proxy configuration scope |
| `istio_envoy_filter` | networking.istio.io | v1alpha3 | Low-level Envoy proxy customization |
| `istio_workload_entry` | networking.istio.io | v1 | Non-Kubernetes workload registration (VMs) |
| `istio_workload_group` | networking.istio.io | v1 | Workload instance templates |
| `istio_proxy_config` | networking.istio.io | v1beta1 | CRD-based proxy configuration |
| `istio_peer_authentication` | security.istio.io | v1 | Mutual TLS requirements |
| `istio_request_authentication` | security.istio.io | v1 | JWT validation rules |
| `istio_authorization_policy` | security.istio.io | v1 | Access control policies |
| `istio_telemetry` | telemetry.istio.io | v1 | Metrics, traces, and access logging |
| `istio_wasm_plugin` | extensions.istio.io | v1alpha1 | WebAssembly proxy extensions |

### Multi-cluster resource

| Resource | Description |
|----------|-------------|
| `istio_remote_cluster_secret` | Creates a Kubernetes Secret with the `istio/multiCluster: true` label containing a kubeconfig for a remote cluster, enabling cross-cluster service discovery |

## Data sources

Every CRD resource has a corresponding read-only data source for looking up existing resources by name and namespace.

| Data Source | Description |
|-------------|-------------|
| `istio_gateway` | Read an existing Gateway |
| `istio_virtual_service` | Read an existing VirtualService |
| `istio_destination_rule` | Read an existing DestinationRule |
| `istio_service_entry` | Read an existing ServiceEntry |
| `istio_sidecar` | Read an existing Sidecar |
| `istio_envoy_filter` | Read an existing EnvoyFilter |
| `istio_workload_entry` | Read an existing WorkloadEntry |
| `istio_workload_group` | Read an existing WorkloadGroup |
| `istio_proxy_config` | Read an existing ProxyConfig |
| `istio_peer_authentication` | Read an existing PeerAuthentication |
| `istio_request_authentication` | Read an existing RequestAuthentication |
| `istio_authorization_policy` | Read an existing AuthorizationPolicy |
| `istio_telemetry` | Read an existing Telemetry |
| `istio_wasm_plugin` | Read an existing WasmPlugin |

## Import

All resources support `terraform import`. After import, run **`terraform plan`** to populate and verify `manifest`.

```bash
terraform import istio_gateway.my_gw istio-system/my-gateway
terraform plan
```

Full guide: [IMPORT.md](IMPORT.md).

## Comparison with other approaches

| Approach | Best for |
|----------|----------|
| **This provider** | Terraform-native Istio CRUD, state, import, multi-cluster secret helper, manifest normalization |
| **Helm / `istioctl`** | Installing and upgrading Istio control plane |
| **`kubernetes_manifest`** (generic) | Any K8s manifest; no Istio-specific normalization, redaction, or remote-cluster secret resource |
| **GitOps (Argo/Flux)** | Continuous reconcile from Git; avoid managing the same objects with Terraform |

## Examples

See the [`examples/`](examples/) directory for usage examples of every resource and data source. Key examples:

- [Provider configuration](examples/provider/provider.tf)
- [Gateway with TLS](examples/resources/istio_gateway/resource.tf)
- [VirtualService routing](examples/resources/istio_virtual_service/resource.tf)
- [PeerAuthentication (STRICT mTLS)](examples/resources/istio_peer_authentication/resource.tf)
- [AuthorizationPolicy](examples/resources/istio_authorization_policy/resource.tf)
- [Remote cluster secret](examples/resources/istio_remote_cluster_secret/resource.tf)

## Documentation

| Document | Description |
|----------|-------------|
| [`docs/`](docs/) | Registry documentation (generated) |
| [CHANGELOG.md](CHANGELOG.md) | Release history |
| [IMPORT.md](IMPORT.md) | Import workflow and manifest refresh |
| [UPGRADING.md](UPGRADING.md) | Provider and Istio upgrade guidance |
| [TROUBLESHOOTING.md](TROUBLESHOOTING.md) | Common errors and fixes |
| [PUBLISHING.md](PUBLISHING.md) | Release and Terraform Registry publish steps |
| [SECURITY.md](SECURITY.md) | Security policy and vulnerability reporting |
| [SUPPORT.md](SUPPORT.md) | Versioning and support expectations |
| [GAPS_AND_ROADMAP.md](GAPS_AND_ROADMAP.md) | Implemented features and future work |

## Versioning

This provider is **0.x** — pin `~> 0.1` in `required_providers`. See [SUPPORT.md](SUPPORT.md) for semver expectations and [CHANGELOG.md](CHANGELOG.md) before upgrading.

To regenerate docs locally:

```bash
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
tfplugindocs generate --provider-name istio
```

## Design decisions

- **Manifest-based CRDs**: All Istio CRD resources use a single `manifest` JSON attribute rather than per-field schemas. This avoids schema drift as Istio evolves and supports any CRD field without provider updates.
- **API version from manifest**: The `apiVersion` in the manifest JSON determines the Kubernetes API version used for CRUD operations. This means the same `istio_gateway` resource works whether your cluster runs v1, v1beta1, or v1alpha3.
- **Server-side field stripping**: On read, server-managed fields (`resourceVersion`, `uid`, `creationTimestamp`, `generation`, `managedFields`, `status`) are stripped to prevent spurious plan diffs.
- **Path-based redaction**: Sensitive JSON keys are stored as `(redacted)` in state; semantic equality reduces drift against configuration.
- **No Istio installation**: This provider manages Istio resources, not the Istio control plane. Use the Helm provider or `istioctl` to install Istio first.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## License

[Mozilla Public License 2.0](LICENSE)

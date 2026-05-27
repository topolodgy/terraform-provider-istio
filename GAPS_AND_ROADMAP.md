# Roadmap

## Implemented

### Provider
- Kubernetes connection via `config_path`, `config_context`, `in_cluster`, or direct `host`/`token`/`cluster_ca_certificate`
- Falls back to `KUBECONFIG` env var, then `~/.kube/config`
- Optional `validate_on_plan` for server-side dry-run validation during `terraform plan`
- Configurable `timeouts` on CRD resources and `istio_remote_cluster_secret`

### Resources (15 total)
- **`istio_remote_cluster_secret`** -- multi-cluster registration (equivalent to `istioctl create-remote-secret`), with full ImportState support
- **14 CRD resources** covering all four Istio API groups:
  - **networking.istio.io**: Gateway, VirtualService, DestinationRule, ServiceEntry, Sidecar, EnvoyFilter, WorkloadEntry, WorkloadGroup, ProxyConfig
  - **security.istio.io**: PeerAuthentication, RequestAuthentication, AuthorizationPolicy
  - **telemetry.istio.io**: Telemetry
  - **extensions.istio.io**: WasmPlugin

### Data sources (14 total)
- One read-only data source per CRD resource (read by name and namespace)

### Features
- **ImportState** on all resources (`terraform import`); see [IMPORT.md](IMPORT.md)
- **API-version-aware**: the `apiVersion` in the manifest drives the Kubernetes API version (v1, v1beta1, v1alpha3)
- **Server-side field stripping**: prevents plan drift from K8s-managed metadata
- **Manifest validation**: `apiVersion`, `kind`, `metadata.name`, and `metadata.namespace` required before API calls
- **Path-based secret redaction**: sensitive JSON keys redacted to `(redacted)` in state; semantic equality suppresses drift vs configuration
- **Manifest normalization**: empty fields pruned and canonical JSON key ordering in state
- **Plan-time dry-run** (optional): `validate_on_plan` on the provider
- **Registry-ready**: MPL-2.0 license, GoReleaser config, terraform-registry-manifest.json, generated docs
- **Operator docs**: [IMPORT.md](IMPORT.md), [UPGRADING.md](UPGRADING.md), [TROUBLESHOOTING.md](TROUBLESHOOTING.md), [PUBLISHING.md](PUBLISHING.md), [SECURITY.md](SECURITY.md), [SUPPORT.md](SUPPORT.md), [CHANGELOG.md](CHANGELOG.md)

### Testing
- **Unit tests** for helpers (strip, parse, GVR, validation, redaction, normalization, semantic equality)
- **Acceptance tests**: plan-only dry-run and full lifecycle for ServiceEntry, VirtualService, PeerAuthentication, Gateway, and remote cluster secret
- **CI**: acceptance tests on pull requests, pushes to `main`, and weekly Kind + Istio schedule

## Future work

Items below are not yet implemented. Contributions welcome.

### Testing
- Lifecycle acceptance tests for additional CRD types (AuthorizationPolicy, DestinationRule, etc.)
- Published compatibility matrix expanded with more Istio minor versions

### Additional CRDs
The generic CRD pattern makes adding new resources straightforward. To add a new Istio CRD:

1. Add a `CRDResourceDef` entry to `AllCRDResources()` in `internal/provider/crd_resource.go`
2. Create an example in `examples/resources/istio_<name>/resource.tf`
3. Create a data source example in `examples/data-sources/istio_<name>/data-source.tf`
4. Run `tfplugindocs generate --provider-name istio`

Potential additions as Istio adds CRDs:
- `istio_service_mesh` (future Istio Ambient API)
- Any new CRDs introduced in upcoming Istio releases

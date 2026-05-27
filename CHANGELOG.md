# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-05-21

### Added

- Release documentation ([PUBLISHING.md](PUBLISHING.md), [SECURITY.md](SECURITY.md), [SUPPORT.md](SUPPORT.md))
- Gateway lifecycle and plan-only acceptance tests
- Acceptance tests on pull requests to `main`

### Changed

- Pinned `go.mod` to Go 1.22

## [0.1.0] - 2026-05-21

### Added

- Terraform provider for 14 Istio CRDs (networking, security, telemetry, extensions API groups) using a manifest-based resource model
- `istio_remote_cluster_secret` resource for multi-cluster kubeconfig secrets (replaces `istioctl create-remote-secret`)
- Read-only data sources for each supported CRD
- Provider configuration: kubeconfig path/context, in-cluster mode, direct API host/token/CA
- Optional `validate_on_plan` for server-side dry-run validation during `terraform plan`
- Configurable `timeouts` on CRD resources and `istio_remote_cluster_secret`
- Manifest validation (`apiVersion`, `kind`, `metadata.name`, `metadata.namespace`)
- Path-based secret redaction in state with semantic equality to reduce drift
- Manifest normalization (server field stripping, empty field pruning, canonical JSON)
- Import support for all resources
- Operator documentation: [IMPORT.md](IMPORT.md), [UPGRADING.md](UPGRADING.md), [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
- Acceptance tests (plan-only and lifecycle) with weekly Kind + Istio CI
- [SECURITY.md](SECURITY.md) and [SUPPORT.md](SUPPORT.md)

### Notes

- This provider manages Istio **custom resources** only; it does not install the Istio control plane
- Pin provider and Istio versions in production; see [UPGRADING.md](UPGRADING.md)

[0.1.1]: https://github.com/Topolodgy/terraform-provider-istio/releases/tag/v0.1.1
[0.1.0]: https://github.com/Topolodgy/terraform-provider-istio/releases/tag/v0.1.0

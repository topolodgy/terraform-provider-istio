# Support policy

## Versioning

This project follows [Semantic Versioning](https://semver.org/):

| Release | Expectations |
|---------|----------------|
| **0.x** | Initial adoption. Minor versions may add features; patch versions are bug fixes. Review [CHANGELOG.md](CHANGELOG.md) before upgrading. |
| **1.x** (future) | Stable API for provider configuration and resource schemas. Breaking changes only in major versions with migration notes. |

Pin the provider explicitly in Terraform:

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

## Compatibility

Supported combinations are documented in [README.md](README.md#compatibility). CI tests against Istio **1.24.2** on Kind; other Istio 1.2x versions may work when CRDs match.

You are responsible for validating against your Istio and Kubernetes versions in a non-production environment before rollout.

## How to get help

| Channel | Use for |
|---------|---------|
| [GitHub Issues](https://github.com/Topolodgy/terraform-provider-istio/issues) | Bugs, feature requests, questions |
| [TROUBLESHOOTING.md](TROUBLESHOOTING.md) | Common errors and diagnostics |
| [IMPORT.md](IMPORT.md) / [UPGRADING.md](UPGRADING.md) | Import and upgrade workflows |

For bugs, include provider version, Istio version, Kubernetes version, Terraform version, and redacted configuration.

## Maintenance

- **Active development**: best-effort review of issues and pull requests on business days
- **Releases**: tagged with `v*`; binaries and registry artifacts via GoReleaser (see [PUBLISHING.md](PUBLISHING.md))
- **End of support**: unsupported versions are listed in [SECURITY.md](SECURITY.md)

## Not covered

- Installing or upgrading Istio itself (use Helm or `istioctl`)
- Application debugging inside the mesh
- Commercial SLAs (unless your organization provides an internal support wrapper)

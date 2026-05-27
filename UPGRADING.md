# Upgrading

Guidance for upgrading this provider and Istio together.

## Provider versions

Follow [semantic versioning](https://semver.org/):

| Version | Expectations |
|---------|----------------|
| `0.x` | Initial adoption; pin exact version in `required_providers` |
| `1.x` | Stable CRUD for documented CRDs; breaking changes noted in CHANGELOG |

Always pin the provider in Terraform:

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

## Istio upgrades

This provider does **not** upgrade Istio. Upgrade Istio with Helm or `istioctl` first, then validate Terraform.

### Checklist

1. Confirm target Istio version CRDs are installed: `kubectl api-resources --api-group=networking.istio.io`
2. Review [Istio upgrade release notes](https://istio.io/latest/news/releases/) for API removals.
3. Update `apiVersion` in each `manifest` if Istio deprecated a version (e.g. `v1beta1` → `v1`).
4. Run `terraform plan` with `validate_on_plan = true` in CI to catch rejected manifests early.
5. Apply non-production workspace first.

### API version in manifests

The `apiVersion` field in each manifest selects the Kubernetes API version. After upgrading Istio, update manifests to match supported versions on the cluster.

```hcl
manifest = jsonencode({
  apiVersion = "networking.istio.io/v1"  # adjust for your Istio version
  kind       = "VirtualService"
  # ...
})
```

## State and drift

After upgrading the provider:

- **Normalization** may change whitespace or remove empty JSON fields — expect a one-time plan diff.
- **Redaction** stores sensitive values as `(redacted)` in state; configuration may still contain real secrets without forcing replacement when semantics match.

Run `terraform plan` after every provider upgrade before applying in production.

## Breaking changes

Breaking changes will be documented in GitHub releases and summarized here for major versions.

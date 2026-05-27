# Publishing to the Terraform Registry

This guide covers releasing a signed provider build to the [Terraform Registry](https://registry.terraform.io/).

## Prerequisites

1. **Terraform Registry account** — sign in at [registry.terraform.io](https://registry.terraform.io/) and link your GitHub organization/user.
2. **Provider namespace** — register `terraform-provider-istio` (or your org namespace) and publish the `istio` provider name.
3. **GPG signing key** — Terraform Registry requires signed releases ([HashiCorp tutorial](https://developer.hashicorp.com/terraform/tutorials/providers/plugin-framework-provider#deploy-the-provider)).
4. **GitHub repository secrets** (Settings → Secrets → Actions):
   - `GPG_PRIVATE_KEY` — ASCII-armored private key
   - `PASSPHRASE` — key passphrase (if any)

## Release steps

1. Ensure `main` is green (CI + acceptance tests).
2. Update [CHANGELOG.md](CHANGELOG.md) with the release date and notes.
3. Commit and push changelog changes.
4. Create and push a version tag:

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

5. The [Release workflow](.github/workflows/release.yml) runs GoReleaser, which:
   - Builds multi-platform binaries
   - Generates SHA256 checksums
   - Signs artifacts with GPG
   - Uploads release assets and `terraform-registry-manifest.json`
6. In the Terraform Registry UI, approve the new provider version if manual approval is required.

## Verify installation

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

```bash
terraform init
terraform providers
```

## Troubleshooting releases

| Issue | Action |
|-------|--------|
| GoReleaser GPG failure | Verify `GPG_PRIVATE_KEY` and `PASSPHRASE` secrets; fingerprint must match `.goreleaser.yml` |
| Registry rejects manifest | Confirm `terraform-registry-manifest.json` protocol `6.0` and provider address in `main.go` |
| Version already exists | Bump tag (e.g. `v0.1.1`) — registry versions are immutable |

## Dev registry (optional)

For local testing before publish, use [provider development overrides](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) in `~/.terraformrc`.

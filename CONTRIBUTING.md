# Contributing

Thank you for your interest in contributing to the Terraform Provider for Istio.

## Development setup

### Requirements

- [Go](https://go.dev/) (match the version in `go.mod`)
- [Terraform](https://www.terraform.io/) >= 1.0
- A Kubernetes cluster with Istio installed (for acceptance testing)

### Building

```bash
git clone https://github.com/terraform-provider-istio/terraform-provider-istio.git
cd terraform-provider-istio
go build -o terraform-provider-istio .
```

### Running locally

Use [dev overrides](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "terraform-provider-istio/istio" = "/path/to/terraform-provider-istio"
  }
  direct {}
}
```

### Running tests

Unit tests (no cluster required):

```bash
go test ./...
go vet ./...
```

Acceptance tests require a Kubernetes cluster with Istio CRDs installed and a valid kubeconfig:

```bash
export TF_ACC=1

# Plan-only: validates manifests via dry-run; does not create or change cluster resources
go test ./internal/provider/... -run TestAccPlanOnly -count=1 -v

# Lifecycle: creates, updates, imports, and destroys test resources (unique tf-acc-* names)
go test ./internal/provider/... -run TestAccLifecycle -count=1 -v
```

Tests are skipped when `TF_ACC` is unset or kubeconfig is missing.

Acceptance tests run on **pull requests** and **pushes** to `main` (`.github/workflows/acc-test.yml`), plus a weekly schedule. They use Kind + Istio 1.24.2. Trigger manually via **workflow_dispatch** if needed.

See also [IMPORT.md](IMPORT.md), [UPGRADING.md](UPGRADING.md), [TROUBLESHOOTING.md](TROUBLESHOOTING.md), and [PUBLISHING.md](PUBLISHING.md).

### Generating docs

After making schema changes, regenerate the docs:

```bash
go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
tfplugindocs generate --provider-name istio
```

## Adding a new Istio CRD resource

The provider uses a generic CRD factory. To add support for a new Istio CRD:

1. Add a `CRDResourceDef` to `AllCRDResources()` in [`internal/provider/crd_resource.go`](internal/provider/crd_resource.go)
2. Create `examples/resources/istio_<suffix>/resource.tf` with a realistic example
3. Create `examples/data-sources/istio_<suffix>/data-source.tf`
4. Run `tfplugindocs generate --provider-name istio` to regenerate docs
5. Update the resource table in `README.md`

The data source is automatically registered -- no additional code needed.

## Pull request guidelines

- One focused change per PR
- Include a clear description of the change and its motivation
- Run `go test ./...` and `go vet ./...` before submitting
- If adding a new resource, include an example `.tf` file and regenerate docs
- Keep backward compatibility -- avoid breaking changes to existing resource schemas

## Versioning

This project follows [semantic versioning](https://semver.org/). Releases are created via GitHub Actions using [GoReleaser](https://goreleaser.com/) when a tag matching `v*` is pushed.

## License

By contributing, you agree that your contributions will be licensed under the [Mozilla Public License 2.0](LICENSE).

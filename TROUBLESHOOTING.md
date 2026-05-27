# Troubleshooting

Common issues when using the Istio Terraform provider.

## Provider cannot connect to the cluster

**Symptoms:** `Kubernetes config failed`, `connection refused`, or `Unauthorized`.

**Checks:**

- `kubectl cluster-info` works with the same kubeconfig/context
- Provider `config_path` / `config_context` match your environment
- For CI: `host`, `token`, and `cluster_ca_certificate` are correct

## CRD not found (404) on create or read

**Symptoms:** `the server could not find the requested resource`.

**Checks:**

- Istio is installed and CRDs exist: `kubectl get crd | grep istio`
- Manifest `apiVersion` matches installed Istio (e.g. `networking.istio.io/v1`)
- You are managing an **Istio** `Gateway` (`networking.istio.io`), not a Kubernetes Gateway API `Gateway` (`gateway.networking.k8s.io`)

## Plan shows perpetual `manifest` changes

**Causes:**

- Configuration differs from normalized cluster state (empty fields pruned, key order)
- Secret fields: state has `(redacted)` while config has plaintext — usually suppressed; if not, align config or accept one-time apply

**Mitigations:**

- Run `terraform plan` after import to refresh state
- Enable `validate_on_plan` only in CI if the API must be reachable at plan time
- Pin Istio version and keep manifests aligned with cluster defaults

## `validate_on_plan` fails during plan

**Symptoms:** Dry-run create/update errors at plan time.

**Checks:**

- API server reachable during plan (not just apply)
- RBAC allows create/update with dry-run on the target namespace
- Manifest is valid Istio configuration

Disable for offline plans:

```hcl
provider "istio" {
  validate_on_plan = false
}
```

## Import succeeded but `manifest` is empty

Run `terraform plan` to trigger refresh. Ensure RBAC allows `get` on the CRD. See [IMPORT.md](IMPORT.md).

## Remote cluster secret import fails

- Import ID is `namespace/cluster-name`, not the Secret object name
- Secret must exist as `istio-remote-secret-<cluster-name>` with label `istio/multiCluster: true`
- Data key must match `cluster_name`

## Conflicts with GitOps (Argo CD, Flux)

This provider uses client **update** semantics. If another controller manages the same object with server-side apply or a different field manager, conflicts may occur.

**Recommendations:**

- One owner per object (either Terraform or GitOps, not both)
- Use Terraform for mesh-wide policy; use GitOps for app-level config, with clear boundaries

## Acceptance tests locally

```bash
export TF_ACC=1
go test ./internal/provider/... -run TestAccPlanOnly -v    # no cluster changes
go test ./internal/provider/... -run TestAccLifecycle -v  # creates/deletes test resources
```

## GitHub Actions failures

### CI: Format check failed

Run `gofmt -w .` locally and commit. CI fails if any `.go` file is not formatted.

### Acceptance Tests or Release: `startup_failure` with no jobs

This usually means your GitHub organization blocks third-party Actions before jobs start. CI only uses `actions/checkout` and `actions/setup-go`; acceptance tests install Kind via shell scripts for the same reason.

Ask your org admin to allow required actions, or run acceptance tests locally (see above).

The Release workflow uses `crazy-max/ghaction-import-gpg` and `goreleaser/goreleaser-action`, plus `GPG_PRIVATE_KEY` and `PASSPHRASE` secrets. See [PUBLISHING.md](PUBLISHING.md).

## Getting help

Open an issue with: provider version, Istio version, Kubernetes version, redacted manifest, and full Terraform error (omit secrets).

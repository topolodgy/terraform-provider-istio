# Importing resources

This guide describes how to import existing Istio resources into Terraform state.

## Before you import

1. Ensure the resource already exists in the cluster and your provider/kubeconfig can read it.
2. Write a matching `resource` block in configuration (same `metadata.name` and `metadata.namespace` as the live object).
3. Run `terraform import`, then **`terraform plan`** to refresh `manifest` and confirm no unexpected changes.

Import sets the resource `id`. For CRD resources, the provider fills `manifest` on the **next read** (during `terraform plan` or `apply`).

## Istio CRD resources

All CRD resources (`istio_gateway`, `istio_virtual_service`, etc.) use import ID:

```text
<namespace>/<name>
```

### Example

```bash
terraform import istio_virtual_service.my_vs default/my-app
terraform plan
```

After import, verify:

- `id` is `default/my-app`
- `manifest` is populated (non-empty JSON)
- Plan shows only acceptable differences (normalization may adjust empty fields; secrets in config may differ from redacted state)

### Import + manifest behavior

| Step | What happens |
|------|----------------|
| `terraform import` | Sets `id` from import ID |
| `terraform plan` (refresh) | Reads object from API, normalizes and redacts sensitive fields into `manifest` |
| Your `.tf` `manifest` | Should describe the **desired** configuration; semantic equality ignores redacted drift |

If `manifest` stays empty after import, run `terraform plan` with a valid kubeconfig and RBAC to read the CRD.

## Remote cluster secret

Import ID format:

```text
<namespace>/<cluster-name>
```

The cluster name is the **logical remote cluster name** (not the Kubernetes Secret name). The provider expects a secret named `istio-remote-secret-<cluster-name>`.

### Example

```bash
terraform import istio_remote_cluster_secret.remote istio-system/my-remote-cluster
terraform plan
```

On import, the provider reads the secret and populates `server`, `ca_cert_base64`, and `token` from the embedded kubeconfig when present.

## Recommended workflow

```bash
# 1. Add resource block to configuration
# 2. Import
terraform import <resource_address> <import_id>
# 3. Refresh and review
terraform plan
# 4. Align configuration with plan until satisfied
# 5. Apply only if changes are intended
terraform apply
```

## Troubleshooting

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common import errors (404, wrong API kind, RBAC).

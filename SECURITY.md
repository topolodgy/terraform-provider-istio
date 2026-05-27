# Security policy

## Supported versions

Security fixes are provided for the latest minor release of the provider.

| Version | Supported |
|---------|-----------|
| 0.1.x   | Yes       |
| < 0.1   | No        |

## Reporting a vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Report security issues privately to the maintainers of [Topolodgy/terraform-provider-istio](https://github.com/Topolodgy/terraform-provider-istio):

1. Use [GitHub private vulnerability reporting](https://github.com/Topolodgy/terraform-provider-istio/security/advisories/new) if enabled, or
2. Contact the repository administrators directly through your organization's security channel.

Include:

- A description of the issue and impact
- Steps to reproduce
- Affected versions
- Suggested remediation (if known)

We aim to acknowledge reports within **5 business days** and provide a remediation plan or status update within **30 days** for confirmed issues.

## Security considerations when using this provider

- **Terraform state** may contain Istio manifests. Sensitive fields are redacted to `(redacted)` in state after read, but configuration in `.tf` files may still contain secrets — use Vault, environment variables, or CI secrets.
- **Provider `token`** and **`istio_remote_cluster_secret.token`** are marked sensitive in Terraform output.
- **RBAC**: the Kubernetes identity used by the provider needs permissions to manage Istio CRDs and secrets in target namespaces. Follow least privilege.
- **Multi-tenant clusters**: scope Terraform workspaces and kubeconfig contexts to avoid cross-namespace changes.

## Dependency updates

Dependencies are updated through normal development and release processes. Critical CVEs in direct dependencies are addressed in patch or minor releases as appropriate.

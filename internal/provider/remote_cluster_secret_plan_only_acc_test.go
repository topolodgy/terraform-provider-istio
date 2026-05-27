package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccPlanOnly_RemoteClusterSecretDryRun validates remote cluster secret configuration via plan only.
// No secret is created on the cluster.
func TestAccPlanOnly_RemoteClusterSecretDryRun(t *testing.T) {
	testAccPreCheck(t)

	name := fmt.Sprintf("tf-plan-%s", resource.UniqueId())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccRemoteClusterSecretPlanOnlyConfig(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRemoteClusterSecretPlanOnlyConfig(clusterName string) string {
	return fmt.Sprintf(`
provider "istio" {}

resource "istio_remote_cluster_secret" "plan_only" {
  cluster_name   = %q
  server         = "https://remote.example.com:443"
  ca_cert_base64 = "Y2E="
  token          = "plan-only-token"
}
`, clusterName)
}

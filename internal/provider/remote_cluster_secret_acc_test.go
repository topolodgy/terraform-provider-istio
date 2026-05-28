package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLifecycle_RemoteClusterSecret(t *testing.T) {
	testAccPreCheck(t)

	clusterName := fmt.Sprintf("tf-acc-%s", resource.UniqueId())
	secretID := fmt.Sprintf("istio-system/istio-remote-secret-%s", clusterName)
	importID := fmt.Sprintf("istio-system/%s", clusterName)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccRemoteClusterSecretConfig(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("istio_remote_cluster_secret.test", "id", secretID),
					resource.TestCheckResourceAttr("istio_remote_cluster_secret.test", "cluster_name", clusterName),
				),
			},
			{
				ResourceName:            "istio_remote_cluster_secret.test",
				ImportState:             true,
				ImportStateId:           importID,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token"},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("istio_remote_cluster_secret.test", "id", secretID),
					resource.TestCheckResourceAttr("istio_remote_cluster_secret.test", "cluster_name", clusterName),
				),
			},
		},
	})
}

func testAccRemoteClusterSecretConfig(clusterName string) string {
	return fmt.Sprintf(`
provider "istio" {}

resource "istio_remote_cluster_secret" "test" {
  cluster_name   = %q
  server         = "https://kubernetes.default.svc"
  ca_cert_base64 = "Y2E="
  token          = "acc-test-token"
  namespace      = "istio-system"

  timeouts {
    create = "10m"
    delete = "10m"
    read   = "10m"
  }
}
`, clusterName)
}

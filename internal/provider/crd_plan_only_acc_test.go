package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccPlanOnly_ServiceEntryDryRun validates a hypothetical ServiceEntry via plan + dry-run only.
// No resources are created or modified on the cluster.
func TestAccPlanOnly_ServiceEntryDryRun(t *testing.T) {
	testAccPlanOnlyCRD(t, "istio_service_entry", "service_entry", testAccServiceEntryManifest)
}

// TestAccPlanOnly_VirtualServiceDryRun validates a hypothetical VirtualService via plan + dry-run only.
func TestAccPlanOnly_VirtualServiceDryRun(t *testing.T) {
	testAccPlanOnlyCRD(t, "istio_virtual_service", "virtual_service", testAccVirtualServiceManifest)
}

// TestAccPlanOnly_PeerAuthenticationDryRun validates a hypothetical PeerAuthentication via plan + dry-run only.
func TestAccPlanOnly_PeerAuthenticationDryRun(t *testing.T) {
	testAccPlanOnlyCRD(t, "istio_peer_authentication", "peer_authentication", testAccPeerAuthenticationManifest)
}

// TestAccPlanOnly_GatewayDryRun validates a hypothetical Gateway via plan + dry-run only.
func TestAccPlanOnly_GatewayDryRun(t *testing.T) {
	testAccPlanOnlyCRD(t, "istio_gateway", "gateway", testAccGatewayManifest)
}

func testAccPlanOnlyCRD(t *testing.T, resourceType, resourceName string, manifestFn func(name string) string) {
	t.Helper()
	testAccPreCheck(t)

	name := fmt.Sprintf("tf-plan-%s", resource.UniqueId())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:             testAccGenericPlanOnlyConfig(resourceType, resourceName, manifestFn(name)),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccGenericPlanOnlyConfig(resourceType, resourceName, manifest string) string {
	return fmt.Sprintf(`
provider "istio" {
  validate_on_plan = true
}

resource %q %q {
  manifest = %s
}
`, resourceType, resourceName, manifest)
}

func testAccServiceEntryManifest(name string) string {
	host := fmt.Sprintf("%s.example.com", name)
	return fmt.Sprintf(`jsonencode({
    apiVersion = "networking.istio.io/v1"
    kind       = "ServiceEntry"
    metadata = {
      name      = %q
      namespace = "default"
    }
    spec = {
      hosts    = [%q]
      location = "MESH_EXTERNAL"
      ports = [
        {
          number   = 443
          name     = "https"
          protocol = "TLS"
        }
      ]
      resolution = "DNS"
    }
  })`, name, host)
}

func testAccVirtualServiceManifest(name string) string {
	return fmt.Sprintf(`jsonencode({
    apiVersion = "networking.istio.io/v1"
    kind       = "VirtualService"
    metadata = {
      name      = %q
      namespace = "default"
    }
    spec = {
      hosts    = ["%s.default.svc.cluster.local"]
      gateways = ["mesh"]
      http = [
        {
          route = [
            {
              destination = {
                host = "%s.default.svc.cluster.local"
              }
            }
          ]
        }
      ]
    }
  })`, name, name, name)
}

func testAccPeerAuthenticationManifest(name string) string {
	return fmt.Sprintf(`jsonencode({
    apiVersion = "security.istio.io/v1"
    kind       = "PeerAuthentication"
    metadata = {
      name      = %q
      namespace = "default"
    }
    spec = {
      mtls = {
        mode = "PERMISSIVE"
      }
    }
  })`, name)
}

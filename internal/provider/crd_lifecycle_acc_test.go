package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLifecycle_ServiceEntry(t *testing.T) {
	testAccLifecycleCRD(
		t,
		"istio_service_entry",
		"test",
		testAccServiceEntryManifest,
		testAccServiceEntryManifestUpdated,
	)
}

func TestAccLifecycle_VirtualService(t *testing.T) {
	testAccLifecycleCRD(
		t,
		"istio_virtual_service",
		"test",
		testAccVirtualServiceManifest,
		testAccVirtualServiceManifestUpdated,
	)
}

func TestAccLifecycle_PeerAuthentication(t *testing.T) {
	testAccLifecycleCRD(
		t,
		"istio_peer_authentication",
		"test",
		testAccPeerAuthenticationManifest,
		testAccPeerAuthenticationManifestUpdated,
	)
}

func TestAccLifecycle_Gateway(t *testing.T) {
	testAccLifecycleCRD(
		t,
		"istio_gateway",
		"test",
		testAccGatewayManifest,
		testAccGatewayManifestUpdated,
	)
}

func testAccLifecycleCRD(
	t *testing.T,
	resourceType, resourceName string,
	manifestFn, manifestUpdatedFn func(name string) string,
) {
	t.Helper()
	testAccPreCheck(t)

	name := fmt.Sprintf("tf-acc-%s", resource.UniqueId())
	id := fmt.Sprintf("default/%s", name)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccCRDResourceConfig(resourceType, resourceName, manifestFn(name)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.%s", resourceType, resourceName), "id", id),
				),
			},
			{
				Config: testAccCRDResourceConfig(resourceType, resourceName, manifestUpdatedFn(name)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.%s", resourceType, resourceName), "id", id),
				),
			},
			{
				ResourceName:            fmt.Sprintf("%s.%s", resourceType, resourceName),
				ImportState:             true,
				ImportStateId:           id,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"manifest"},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(fmt.Sprintf("%s.%s", resourceType, resourceName), "manifest"),
					resource.TestCheckResourceAttr(fmt.Sprintf("%s.%s", resourceType, resourceName), "id", id),
				),
			},
		},
	})
}

func testAccCRDResourceConfig(resourceType, resourceName, manifest string) string {
	return fmt.Sprintf(`
provider "istio" {}

resource %q %q {
  manifest = %s

  timeouts {
    create = "10m"
    update = "10m"
    delete = "10m"
    read   = "10m"
  }
}
`, resourceType, resourceName, manifest)
}

func testAccServiceEntryManifestUpdated(name string) string {
	host := fmt.Sprintf("%s.example.com", name)
	return fmt.Sprintf(`jsonencode({
    apiVersion = "networking.istio.io/v1"
    kind       = "ServiceEntry"
    metadata = {
      name      = %q
      namespace = "default"
      labels = { managed = "terraform-acc" }
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

func testAccVirtualServiceManifestUpdated(name string) string {
	return fmt.Sprintf(`jsonencode({
    apiVersion = "networking.istio.io/v1"
    kind       = "VirtualService"
    metadata = {
      name      = %q
      namespace = "default"
      labels = { managed = "terraform-acc" }
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

func testAccPeerAuthenticationManifestUpdated(name string) string {
	return fmt.Sprintf(`jsonencode({
    apiVersion = "security.istio.io/v1"
    kind       = "PeerAuthentication"
    metadata = {
      name      = %q
      namespace = "default"
      labels = { managed = "terraform-acc" }
    }
    spec = {
      mtls = {
        mode = "PERMISSIVE"
      }
    }
  })`, name)
}

func testAccGatewayManifest(name string) string {
	return fmt.Sprintf(`jsonencode({
    apiVersion = "networking.istio.io/v1"
    kind       = "Gateway"
    metadata = {
      name      = %q
      namespace = "default"
    }
    spec = {
      selector = {
        istio = "ingressgateway"
      }
      servers = [
        {
          port = {
            number   = 80
            name     = "http"
            protocol = "HTTP"
          }
          hosts = ["*"]
        }
      ]
    }
  })`, name)
}

func testAccGatewayManifestUpdated(name string) string {
	return fmt.Sprintf(`jsonencode({
    apiVersion = "networking.istio.io/v1"
    kind       = "Gateway"
    metadata = {
      name      = %q
      namespace = "default"
      labels = { managed = "terraform-acc" }
    }
    spec = {
      selector = {
        istio = "ingressgateway"
      }
      servers = [
        {
          port = {
            number   = 80
            name     = "http"
            protocol = "HTTP"
          }
          hosts = ["*"]
        }
      ]
    }
  })`, name)
}

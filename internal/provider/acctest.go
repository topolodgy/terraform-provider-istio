package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// ProtoV6ProviderFactories returns the provider factory for acceptance tests.
func ProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"istio": providerserver.NewProtocol6WithError(New("test")()),
	}
}

// testAccPreCheck skips acceptance tests unless TF_ACC is set and a kubeconfig is available.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test; set TF_ACC=1 to run tests against a cluster")
	}
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("unable to determine home directory: %v", err)
		}
		kubeconfig = home + "/.kube/config"
	}
	if _, err := os.Stat(kubeconfig); err != nil {
		t.Skipf("Skipping acceptance test; kubeconfig not found at %s", kubeconfig)
	}
}

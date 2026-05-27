package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestIstioProvider_Metadata(t *testing.T) {
	t.Parallel()

	p := New("test")().(*IstioProvider)
	var resp provider.MetadataResponse
	p.Metadata(context.Background(), provider.MetadataRequest{}, &resp)

	if resp.TypeName != "istio" {
		t.Fatalf("unexpected provider type name: got %q want %q", resp.TypeName, "istio")
	}
}

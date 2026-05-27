package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/terraform-provider-istio/istio/internal/helper"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
)

// parsePlanManifest unmarshals and validates a CRD manifest from the Terraform plan.
func (r *istioCRDResource) parsePlanManifest(manifestJSON string) (map[string]interface{}, k8sschema.GroupVersionResource, *unstructured.Unstructured, diag.Diagnostics) {
	var diags diag.Diagnostics

	manifest, err := helper.ParseManifestJSON(manifestJSON)
	if err != nil {
		diags.AddError("Invalid manifest JSON", err.Error())
		return nil, k8sschema.GroupVersionResource{}, nil, diags
	}
	if err := helper.ValidateManifest(manifest, r.kind, true); err != nil {
		diags.AddError("Invalid manifest", err.Error())
		return nil, k8sschema.GroupVersionResource{}, nil, diags
	}

	gvr := helper.GVRFromManifest(r.defaultGVR, manifest)
	obj, err := helper.ManifestToUnstructured(manifest)
	if err != nil {
		diags.AddError("Manifest conversion failed", err.Error())
		return nil, k8sschema.GroupVersionResource{}, nil, diags
	}
	return manifest, gvr, obj, diags
}

// validatePlanDryRun performs a server-side dry-run create or update when enabled on the provider.
func (r *istioCRDResource) validatePlanDryRun(ctx context.Context, plan, state crdModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if r.data == nil || !r.data.ValidateOnPlan {
		return diags
	}

	_, gvr, obj, parseDiags := r.parsePlanManifest(plan.Manifest.ValueString())
	diags.Append(parseDiags...)
	if diags.HasError() {
		return diags
	}

	if state.ID.IsNull() || state.ID.IsUnknown() {
		if err := helper.CreateCRD(ctx, r.data.Dynamic, gvr, obj, true); err != nil {
			diags.AddError(fmt.Sprintf("Plan validation failed for %s (dry-run create)", r.kind), err.Error())
		}
		return diags
	}

	equal, err := helper.ManifestsEqualIgnoringRedacted(plan.Manifest.ValueString(), state.Manifest.ValueString())
	if err != nil {
		diags.AddError("Plan validation failed", err.Error())
		return diags
	}
	if equal {
		return diags
	}

	if err := helper.UpdateCRD(ctx, r.data.Dynamic, gvr, obj, true); err != nil {
		diags.AddError(fmt.Sprintf("Plan validation failed for %s (dry-run update)", r.kind), err.Error())
	}
	return diags
}

package helper

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ManifestStringType is a string type with semantic equality that ignores redacted fields.
type ManifestStringType struct {
	basetypes.StringType
}

var _ basetypes.StringTypable = ManifestStringType{}

func (t ManifestStringType) Equal(o attr.Type) bool {
	other, ok := o.(ManifestStringType)
	return ok && t.StringType.Equal(other.StringType)
}

func (t ManifestStringType) String() string {
	return "ManifestStringType"
}

func (t ManifestStringType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return ManifestStringValue{StringValue: in}, nil
}

func (t ManifestStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t ManifestStringType) ValueType(_ context.Context) attr.Value {
	return ManifestStringValue{}
}

// ManifestStringValue stores manifest JSON with semantic equality for drift suppression.
type ManifestStringValue struct {
	basetypes.StringValue
}

var (
	_ basetypes.StringValuable                   = ManifestStringValue{}
	_ basetypes.StringValuableWithSemanticEquals = ManifestStringValue{}
)

func (v ManifestStringValue) Equal(o attr.Value) bool {
	other, ok := o.(ManifestStringValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

func (v ManifestStringValue) Type(context.Context) attr.Type {
	return ManifestStringType{}
}

func (v ManifestStringValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(ManifestStringValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			fmt.Sprintf("expected %T, got %T", v, newValuable),
		)
		return false, diags
	}

	prior := v.ValueString()
	newStr := newValue.ValueString()

	// Prefer redacted state over plaintext when refreshing from the API.
	if ManifestContainsRedacted(newStr) && !ManifestContainsRedacted(prior) {
		return false, diags
	}

	equal, err := ManifestsEqualIgnoringRedacted(prior, newStr)
	if err != nil {
		diags.AddError("Semantic Equality Check Error", err.Error())
		return false, diags
	}
	return equal, diags
}

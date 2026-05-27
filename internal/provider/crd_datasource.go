package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-provider-istio/istio/internal/helper"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = &istioCRDDataSource{}
	_ datasource.DataSourceWithConfigure = &istioCRDDataSource{}
)

// istioCRDDataSource is a generic Terraform data source for reading any Istio CRD.
type istioCRDDataSource struct {
	data *ProviderData
	def  CRDResourceDef
}

// crdDataSourceModel is the Terraform state model for CRD data sources.
type crdDataSourceModel struct {
	Name      types.String `tfsdk:"name"`
	Namespace types.String `tfsdk:"namespace"`
	ID        types.String `tfsdk:"id"`
	Manifest  types.String `tfsdk:"manifest"`
}

// NewIstioCRDDataSource creates a datasource.DataSource factory for the given CRD definition.
func NewIstioCRDDataSource(def CRDResourceDef) func() datasource.DataSource {
	return func() datasource.DataSource {
		return &istioCRDDataSource{def: def}
	}
}

func (d *istioCRDDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.def.TypeSuffix
}

func (d *istioCRDDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: fmt.Sprintf("Read an existing Istio %s by name and namespace.", d.def.Kind),
		Attributes: map[string]dschema.Attribute{
			"name": dschema.StringAttribute{
				Required:    true,
				Description: fmt.Sprintf("Name of the %s resource.", d.def.Kind),
			},
			"namespace": dschema.StringAttribute{
				Required:    true,
				Description: "Kubernetes namespace.",
			},
			"id": dschema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier in the format `namespace/name`.",
			},
			"manifest": dschema.StringAttribute{
				Computed:    true,
				CustomType:  helper.ManifestStringType{},
				Description: fmt.Sprintf("Full %s manifest as a JSON string. Sensitive fields are redacted to %q.", d.def.Kind, helper.RedactedPlaceholder),
			},
		},
	}
}

func (d *istioCRDDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *ProviderData, got %T", req.ProviderData))
		return
	}
	d.data = data
}

func (d *istioCRDDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config crdDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gvr := helper.GVR(d.def.Group, d.def.Version, d.def.Plural)
	obj, err := helper.GetCRD(ctx, d.data.Dynamic, gvr, config.Namespace.ValueString(), config.Name.ValueString())
	if err != nil {
		if helper.IsNotFound(err) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("%s not found", d.def.Kind),
				fmt.Sprintf("%s/%s", config.Namespace.ValueString(), config.Name.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("Read %s failed", d.def.Kind), err.Error())
		return
	}

	raw, err := helper.ManifestJSONFromObject(obj.Object)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Read %s failed", d.def.Kind), err.Error())
		return
	}

	state := crdDataSourceModel{
		Name:      config.Name,
		Namespace: config.Namespace,
		ID:        types.StringValue(helper.IdFromObject(obj)),
		Manifest:  types.StringValue(string(raw)),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

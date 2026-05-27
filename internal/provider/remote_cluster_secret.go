package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ resource.Resource = &istioRemoteClusterSecretResource{}
var _ resource.ResourceWithConfigure = &istioRemoteClusterSecretResource{}
var _ resource.ResourceWithImportState = &istioRemoteClusterSecretResource{}

type istioRemoteClusterSecretResource struct {
	data *ProviderData
}

func NewIstioRemoteClusterSecretResource() resource.Resource {
	return &istioRemoteClusterSecretResource{}
}

type remoteClusterSecretModel struct {
	ID           types.String   `tfsdk:"id"`
	ClusterName  types.String   `tfsdk:"cluster_name"`
	Server       types.String   `tfsdk:"server"`
	CACertBase64 types.String   `tfsdk:"ca_cert_base64"`
	Token        types.String   `tfsdk:"token"`
	Namespace    types.String   `tfsdk:"namespace"`
	Timeouts     timeouts.Value `tfsdk:"timeouts"`
}

func (r *istioRemoteClusterSecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_remote_cluster_secret"
}

func (r *istioRemoteClusterSecretResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Istio remote cluster secret for multi-cluster service mesh endpoint discovery. " +
			"Creates a Kubernetes Secret with the `istio/multiCluster: true` label containing a kubeconfig " +
			"for the remote cluster, equivalent to `istioctl create-remote-secret`. " +
			"Istiod watches these secrets to discover services on remote clusters.",
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier in the format `namespace/secret-name`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the remote cluster (e.g. `aks-dev`). Used as the secret data key and in the kubeconfig context.",
			},
			"server": schema.StringAttribute{
				Required:    true,
				Description: "Kubernetes API server URL of the remote cluster (e.g. `https://my-cluster.example.com:443`).",
			},
			"ca_cert_base64": schema.StringAttribute{
				Required:    true,
				Description: "Base64-encoded CA certificate for the remote cluster's API server. Typically from the kubeconfig `certificate-authority-data` field.",
			},
			"token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Bearer token for authenticating to the remote cluster. Create with `kubectl create token -n istio-system istiod --context=<remote> --duration=8760h`.",
			},
			"namespace": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("istio-system"),
				Description: "Namespace where the secret is created. Defaults to `istio-system`.",
			},
		},
	}
}

func (r *istioRemoteClusterSecretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *ProviderData, got %T", req.ProviderData))
		return
	}
	r.data = data
}

func (r *istioRemoteClusterSecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan remoteClusterSecretModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := contextWithCreateTimeout(ctx, plan.Timeouts)
	defer cancel()

	ns := plan.Namespace.ValueString()
	if ns == "" {
		ns = "istio-system"
	}
	name := "istio-remote-secret-" + plan.ClusterName.ValueString()

	kubeconfig := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Config",
		"clusters": []map[string]interface{}{
			{
				"name": plan.ClusterName.ValueString(),
				"cluster": map[string]interface{}{
					"server":                     plan.Server.ValueString(),
					"certificate-authority-data": plan.CACertBase64.ValueString(),
				},
			},
		},
		"users": []map[string]interface{}{
			{
				"name": plan.ClusterName.ValueString(),
				"user": map[string]interface{}{"token": plan.Token.ValueString()},
			},
		},
		"contexts": []map[string]interface{}{
			{
				"name": plan.ClusterName.ValueString(),
				"context": map[string]interface{}{
					"cluster": plan.ClusterName.ValueString(),
					"user":    plan.ClusterName.ValueString(),
				},
			},
		},
		"current-context": plan.ClusterName.ValueString(),
	}
	raw, _ := json.Marshal(kubeconfig)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   ns,
			Labels:      map[string]string{"istio/multiCluster": "true"},
			Annotations: map[string]string{"networking.istio.io/cluster": plan.ClusterName.ValueString()},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{plan.ClusterName.ValueString(): raw},
	}
	_, err := r.data.Clientset.CoreV1().Secrets(ns).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Create remote cluster secret failed", err.Error())
		return
	}

	plan.ID = types.StringValue(ns + "/" + name)
	plan.Namespace = types.StringValue(ns)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *istioRemoteClusterSecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state remoteClusterSecretModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := contextWithReadTimeout(ctx, state.Timeouts)
	defer cancel()

	ns := state.Namespace.ValueString()
	name := "istio-remote-secret-" + state.ClusterName.ValueString()
	secret, err := r.data.Clientset.CoreV1().Secrets(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read remote cluster secret failed", err.Error())
		return
	}

	state.ID = types.StringValue(ns + "/" + secret.Name)
	state.Namespace = types.StringValue(secret.Namespace)
	// Do not overwrite server/token/ca in state (sensitive); keep from state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *istioRemoteClusterSecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan remoteClusterSecretModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := contextWithUpdateTimeout(ctx, plan.Timeouts)
	defer cancel()

	ns := plan.Namespace.ValueString()
	if ns == "" {
		ns = "istio-system"
	}
	name := "istio-remote-secret-" + plan.ClusterName.ValueString()

	kubeconfig := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Config",
		"clusters": []map[string]interface{}{
			{"name": plan.ClusterName.ValueString(), "cluster": map[string]interface{}{"server": plan.Server.ValueString(), "certificate-authority-data": plan.CACertBase64.ValueString()}},
		},
		"users": []map[string]interface{}{
			{"name": plan.ClusterName.ValueString(), "user": map[string]interface{}{"token": plan.Token.ValueString()}},
		},
		"contexts": []map[string]interface{}{
			{"name": plan.ClusterName.ValueString(), "context": map[string]interface{}{"cluster": plan.ClusterName.ValueString(), "user": plan.ClusterName.ValueString()}},
		},
		"current-context": plan.ClusterName.ValueString(),
	}
	raw, _ := json.Marshal(kubeconfig)

	existing, err := r.data.Clientset.CoreV1().Secrets(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Get existing secret failed", err.Error())
		return
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       ns,
			ResourceVersion: existing.ResourceVersion,
			Labels:          map[string]string{"istio/multiCluster": "true"},
			Annotations:     map[string]string{"networking.istio.io/cluster": plan.ClusterName.ValueString()},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{plan.ClusterName.ValueString(): raw},
	}
	_, err = r.data.Clientset.CoreV1().Secrets(ns).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Update remote cluster secret failed", err.Error())
		return
	}

	plan.ID = types.StringValue(ns + "/" + name)
	plan.Namespace = types.StringValue(ns)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *istioRemoteClusterSecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state remoteClusterSecretModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := contextWithDeleteTimeout(ctx, state.Timeouts)
	defer cancel()

	ns := state.Namespace.ValueString()
	name := "istio-remote-secret-" + state.ClusterName.ValueString()
	err := r.data.Clientset.CoreV1().Secrets(ns).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete remote cluster secret failed", err.Error())
	}
}

// ImportState supports `terraform import istio_remote_cluster_secret.<name> <namespace>/<cluster-name>`.
// The import ID format is `namespace/cluster-name` (e.g. `istio-system/aks-dev`).
func (r *istioRemoteClusterSecretResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format: namespace/cluster-name (e.g. istio-system/aks-dev), got: %s", req.ID),
		)
		return
	}
	ns := parts[0]
	clusterName := parts[1]

	// Read the secret to verify it exists
	secretName := "istio-remote-secret-" + clusterName
	secret, err := r.data.Clientset.CoreV1().Secrets(ns).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		resp.Diagnostics.AddError("Import failed: secret not found", fmt.Sprintf("%s/%s: %v", ns, secretName, err))
		return
	}

	// Parse the kubeconfig from the secret data to populate state
	kubeconfigData, ok := secret.Data[clusterName]
	if !ok {
		resp.Diagnostics.AddError("Import failed: missing data key", fmt.Sprintf("Secret %s/%s has no data key %q", ns, secretName, clusterName))
		return
	}
	var kubeconfig map[string]interface{}
	if err := json.Unmarshal(kubeconfigData, &kubeconfig); err != nil {
		resp.Diagnostics.AddError("Import failed: invalid kubeconfig", err.Error())
		return
	}

	// Extract server and CA from kubeconfig
	server := ""
	caCert := ""
	token := ""
	if clusters, ok := kubeconfig["clusters"].([]interface{}); ok && len(clusters) > 0 {
		if cluster, ok := clusters[0].(map[string]interface{}); ok {
			if c, ok := cluster["cluster"].(map[string]interface{}); ok {
				if s, ok := c["server"].(string); ok {
					server = s
				}
				if ca, ok := c["certificate-authority-data"].(string); ok {
					caCert = ca
				}
			}
		}
	}
	if users, ok := kubeconfig["users"].([]interface{}); ok && len(users) > 0 {
		if user, ok := users[0].(map[string]interface{}); ok {
			if u, ok := user["user"].(map[string]interface{}); ok {
				if t, ok := u["token"].(string); ok {
					token = t
				}
			}
		}
	}

	state := remoteClusterSecretModel{
		ID:           types.StringValue(ns + "/" + secretName),
		ClusterName:  types.StringValue(clusterName),
		Server:       types.StringValue(server),
		CACertBase64: types.StringValue(caCert),
		Token:        types.StringValue(token),
		Namespace:    types.StringValue(ns),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

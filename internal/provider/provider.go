package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var _ provider.Provider = &IstioProvider{}

type IstioProvider struct {
	version string
}

type istioProviderModel struct {
	ConfigPath     *string `tfsdk:"config_path"`
	ConfigContext  *string `tfsdk:"config_context"`
	InCluster      *bool   `tfsdk:"in_cluster"`
	Host           *string `tfsdk:"host"`
	Token          *string `tfsdk:"token"`
	Insecure       *bool   `tfsdk:"insecure"`
	ClusterCACert  *string `tfsdk:"cluster_ca_certificate"`
	ValidateOnPlan *bool   `tfsdk:"validate_on_plan"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &IstioProvider{version: version}
	}
}

func (p *IstioProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "istio"
	resp.Version = p.version
}

func (p *IstioProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Istio (CRDs and multi-cluster). Connects to Kubernetes where Istio is installed. " +
			"CRD manifest attributes redact sensitive fields to \"(redacted)\" in state after read while non-sensitive fields remain visible in plan output. " +
			"Set validate_on_plan to run server-side dry-run validation during terraform plan.",
		Attributes: map[string]schema.Attribute{
			"config_path": schema.StringAttribute{
				Optional:    true,
				Description: "Path to kubeconfig. Defaults to KUBECONFIG or ~/.kube/config.",
			},
			"config_context": schema.StringAttribute{
				Optional:    true,
				Description: "Context name from kubeconfig.",
			},
			"in_cluster": schema.BoolAttribute{
				Optional:    true,
				Description: "Use in-cluster config (when running inside a pod).",
			},
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "API server host (when not using kubeconfig).",
			},
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Bearer token for API server.",
			},
			"insecure": schema.BoolAttribute{
				Optional:    true,
				Description: "Skip TLS verification.",
			},
			"cluster_ca_certificate": schema.StringAttribute{
				Optional:    true,
				Description: "Base64 cluster CA certificate.",
			},
			"validate_on_plan": schema.BoolAttribute{
				Optional:    true,
				Description: "When true, runs server-side dry-run create/update against the cluster during terraform plan to validate manifests. Requires a reachable API server at plan time.",
			},
		},
	}
}

func (p *IstioProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config istioProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var restConfig *rest.Config
	var err error

	inCluster := config.InCluster != nil && *config.InCluster
	if inCluster {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			resp.Diagnostics.AddError("In-cluster config failed", err.Error())
			return
		}
	} else if config.Host != nil && *config.Host != "" {
		host := *config.Host
		token := ""
		if config.Token != nil {
			token = *config.Token
		}
		insecure := config.Insecure != nil && *config.Insecure
		caCert := ""
		if config.ClusterCACert != nil {
			caCert = *config.ClusterCACert
		}
		restConfig = &rest.Config{
			Host:            host,
			BearerToken:     token,
			TLSClientConfig: rest.TLSClientConfig{Insecure: insecure, CAData: []byte(caCert)},
		}
	} else {
		kubeconfigPath := ""
		if config.ConfigPath != nil {
			kubeconfigPath = *config.ConfigPath
		}
		if kubeconfigPath == "" {
			kubeconfigPath = os.Getenv("KUBECONFIG")
			if kubeconfigPath == "" {
				home, _ := os.UserHomeDir()
				kubeconfigPath = filepath.Join(home, ".kube", "config")
			}
		}
		loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
		overrides := &clientcmd.ConfigOverrides{}
		if config.ConfigContext != nil && *config.ConfigContext != "" {
			overrides.CurrentContext = *config.ConfigContext
		}
		restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides).ClientConfig()
		if err != nil {
			resp.Diagnostics.AddError("Kubernetes config failed", fmt.Sprintf("loading kubeconfig: %v", err))
			return
		}
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		resp.Diagnostics.AddError("Kubernetes client failed", err.Error())
		return
	}
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		resp.Diagnostics.AddError("Dynamic client failed", err.Error())
		return
	}

	validateOnPlan := config.ValidateOnPlan != nil && *config.ValidateOnPlan

	tflog.Info(ctx, "Configured Kubernetes client for Istio provider")
	providerData := &ProviderData{
		Clientset:      clientset,
		Dynamic:        dynamicClient,
		ValidateOnPlan: validateOnPlan,
	}
	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *IstioProvider) Resources(_ context.Context) []func() resource.Resource {
	// Start with the typed resource (remote_cluster_secret has its own implementation).
	resources := []func() resource.Resource{
		NewIstioRemoteClusterSecretResource,
	}
	// Register all Istio CRD resources via the generic factory.
	for _, def := range AllCRDResources() {
		resources = append(resources, NewIstioCRDResource(def))
	}
	return resources
}

func (p *IstioProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	// Register a data source for every CRD resource.
	var dataSources []func() datasource.DataSource
	for _, def := range AllCRDResources() {
		dataSources = append(dataSources, NewIstioCRDDataSource(def))
	}
	return dataSources
}

// ProviderData is stored in resp.ResourceData / resp.DataSourceData and passed to resources/datasources.
type ProviderData struct {
	Clientset      *kubernetes.Clientset
	Dynamic        dynamic.Interface
	ValidateOnPlan bool
}

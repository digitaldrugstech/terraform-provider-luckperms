package provider

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/client"
	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/datasources"
	"github.com/digitaldrugstech/terraform-provider-luckperms/internal/resources"
)

var _ provider.Provider = &LuckPermsProvider{}

// LuckPermsProvider implements the provider.Provider interface.
type LuckPermsProvider struct {
	version string
}

type providerModel struct {
	BaseURL  types.String `tfsdk:"base_url"`
	APIKey   types.String `tfsdk:"api_key"`
	Timeout  types.Int64  `tfsdk:"timeout"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

// New returns a factory function for the provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &LuckPermsProvider{version: version}
	}
}

func (p *LuckPermsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "luckperms"
	resp.Version = p.version
}

func (p *LuckPermsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage LuckPerms permissions, groups, and tracks via the REST API.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Description: "Base URL of the LuckPerms REST API. Can also be set via LUCKPERMS_BASE_URL.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "API key for authentication. Can also be set via LUCKPERMS_API_KEY.",
				Optional:    true,
				Sensitive:   true,
			},
			"timeout": schema.Int64Attribute{
				Description: "HTTP request timeout in seconds. Default: 30.",
				Optional:    true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. Default: false.",
				Optional:    true,
			},
		},
	}
}

func (p *LuckPermsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := os.Getenv("LUCKPERMS_BASE_URL")
	if !config.BaseURL.IsNull() && !config.BaseURL.IsUnknown() {
		baseURL = config.BaseURL.ValueString()
	}
	if baseURL == "" {
		resp.Diagnostics.AddError(
			"Missing base_url",
			"Set base_url in the provider configuration or the LUCKPERMS_BASE_URL environment variable.",
		)
		return
	}

	apiKey := os.Getenv("LUCKPERMS_API_KEY")
	if !config.APIKey.IsNull() && !config.APIKey.IsUnknown() {
		apiKey = config.APIKey.ValueString()
	}

	timeout := 30 * time.Second
	if envTimeout := os.Getenv("LUCKPERMS_TIMEOUT"); envTimeout != "" {
		if secs, err := strconv.Atoi(envTimeout); err == nil {
			timeout = time.Duration(secs) * time.Second
		}
	}
	if !config.Timeout.IsNull() && !config.Timeout.IsUnknown() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	insecure := false
	if env := os.Getenv("LUCKPERMS_INSECURE"); env == "true" || env == "1" {
		insecure = true
	}
	if !config.Insecure.IsNull() && !config.Insecure.IsUnknown() {
		insecure = config.Insecure.ValueBool()
	}

	c := client.New(baseURL, apiKey, timeout, insecure)

	if err := c.Health(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect to LuckPerms REST API",
			"Ensure the API is running and reachable at "+baseURL+". Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *LuckPermsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewGroupResource,
		resources.NewGroupNodesResource,
		resources.NewTrackResource,
	}
}

func (p *LuckPermsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewGroupDataSource,
		datasources.NewGroupsDataSource,
		datasources.NewTrackDataSource,
		datasources.NewTracksDataSource,
	}
}

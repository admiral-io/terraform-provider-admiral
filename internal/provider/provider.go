package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"go.admiral.io/sdk/client"
)

var _ provider.Provider = &admiralProvider{}

type admiralProvider struct {
	version string
}

type admiralProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Token    types.String `tfsdk:"token"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &admiralProvider{
			version: version,
		}
	}
}

func (p *admiralProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "admiral"
	resp.Version = p.version
}

func (p *admiralProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Admiral provider is used to manage [Admiral](https://admiral.io) platform resources.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The Admiral API host. Defaults to `api.admiral.io:443`. Can also be set with the `ADMIRAL_HOST` environment variable.",
			},
			"token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The Admiral API token. Can also be set with the `ADMIRAL_TOKEN` environment variable.",
			},
			"insecure": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Disable TLS verification. Defaults to `false`.",
			},
		},
	}
}

func (p *admiralProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config admiralProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that known values are provided.
	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Admiral API Host",
			"The provider cannot create the Admiral API client as there is an unknown configuration value for the host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ADMIRAL_HOST environment variable.",
		)
	}
	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Admiral API Token",
			"The provider cannot create the Admiral API client as there is an unknown configuration value for the token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ADMIRAL_TOKEN environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve host.
	host := "api.admiral.io:443"
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	} else if v := os.Getenv("ADMIRAL_HOST"); v != "" {
		host = v
	}

	// Default to port 443 if no port is specified.
	if !strings.Contains(host, ":") {
		host += ":443"
	}

	// Resolve token.
	token := os.Getenv("ADMIRAL_TOKEN")
	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Admiral API Token",
			"The provider requires an Admiral API token. Set the `token` attribute in the provider block or the ADMIRAL_TOKEN environment variable.",
		)
		return
	}

	// Resolve insecure.
	insecure := false
	if !config.Insecure.IsNull() {
		insecure = config.Insecure.ValueBool()
	}

	// Create client
	cfg := client.Config{
		HostPort:  host,
		AuthToken: token,
		ConnectionOptions: client.ConnectionOptions{
			Insecure: insecure,
		},
	}

	c, err := client.New(ctx, cfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Admiral Client",
			"An unexpected error occurred when creating the Admiral API client: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *admiralProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewApplicationResource,
	}
}

func (p *admiralProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewApplicationDataSource,
	}
}

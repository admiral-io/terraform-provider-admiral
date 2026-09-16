package provider

import (
	"context"
	"crypto/tls"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/grpc"

	"go.admiral.io/sdk/client"
)

const (
	// defaultServer mirrors the CLI's default; see admiral-cli/cmd/root.go.
	defaultServer = "api.admiral.io:443"

	// defaultTimeout bounds each RPC so a hung API does not hang a plan or
	// apply. The value matches the CLI's DefaultTimeout.
	defaultTimeout = 60 * time.Second

	envServer = "ADMIRAL_SERVER"
	envAPIKey = "ADMIRAL_API_KEY"
)

var _ provider.Provider = &admiralProvider{}

type admiralProvider struct {
	version string
}

type admiralProviderModel struct {
	Server    types.String `tfsdk:"server"`
	APIKey    types.String `tfsdk:"api_key"`
	Insecure  types.Bool   `tfsdk:"insecure"`
	Plaintext types.Bool   `tfsdk:"plaintext"`
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
			"server": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The Admiral API server as `host:port`. Defaults to `" + defaultServer + "`. Can also be set with the `" + envServer + "` environment variable.",
			},
			"api_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The Admiral API key. Can also be set with the `" + envAPIKey + "` environment variable.",
			},
			"insecure": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Connect over TLS but do not verify the server certificate. Defaults to `false`.",
			},
			"plaintext": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Connect without TLS. The API key travels unencrypted, so only use this against a local server. Defaults to `false`.",
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
	if config.Server.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("server"),
			"Unknown Admiral API Server",
			"The provider cannot create the Admiral API client as there is an unknown configuration value for the server. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the "+envServer+" environment variable.",
		)
	}
	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Admiral API Key",
			"The provider cannot create the Admiral API client as there is an unknown configuration value for the API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the "+envAPIKey+" environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve server.
	server := defaultServer
	if !config.Server.IsNull() {
		server = config.Server.ValueString()
	} else if v := os.Getenv(envServer); v != "" {
		server = v
	}

	// Default to port 443 if no port is specified.
	if !strings.Contains(server, ":") {
		server += ":443"
	}

	// Resolve API key.
	apiKey := os.Getenv(envAPIKey)
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Admiral API Key",
			"The provider requires an Admiral API key. Set the `api_key` attribute in the provider block or the "+envAPIKey+" environment variable.",
		)
		return
	}

	// The SDK checks this too, but here it can point at the attribute.
	if err := client.ValidateAuthToken(apiKey); err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Invalid Admiral API Key",
			"The API key is not in the expected format: "+err.Error(),
		)
		return
	}

	// Resolve transport. Plaintext and insecure are distinct promises, as
	// they are in the CLI: plaintext means no TLS at all, insecure means TLS
	// with the server certificate unverified. The SDK's Insecure flag is the
	// former; the latter goes through TLSConfig.
	plaintext := !config.Plaintext.IsNull() && config.Plaintext.ValueBool()
	insecure := !config.Insecure.IsNull() && config.Insecure.ValueBool()

	var tlsConfig *tls.Config
	if insecure && !plaintext {
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: true} //nolint:gosec // that is what insecure asks for
	}

	// Create client. API keys are sent with the Token scheme; Bearer is for
	// login sessions, which the provider does not support.
	cfg := client.Config{
		HostPort:   server,
		AuthToken:  apiKey,
		AuthScheme: client.AuthSchemeToken,
		ConnectionOptions: client.ConnectionOptions{
			Insecure:  plaintext,
			TLSConfig: tlsConfig,
			DialOptions: []grpc.DialOption{
				grpc.WithChainUnaryInterceptor(deadlineInterceptor(defaultTimeout)),
			},
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

// deadlineInterceptor gives every unary call a deadline unless the caller
// already set one.
func deadlineInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if _, has := ctx.Deadline(); !has && timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

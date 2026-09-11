package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

// Ensure the implementation satisfies the provider.Provider interface.
var _ provider.Provider = &CorexProvider{}

// CorexProvider implements the provider.Provider interface.
type CorexProvider struct {
	version string
}

// corexProviderModel maps the provider schema to a Go struct.
type corexProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	TOTPCode types.String `tfsdk:"totp_code"`
	Token    types.String `tfsdk:"token"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

// Metadata returns the provider metadata.
func (p *CorexProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "corex"
	resp.Version = p.version
}

// Schema returns the provider schema.
func (p *CorexProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "Base URL of the coreX Manager API (e.g. https://corex.example.com).",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "OAuth2 username for authentication. Mutually exclusive with token.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "OAuth2 password for authentication. Sensitive.",
				Optional:    true,
				Sensitive:   true,
			},
			"totp_code": schema.StringAttribute{
				Description: "TOTP code if 2FA is enabled. Sensitive.",
				Optional:    true,
				Sensitive:   true,
			},
			"token": schema.StringAttribute{
				Description: "Pre-issued JWT bearer token. Mutually exclusive with username/password. Sensitive.",
				Optional:    true,
				Sensitive:   true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. Default false.",
				Optional:    true,
			},
		},
	}
}

// Configure configures the provider.
func (p *CorexProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config corexProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fall back to env vars if not set in config.
	host := config.Host.ValueString()
	if host == "" {
		host = os.Getenv("COREX_HOST")
	}
	if host == "" {
		resp.Diagnostics.AddError("host is required", "Set the host in the provider config or COREX_HOST env var.")
		return
	}

	username := config.Username.ValueString()
	if username == "" {
		username = os.Getenv("COREX_USERNAME")
	}
	password := config.Password.ValueString()
	if password == "" {
		password = os.Getenv("COREX_PASSWORD")
	}
	totpCode := config.TOTPCode.ValueString()
	if totpCode == "" {
		totpCode = os.Getenv("COREX_TOTP_CODE")
	}
	token := config.Token.ValueString()
	if token == "" {
		token = os.Getenv("COREX_TOKEN")
	}
	insecure := config.Insecure.ValueBool()

	if token == "" && (username == "" || password == "") {
		resp.Diagnostics.AddError(
			"authentication required",
			"Either set token (static JWT) or username+password (OAuth2) in the provider config or env vars.",
		)
		return
	}

	cli, err := client.New(client.Config{
		Host:     host,
		Username: username,
		Password: password,
		TOTPCode: totpCode,
		Token:    token,
		Insecure: insecure,
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to create client", err.Error())
		return
	}

	resp.DataSourceData = cli
	resp.ResourceData = cli
}

// Resources returns the list of resources the provider supports.
func (p *CorexProvider) Resources(_ context.Context) []func() resource.Resource {
	return allResources()
}

// DataSources returns the list of data sources the provider supports.
func (p *CorexProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return allDataSources()
}

// New returns a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CorexProvider{version: version}
	}
}

// formatErr is a helper to format errors with context.
func formatErr(action, resource string, err error) error {
	return fmt.Errorf("%s %s: %w", action, resource, err)
}

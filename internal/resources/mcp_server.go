package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

var _ resource.Resource = &McpServerResource{}
var _ resource.ResourceWithImportState = &McpServerResource{}

type McpServerResource struct {
	cli *client.Client
}

func NewMcpServerResource() resource.Resource {
	return &McpServerResource{}
}

type mcpServerModel struct {
	ID                                types.Int64  `tfsdk:"id"`
	TeamID                            types.Int64  `tfsdk:"team_id"`
	Name                              types.String `tfsdk:"name"`
	DisplayName                       types.String `tfsdk:"display_name"`
	Description                       types.String `tfsdk:"description"`
	URL                               types.String `tfsdk:"url"`
	Enabled                           types.Bool   `tfsdk:"enabled"`
	VerifyTLS                         types.Bool   `tfsdk:"verify_tls"`
	AuthType                          types.String `tfsdk:"auth_type"`
	AuthHeader                        types.String `tfsdk:"auth_header"`
	AuthSecret                        types.String `tfsdk:"auth_secret"`
	HasSecret                         types.Bool   `tfsdk:"has_secret"`
	TimeoutMs                         types.Int64  `tfsdk:"timeout_ms"`
	MaxBodyBytes                      types.Int64  `tfsdk:"max_body_bytes"`
	Namespace                         types.String `tfsdk:"namespace"`
	HealthStatus                      types.String `tfsdk:"health_status"`
	LastSeenAt                        types.String `tfsdk:"last_seen_at"`
	LastError                         types.String `tfsdk:"last_error"`
	LastCatalogAt                     types.String `tfsdk:"last_catalog_at"`
	TransportType                     types.String `tfsdk:"transport_type"`
	Command                           types.String `tfsdk:"command"`
	Args                              types.List   `tfsdk:"args"`
	EnvVars                           types.Map    `tfsdk:"env_vars"`
	HasEnvVars                        types.Bool   `tfsdk:"has_env_vars"`
	EnvVarNames                       types.List   `tfsdk:"env_var_names"`
	PackageManager                    types.String `tfsdk:"package_manager"`
	SourcePackageName                 types.String `tfsdk:"source_package_name"`
	InstalledVersion                  types.String `tfsdk:"installed_version"`
	OAuthEnabled                      types.Bool   `tfsdk:"oauth_enabled"`
	OAuthAuthStatus                   types.String `tfsdk:"oauth_auth_status"`
	OAuthClientID                     types.String `tfsdk:"oauth_client_id"`
	OAuthClientSecret                 types.String `tfsdk:"oauth_client_secret"`
	OAuthScopes                       types.String `tfsdk:"oauth_scopes"`
	OAuthAuthServerMetadataURL        types.String `tfsdk:"oauth_auth_server_metadata_url"`
	OAuthProtectedResourceMetadataURL types.String `tfsdk:"oauth_protected_resource_metadata_url"`
}

func (r *McpServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_server"
}

func (r *McpServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an MCP gateway server in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":                                  idAttr(),
			"team_id":                             intAttr("Team ID that owns this server.", true),
			"name":                                stringAttr("Server name.", true),
			"display_name":                        stringAttr("Human-friendly display name.", false),
			"description":                         stringAttr("Server description.", false),
			"url":                                 stringAttr("Upstream MCP server URL (for streamable_http transport).", false),
			"enabled":                             boolAttr("Whether the server is enabled.", false),
			"verify_tls":                          boolAttr("Verify upstream TLS certificate.", false),
			"auth_type":                           stringAttr("Auth type: none, bearer, header, or oauth.", false),
			"auth_header":                         stringAttr("Auth header name (for header auth type).", false),
			"auth_secret":                         stringAttrSensitive("Auth secret/token (write-only, never returned).", false),
			"has_secret":                          boolAttrComputed("Whether a secret is configured."),
			"timeout_ms":                          intAttr("Request timeout in milliseconds.", false),
			"max_body_bytes":                      intAttr("Max request body size in bytes.", false),
			"namespace":                           stringAttr("Server namespace for tool/resource/prompt namespacing.", false),
			"health_status":                       stringAttrComputed("Last health check status."),
			"last_seen_at":                        stringAttrComputed("Last time the server was seen healthy."),
			"last_error":                          stringAttrComputed("Last health check error message."),
			"last_catalog_at":                     stringAttrComputed("Last time the server catalog was refreshed."),
			"transport_type":                      stringAttr("Transport type: streamable_http or stdio.", false),
			"command":                             stringAttr("Command to run (for stdio transport).", false),
			"args":                                listAttr("Command arguments (for stdio transport).", false),
			"env_vars":                            mapAttrSensitive("Environment variables (write-only, never returned).", false),
			"has_env_vars":                        boolAttrComputed("Whether env vars are configured."),
			"env_var_names":                       listAttrComputed("Names of configured env vars."),
			"package_manager":                     stringAttr("Package manager for marketplace installs (npm or pypi).", false),
			"source_package_name":                 stringAttr("Source package name for marketplace installs.", false),
			"installed_version":                   stringAttrComputed("Installed package version."),
			"oauth_enabled":                       boolAttr("Whether upstream OAuth is enabled.", false),
			"oauth_auth_status":                   stringAttrComputed("OAuth authorization status."),
			"oauth_client_id":                     stringAttr("OAuth client ID.", false),
			"oauth_client_secret":                 stringAttrSensitive("OAuth client secret (write-only, never returned).", false),
			"oauth_scopes":                        stringAttr("OAuth scopes.", false),
			"oauth_auth_server_metadata_url":      stringAttr("OAuth authorization server metadata URL.", false),
			"oauth_protected_resource_metadata_url": stringAttr("OAuth protected resource metadata URL.", false),
		},
	}
}

func (r *McpServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *McpServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mcpServerModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI(ctx)
	result, err := r.cli.CreateMcpServer(ctx, s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create MCP server", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	// Write-only sensitive fields (auth_secret, oauth_client_secret, env_vars) are
	// preserved in the plan since fromAPI doesn't overwrite them (API never returns them).
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mcpServerModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.cli.GetMcpServer(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read MCP server", err.Error())
		return
	}

	// Preserve write-only fields from prior state.
	authSecret := state.AuthSecret
	oauthClientSecret := state.OAuthClientSecret
	envVars := state.EnvVars

	state.fromAPI(ctx, s)

	// Restore write-only fields (API never returns them).
	state.AuthSecret = authSecret
	state.OAuthClientSecret = oauthClientSecret
	state.EnvVars = envVars

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *McpServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan mcpServerModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI(ctx)
	result, err := r.cli.UpdateMcpServer(ctx, int(plan.ID.ValueInt64()), s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update MCP server", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mcpServerModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteMcpServer(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete MCP server", err.Error())
		return
	}
}

func (r *McpServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	servers, err := r.cli.ListMcpServers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list MCP servers for import", err.Error())
		return
	}

	for _, s := range servers {
		if s.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(s.ID)))...)
			return
		}
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("MCP server not found", fmt.Sprintf("No MCP server with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *mcpServerModel) toAPI(ctx context.Context) *client.McpServer {
	s := &client.McpServer{
		TeamID:        int(m.TeamID.ValueInt64()),
		Name:          m.Name.ValueString(),
		Enabled:       m.Enabled.ValueBool(),
		VerifyTLS:     m.VerifyTLS.ValueBool(),
		AuthType:      m.AuthType.ValueString(),
		TimeoutMs:     int(m.TimeoutMs.ValueInt64()),
		MaxBodyBytes:  int(m.MaxBodyBytes.ValueInt64()),
		TransportType: m.TransportType.ValueString(),
		OAuthEnabled:  m.OAuthEnabled.ValueBool(),
	}
	if !m.DisplayName.IsNull() {
		s.DisplayName = stringPtr(m.DisplayName.ValueString())
	}
	if !m.Description.IsNull() {
		s.Description = stringPtr(m.Description.ValueString())
	}
	if !m.URL.IsNull() {
		s.URL = stringPtr(m.URL.ValueString())
	}
	if !m.AuthHeader.IsNull() {
		s.AuthHeader = stringPtr(m.AuthHeader.ValueString())
	}
	if !m.AuthSecret.IsNull() {
		s.AuthSecret = stringPtr(m.AuthSecret.ValueString())
	}
	if !m.Namespace.IsNull() {
		s.Namespace = stringPtr(m.Namespace.ValueString())
	}
	if !m.Command.IsNull() {
		s.Command = stringPtr(m.Command.ValueString())
	}
	if !m.Args.IsNull() {
		s.Args = stringListToSlice(ctx, m.Args)
	}
	if !m.EnvVars.IsNull() {
		s.EnvVars = stringMapToGo(ctx, m.EnvVars)
	}
	if !m.PackageManager.IsNull() {
		s.PackageManager = stringPtr(m.PackageManager.ValueString())
	}
	if !m.SourcePackageName.IsNull() {
		s.SourcePackageName = stringPtr(m.SourcePackageName.ValueString())
	}
	if !m.OAuthClientID.IsNull() {
		s.OAuthClientID = stringPtr(m.OAuthClientID.ValueString())
	}
	if !m.OAuthClientSecret.IsNull() {
		s.OAuthClientSecret = stringPtr(m.OAuthClientSecret.ValueString())
	}
	if !m.OAuthScopes.IsNull() {
		s.OAuthScopes = stringPtr(m.OAuthScopes.ValueString())
	}
	if !m.OAuthAuthServerMetadataURL.IsNull() {
		s.OAuthAuthServerMetadataURL = stringPtr(m.OAuthAuthServerMetadataURL.ValueString())
	}
	if !m.OAuthProtectedResourceMetadataURL.IsNull() {
		s.OAuthProtectedResourceMetadataURL = stringPtr(m.OAuthProtectedResourceMetadataURL.ValueString())
	}
	return s
}

func (m *mcpServerModel) fromAPI(ctx context.Context, s *client.McpServer) {
	m.ID = types.Int64Value(int64(s.ID))
	m.TeamID = types.Int64Value(int64(s.TeamID))
	m.Name = types.StringValue(s.Name)
	m.DisplayName = types.StringValue(emptyIfNil(s.DisplayName))
	m.Description = types.StringValue(emptyIfNil(s.Description))
	m.URL = types.StringValue(emptyIfNil(s.URL))
	m.Enabled = types.BoolValue(s.Enabled)
	m.VerifyTLS = types.BoolValue(s.VerifyTLS)
	m.AuthType = types.StringValue(s.AuthType)
	m.AuthHeader = types.StringValue(emptyIfNil(s.AuthHeader))
	// AuthSecret, EnvVars, OAuthClientSecret are write-only — not read from API.
	m.HasSecret = types.BoolValue(s.HasSecret)
	m.TimeoutMs = types.Int64Value(int64(s.TimeoutMs))
	m.MaxBodyBytes = types.Int64Value(int64(s.MaxBodyBytes))
	m.Namespace = types.StringValue(emptyIfNil(s.Namespace))
	m.HealthStatus = types.StringValue(emptyIfNil(s.HealthStatus))
	m.LastSeenAt = types.StringValue(emptyIfNil(s.LastSeenAt))
	m.LastError = types.StringValue(emptyIfNil(s.LastError))
	m.LastCatalogAt = types.StringValue(emptyIfNil(s.LastCatalogAt))
	m.TransportType = types.StringValue(s.TransportType)
	m.Command = types.StringValue(emptyIfNil(s.Command))
	m.Args = sliceToStringList(ctx, s.Args)
	m.HasEnvVars = types.BoolValue(s.HasEnvVars)
	m.EnvVarNames = sliceToStringList(ctx, s.EnvVarNames)
	m.PackageManager = types.StringValue(emptyIfNil(s.PackageManager))
	m.SourcePackageName = types.StringValue(emptyIfNil(s.SourcePackageName))
	m.InstalledVersion = types.StringValue(emptyIfNil(s.InstalledVersion))
	m.OAuthEnabled = types.BoolValue(s.OAuthEnabled)
	m.OAuthAuthStatus = types.StringValue(emptyIfNil(s.OAuthAuthStatus))
	m.OAuthClientID = types.StringValue(emptyIfNil(s.OAuthClientID))
	m.OAuthScopes = types.StringValue(emptyIfNil(s.OAuthScopes))
	m.OAuthAuthServerMetadataURL = types.StringValue(emptyIfNil(s.OAuthAuthServerMetadataURL))
	m.OAuthProtectedResourceMetadataURL = types.StringValue(emptyIfNil(s.OAuthProtectedResourceMetadataURL))
}

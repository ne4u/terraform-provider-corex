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

var _ resource.Resource = &McpIdentityResource{}
var _ resource.ResourceWithImportState = &McpIdentityResource{}

type McpIdentityResource struct {
	cli *client.Client
}

func NewMcpIdentityResource() resource.Resource {
	return &McpIdentityResource{}
}

type mcpIdentityModel struct {
	ID            types.Int64  `tfsdk:"id"`
	TeamID        types.Int64  `tfsdk:"team_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Subject       types.String `tfsdk:"subject"`
	Kind          types.String `tfsdk:"kind"`
	PatPrefix     types.String `tfsdk:"pat_prefix"`
	JwtIssuer     types.String `tfsdk:"jwt_issuer"`
	JwtAudience   types.String `tfsdk:"jwt_audience"`
	JwtJwksURL    types.String `tfsdk:"jwt_jwks_url"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	ExpiresAt     types.String `tfsdk:"expires_at"`
	IdpSource     types.String `tfsdk:"idp_source"`
	IdpExternalID types.String `tfsdk:"idp_external_id"`
	IdpUserInfo   types.String `tfsdk:"idp_user_info"`
	LastUsedAt    types.String `tfsdk:"last_used_at"`
}

func (r *McpIdentityResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_identity"
}

func (r *McpIdentityResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an MCP gateway identity (PAT or JWT) in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":             idAttr(),
			"team_id":        intAttr("Team ID that owns this identity.", true),
			"name":           stringAttr("Identity name.", true),
			"description":    stringAttr("Identity description.", false),
			"subject":        stringAttr("Subject identifier.", false),
			"kind":           stringAttr("Identity kind: pat or jwt.", false),
			"pat_prefix":     stringAttrComputed("PAT prefix (shown once on creation)."),
			"jwt_issuer":     stringAttr("JWT issuer.", false),
			"jwt_audience":   stringAttr("JWT audience.", false),
			"jwt_jwks_url":   stringAttr("JWT JWKS URL.", false),
			"enabled":        boolAttr("Whether the identity is enabled.", false),
			"expires_at":     stringAttr("Expiration timestamp (ISO 8601).", false),
			"idp_source":     stringAttr("Identity provider source: manual or auth0.", false),
			"idp_external_id": stringAttr("External ID from the identity provider.", false),
			"idp_user_info":  stringAttr("JSON-encoded user info from the identity provider.", false),
			"last_used_at":   stringAttrComputed("Last time this identity was used."),
		},
	}
}

func (r *McpIdentityResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *McpIdentityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mcpIdentityModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	i := plan.toAPI()
	result, err := r.cli.CreateMcpIdentity(ctx, i)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create MCP identity", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpIdentityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mcpIdentityModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	i, err := r.cli.GetMcpIdentity(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read MCP identity", err.Error())
		return
	}

	state.fromAPI(i)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *McpIdentityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan mcpIdentityModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	i := plan.toAPI()
	result, err := r.cli.UpdateMcpIdentity(ctx, int(plan.ID.ValueInt64()), i)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update MCP identity", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpIdentityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mcpIdentityModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteMcpIdentity(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete MCP identity", err.Error())
		return
	}
}

func (r *McpIdentityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	identities, err := r.cli.ListMcpIdentities(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list MCP identities for import", err.Error())
		return
	}

	for _, i := range identities {
		if i.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(i.ID)))...)
			return
		}
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("MCP identity not found", fmt.Sprintf("No MCP identity with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *mcpIdentityModel) toAPI() *client.McpIdentity {
	i := &client.McpIdentity{
		TeamID:    int(m.TeamID.ValueInt64()),
		Name:      m.Name.ValueString(),
		Kind:      m.Kind.ValueString(),
		Enabled:   m.Enabled.ValueBool(),
		IdpSource: m.IdpSource.ValueString(),
	}
	if !m.Description.IsNull() {
		i.Description = stringPtr(m.Description.ValueString())
	}
	if !m.Subject.IsNull() {
		i.Subject = stringPtr(m.Subject.ValueString())
	}
	if !m.JwtIssuer.IsNull() {
		i.JwtIssuer = stringPtr(m.JwtIssuer.ValueString())
	}
	if !m.JwtAudience.IsNull() {
		i.JwtAudience = stringPtr(m.JwtAudience.ValueString())
	}
	if !m.JwtJwksURL.IsNull() {
		i.JwtJwksURL = stringPtr(m.JwtJwksURL.ValueString())
	}
	if !m.ExpiresAt.IsNull() {
		i.ExpiresAt = stringPtr(m.ExpiresAt.ValueString())
	}
	if !m.IdpExternalID.IsNull() {
		i.IdpExternalID = stringPtr(m.IdpExternalID.ValueString())
	}
	if !m.IdpUserInfo.IsNull() {
		i.IdpUserInfo = jsonMapToGo(m.IdpUserInfo.ValueString())
	}
	return i
}

func (m *mcpIdentityModel) fromAPI(i *client.McpIdentity) {
	m.ID = types.Int64Value(int64(i.ID))
	m.TeamID = types.Int64Value(int64(i.TeamID))
	m.Name = types.StringValue(i.Name)
	m.Description = types.StringValue(emptyIfNil(i.Description))
	m.Subject = types.StringValue(emptyIfNil(i.Subject))
	m.Kind = types.StringValue(i.Kind)
	m.PatPrefix = types.StringValue(emptyIfNil(i.PatPrefix))
	m.JwtIssuer = types.StringValue(emptyIfNil(i.JwtIssuer))
	m.JwtAudience = types.StringValue(emptyIfNil(i.JwtAudience))
	m.JwtJwksURL = types.StringValue(emptyIfNil(i.JwtJwksURL))
	m.Enabled = types.BoolValue(i.Enabled)
	m.ExpiresAt = types.StringValue(emptyIfNil(i.ExpiresAt))
	m.IdpSource = types.StringValue(i.IdpSource)
	m.IdpExternalID = types.StringValue(emptyIfNil(i.IdpExternalID))
	m.IdpUserInfo = types.StringValue(goMapToJSON(i.IdpUserInfo))
	m.LastUsedAt = types.StringValue(emptyIfNil(i.LastUsedAt))
}

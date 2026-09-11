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

var _ resource.Resource = &ApiArmorAuthPolicyResource{}
var _ resource.ResourceWithImportState = &ApiArmorAuthPolicyResource{}

// ApiArmorAuthPolicyResource defines the corex_api_armor_auth_policy resource.
type ApiArmorAuthPolicyResource struct {
	cli *client.Client
}

func NewApiArmorAuthPolicyResource() resource.Resource {
	return &ApiArmorAuthPolicyResource{}
}

type apiArmorAuthPolicyModel struct {
	ID              types.Int64  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	ListenerIDs     types.List   `tfsdk:"listener_ids"`
	BackendIDs      types.List   `tfsdk:"backend_ids"`
	AuthType        types.String `tfsdk:"auth_type"`
	JwtAlgorithm    types.String `tfsdk:"jwt_algorithm"`
	JwtSecretEnv    types.String `tfsdk:"jwt_secret_env"`
	JwtJwksURL      types.String `tfsdk:"jwt_jwks_url"`
	JwtIssuer       types.String `tfsdk:"jwt_issuer"`
	JwtAudience     types.String `tfsdk:"jwt_audience"`
	JwtClaimHeaders types.List   `tfsdk:"jwt_claim_headers"`
	ApiKeyHeader    types.String `tfsdk:"api_key_header"`
	ApiKeyListID    types.Int64  `tfsdk:"api_key_list_id"`
	OnFailure       types.String `tfsdk:"on_failure"`
	Enabled         types.Bool   `tfsdk:"enabled"`
}

func (r *ApiArmorAuthPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_armor_auth_policy"
}

func (r *ApiArmorAuthPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an API Armor auth policy in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":                 idAttr(),
			"name":               stringAttr("Policy name.", true),
			"listener_ids":       listIntAttr("Listener IDs.", false),
			"backend_ids":        listIntAttr("Backend IDs.", false),
			"auth_type":          stringAttr("Auth type (jwt, api_key, both).", false),
			"jwt_algorithm":      stringAttr("JWT algorithm.", false),
			"jwt_secret_env":     stringAttr("Env var holding the JWT secret.", false),
			"jwt_jwks_url":       stringAttr("JWKS URL for JWT verification.", false),
			"jwt_issuer":         stringAttr("Expected JWT issuer.", false),
			"jwt_audience":       stringAttr("Expected JWT audience.", false),
			"jwt_claim_headers":  listAttr("JWT claim headers to forward.", false),
			"api_key_header":     stringAttr("API key header name.", false),
			"api_key_list_id":    intAttr("API key list ID.", false),
			"on_failure":         stringAttr("On-failure action (block, challenge, log_only).", false),
			"enabled":            boolAttr("Whether the policy is enabled.", false),
		},
	}
}

func (r *ApiArmorAuthPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *ApiArmorAuthPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiArmorAuthPolicyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := plan.toAPI(ctx)
	result, err := r.cli.CreateApiArmorAuthPolicy(ctx, p)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create api armor auth policy", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiArmorAuthPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiArmorAuthPolicyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := r.cli.GetApiArmorAuthPolicy(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read api armor auth policy", err.Error())
		return
	}

	state.fromAPI(ctx, p)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiArmorAuthPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apiArmorAuthPolicyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := plan.toAPI(ctx)
	result, err := r.cli.UpdateApiArmorAuthPolicy(ctx, int(plan.ID.ValueInt64()), p)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update api armor auth policy", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiArmorAuthPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiArmorAuthPolicyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteApiArmorAuthPolicy(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete api armor auth policy", err.Error())
		return
	}
}

func (r *ApiArmorAuthPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	policies, err := r.cli.ListApiArmorAuthPolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list api armor auth policies for import", err.Error())
		return
	}

	for _, p := range policies {
		if p.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(p.ID)))...)
			return
		}
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Api armor auth policy not found", fmt.Sprintf("No policy with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *apiArmorAuthPolicyModel) toAPI(ctx context.Context) *client.ApiArmorAuthPolicy {
	p := &client.ApiArmorAuthPolicy{
		Name:            m.Name.ValueString(),
		ListenerIDs:     intListToSlice(ctx, m.ListenerIDs),
		BackendIDs:      intListToSlice(ctx, m.BackendIDs),
		AuthType:        m.AuthType.ValueString(),
		JwtAlgorithm:    m.JwtAlgorithm.ValueString(),
		JwtClaimHeaders: stringListToSlice(ctx, m.JwtClaimHeaders),
		OnFailure:       m.OnFailure.ValueString(),
		Enabled:         m.Enabled.ValueBool(),
	}
	if !m.JwtSecretEnv.IsNull() {
		p.JwtSecretEnv = nilIfEmpty(m.JwtSecretEnv.ValueString())
	}
	if !m.JwtJwksURL.IsNull() {
		p.JwtJwksURL = nilIfEmpty(m.JwtJwksURL.ValueString())
	}
	if !m.JwtIssuer.IsNull() {
		p.JwtIssuer = nilIfEmpty(m.JwtIssuer.ValueString())
	}
	if !m.JwtAudience.IsNull() {
		p.JwtAudience = nilIfEmpty(m.JwtAudience.ValueString())
	}
	if !m.ApiKeyHeader.IsNull() {
		p.ApiKeyHeader = nilIfEmpty(m.ApiKeyHeader.ValueString())
	}
	if !m.ApiKeyListID.IsNull() {
		v := int(m.ApiKeyListID.ValueInt64())
		p.ApiKeyListID = &v
	}
	return p
}

func (m *apiArmorAuthPolicyModel) fromAPI(ctx context.Context, p *client.ApiArmorAuthPolicy) {
	m.ID = types.Int64Value(int64(p.ID))
	m.Name = types.StringValue(p.Name)
	m.ListenerIDs = sliceToIntList(ctx, p.ListenerIDs)
	m.BackendIDs = sliceToIntList(ctx, p.BackendIDs)
	m.AuthType = types.StringValue(p.AuthType)
	m.JwtAlgorithm = types.StringValue(p.JwtAlgorithm)
	if p.JwtSecretEnv != nil {
		m.JwtSecretEnv = types.StringValue(*p.JwtSecretEnv)
	} else {
		m.JwtSecretEnv = types.StringNull()
	}
	if p.JwtJwksURL != nil {
		m.JwtJwksURL = types.StringValue(*p.JwtJwksURL)
	} else {
		m.JwtJwksURL = types.StringNull()
	}
	if p.JwtIssuer != nil {
		m.JwtIssuer = types.StringValue(*p.JwtIssuer)
	} else {
		m.JwtIssuer = types.StringNull()
	}
	if p.JwtAudience != nil {
		m.JwtAudience = types.StringValue(*p.JwtAudience)
	} else {
		m.JwtAudience = types.StringNull()
	}
	m.JwtClaimHeaders = sliceToStringList(ctx, p.JwtClaimHeaders)
	if p.ApiKeyHeader != nil {
		m.ApiKeyHeader = types.StringValue(*p.ApiKeyHeader)
	} else {
		m.ApiKeyHeader = types.StringNull()
	}
	if p.ApiKeyListID != nil {
		m.ApiKeyListID = types.Int64Value(int64(*p.ApiKeyListID))
	} else {
		m.ApiKeyListID = types.Int64Null()
	}
	m.OnFailure = types.StringValue(p.OnFailure)
	m.Enabled = types.BoolValue(p.Enabled)
}

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

var _ resource.Resource = &McpPolicyResource{}
var _ resource.ResourceWithImportState = &McpPolicyResource{}

type McpPolicyResource struct {
	cli *client.Client
}

func NewMcpPolicyResource() resource.Resource {
	return &McpPolicyResource{}
}

type mcpPolicyModel struct {
	ID            types.Int64  `tfsdk:"id"`
	TeamID        types.Int64  `tfsdk:"team_id"`
	Name          types.String `tfsdk:"name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Priority      types.Int64  `tfsdk:"priority"`
	Expression    types.String `tfsdk:"expression"`
	ExpressionAST types.String `tfsdk:"expression_ast"`
	Action        types.String `tfsdk:"action"`
	Log           types.Bool   `tfsdk:"log"`
	NoLog         types.Bool   `tfsdk:"no_log"`
}

func (r *McpPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_policy"
}

func (r *McpPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an MCP gateway access policy in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":             idAttr(),
			"team_id":        intAttr("Team ID that owns this policy.", true),
			"name":           stringAttr("Policy name.", true),
			"enabled":        boolAttr("Whether the policy is enabled.", false),
			"priority":       intAttrComputed("Policy priority (assigned by server)."),
			"expression":     stringAttr("Policy expression (CEL or similar).", true),
			"expression_ast": stringAttrComputed("JSON-encoded AST of the expression."),
			"action":         stringAttr("Policy action: allow, deny, skip_dlp, or skip_ratelimit.", false),
			"log":            boolAttr("Whether to log policy matches.", false),
			"no_log":         boolAttr("Suppress logging for this policy.", false),
		},
	}
}

func (r *McpPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *McpPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mcpPolicyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := plan.toAPI()
	result, err := r.cli.CreateMcpPolicy(ctx, p)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create MCP policy", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mcpPolicyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := r.cli.GetMcpPolicy(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read MCP policy", err.Error())
		return
	}

	state.fromAPI(p)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *McpPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan mcpPolicyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := plan.toAPI()
	result, err := r.cli.UpdateMcpPolicy(ctx, int(plan.ID.ValueInt64()), p)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update MCP policy", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mcpPolicyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteMcpPolicy(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete MCP policy", err.Error())
		return
	}
}

func (r *McpPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	policies, err := r.cli.ListMcpPolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list MCP policies for import", err.Error())
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
		resp.Diagnostics.AddError("MCP policy not found", fmt.Sprintf("No MCP policy with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *mcpPolicyModel) toAPI() *client.McpPolicy {
	return &client.McpPolicy{
		TeamID:     int(m.TeamID.ValueInt64()),
		Name:       m.Name.ValueString(),
		Enabled:    m.Enabled.ValueBool(),
		Expression: m.Expression.ValueString(),
		Action:     m.Action.ValueString(),
		Log:        m.Log.ValueBool(),
		NoLog:      m.NoLog.ValueBool(),
	}
}

func (m *mcpPolicyModel) fromAPI(p *client.McpPolicy) {
	m.ID = types.Int64Value(int64(p.ID))
	m.TeamID = types.Int64Value(int64(p.TeamID))
	m.Name = types.StringValue(p.Name)
	m.Enabled = types.BoolValue(p.Enabled)
	m.Priority = types.Int64Value(int64(p.Priority))
	m.Expression = types.StringValue(p.Expression)
	m.ExpressionAST = types.StringValue(goMapToJSON(p.ExpressionAST))
	m.Action = types.StringValue(p.Action)
	m.Log = types.BoolValue(p.Log)
	m.NoLog = types.BoolValue(p.NoLog)
}

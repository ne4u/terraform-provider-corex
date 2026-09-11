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

var _ resource.Resource = &McpGuardrailResource{}
var _ resource.ResourceWithImportState = &McpGuardrailResource{}

type McpGuardrailResource struct {
	cli *client.Client
}

func NewMcpGuardrailResource() resource.Resource {
	return &McpGuardrailResource{}
}

type mcpGuardrailModel struct {
	ID        types.Int64  `tfsdk:"id"`
	TeamID    types.Int64  `tfsdk:"team_id"`
	Name      types.String `tfsdk:"name"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Priority  types.Int64  `tfsdk:"priority"`
	Direction types.String `tfsdk:"direction"`
	Pack      types.String `tfsdk:"pack"`
	FindRegex types.String `tfsdk:"find_regex"`
	Action    types.String `tfsdk:"action"`
}

func (r *McpGuardrailResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_guardrail"
}

func (r *McpGuardrailResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an MCP gateway guardrail in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":         idAttr(),
			"team_id":    intAttr("Team ID that owns this guardrail.", true),
			"name":       stringAttr("Guardrail name.", true),
			"enabled":    boolAttr("Whether the guardrail is enabled.", false),
			"priority":   intAttrComputed("Guardrail priority (assigned by server)."),
			"direction":  stringAttr("Direction: request, response, or both.", false),
			"pack":       stringAttr("Guardrail pack: builtin:jailbreak_v1, builtin:instruction_override, builtin:obfuscation, or custom.", false),
			"find_regex": stringAttr("Custom regex pattern (for custom pack).", false),
			"action":     stringAttr("Action: block, redact, or log.", false),
		},
	}
}

func (r *McpGuardrailResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *McpGuardrailResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mcpGuardrailModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	g := plan.toAPI()
	result, err := r.cli.CreateMcpGuardrail(ctx, g)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create MCP guardrail", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpGuardrailResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mcpGuardrailModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	g, err := r.cli.GetMcpGuardrail(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read MCP guardrail", err.Error())
		return
	}

	state.fromAPI(g)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *McpGuardrailResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan mcpGuardrailModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	g := plan.toAPI()
	result, err := r.cli.UpdateMcpGuardrail(ctx, int(plan.ID.ValueInt64()), g)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update MCP guardrail", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpGuardrailResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mcpGuardrailModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteMcpGuardrail(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete MCP guardrail", err.Error())
		return
	}
}

func (r *McpGuardrailResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	guardrails, err := r.cli.ListMcpGuardrails(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list MCP guardrails for import", err.Error())
		return
	}

	for _, g := range guardrails {
		if g.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(g.ID)))...)
			return
		}
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("MCP guardrail not found", fmt.Sprintf("No MCP guardrail with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *mcpGuardrailModel) toAPI() *client.McpGuardrail {
	g := &client.McpGuardrail{
		TeamID:    int(m.TeamID.ValueInt64()),
		Name:      m.Name.ValueString(),
		Enabled:   m.Enabled.ValueBool(),
		Direction: m.Direction.ValueString(),
		Pack:      m.Pack.ValueString(),
		Action:    m.Action.ValueString(),
	}
	if !m.FindRegex.IsNull() {
		g.FindRegex = stringPtr(m.FindRegex.ValueString())
	}
	return g
}

func (m *mcpGuardrailModel) fromAPI(g *client.McpGuardrail) {
	m.ID = types.Int64Value(int64(g.ID))
	m.TeamID = types.Int64Value(int64(g.TeamID))
	m.Name = types.StringValue(g.Name)
	m.Enabled = types.BoolValue(g.Enabled)
	m.Priority = types.Int64Value(int64(g.Priority))
	m.Direction = types.StringValue(g.Direction)
	m.Pack = types.StringValue(g.Pack)
	m.FindRegex = types.StringValue(emptyIfNil(g.FindRegex))
	m.Action = types.StringValue(g.Action)
}

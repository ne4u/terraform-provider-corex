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

var _ resource.Resource = &McpDlpRuleResource{}
var _ resource.ResourceWithImportState = &McpDlpRuleResource{}

type McpDlpRuleResource struct {
	cli *client.Client
}

func NewMcpDlpRuleResource() resource.Resource {
	return &McpDlpRuleResource{}
}

type mcpDlpRuleModel struct {
	ID          types.Int64  `tfsdk:"id"`
	TeamID      types.Int64  `tfsdk:"team_id"`
	Name        types.String `tfsdk:"name"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Priority    types.Int64  `tfsdk:"priority"`
	Direction   types.String `tfsdk:"direction"`
	Detector    types.String `tfsdk:"detector"`
	FindRegex   types.String `tfsdk:"find_regex"`
	Action      types.String `tfsdk:"action"`
	TokenPrefix types.String `tfsdk:"token_prefix"`
	TokenTTL    types.Int64  `tfsdk:"token_ttl"`
	ApplyTo     types.String `tfsdk:"apply_to"`
}

func (r *McpDlpRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_dlp_rule"
}

func (r *McpDlpRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an MCP gateway DLP rule in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":            idAttr(),
			"team_id":       intAttr("Team ID that owns this DLP rule.", true),
			"name":          stringAttr("DLP rule name.", true),
			"enabled":       boolAttr("Whether the rule is enabled.", false),
			"priority":      intAttrComputed("Rule priority (assigned by server)."),
			"direction":     stringAttr("Direction: request, response, or both.", false),
			"detector":      stringAttr("Detector type: email, phone, ssn, credit_card, ip, aws_key, private_key, github_token, slack_token, or custom.", true),
			"find_regex":    stringAttr("Custom regex pattern (for custom detector).", false),
			"action":        stringAttr("Action: block, redact, or tokenize.", false),
			"token_prefix":  stringAttr("Token prefix for tokenize action.", false),
			"token_ttl":     intAttr("Token TTL in seconds for tokenize action.", false),
			"apply_to":      stringAttr("Apply to: json_strings or all_text.", false),
		},
	}
}

func (r *McpDlpRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *McpDlpRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mcpDlpRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := plan.toAPI()
	result, err := r.cli.CreateMcpDlpRule(ctx, rule)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create MCP DLP rule", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpDlpRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mcpDlpRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.cli.GetMcpDlpRule(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read MCP DLP rule", err.Error())
		return
	}

	state.fromAPI(rule)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *McpDlpRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan mcpDlpRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := plan.toAPI()
	result, err := r.cli.UpdateMcpDlpRule(ctx, int(plan.ID.ValueInt64()), rule)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update MCP DLP rule", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpDlpRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mcpDlpRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteMcpDlpRule(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete MCP DLP rule", err.Error())
		return
	}
}

func (r *McpDlpRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	rules, err := r.cli.ListMcpDlpRules(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list MCP DLP rules for import", err.Error())
		return
	}

	for _, rule := range rules {
		if rule.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(rule.ID)))...)
			return
		}
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("MCP DLP rule not found", fmt.Sprintf("No MCP DLP rule with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *mcpDlpRuleModel) toAPI() *client.McpDlpRule {
	r := &client.McpDlpRule{
		TeamID:    int(m.TeamID.ValueInt64()),
		Name:      m.Name.ValueString(),
		Enabled:   m.Enabled.ValueBool(),
		Direction: m.Direction.ValueString(),
		Detector:  m.Detector.ValueString(),
		Action:    m.Action.ValueString(),
		ApplyTo:   m.ApplyTo.ValueString(),
	}
	if !m.FindRegex.IsNull() {
		r.FindRegex = stringPtr(m.FindRegex.ValueString())
	}
	if !m.TokenPrefix.IsNull() {
		r.TokenPrefix = stringPtr(m.TokenPrefix.ValueString())
	}
	if !m.TokenTTL.IsNull() {
		v := int(m.TokenTTL.ValueInt64())
		r.TokenTTL = &v
	}
	return r
}

func (m *mcpDlpRuleModel) fromAPI(r *client.McpDlpRule) {
	m.ID = types.Int64Value(int64(r.ID))
	m.TeamID = types.Int64Value(int64(r.TeamID))
	m.Name = types.StringValue(r.Name)
	m.Enabled = types.BoolValue(r.Enabled)
	m.Priority = types.Int64Value(int64(r.Priority))
	m.Direction = types.StringValue(r.Direction)
	m.Detector = types.StringValue(r.Detector)
	m.FindRegex = types.StringValue(emptyIfNil(r.FindRegex))
	m.Action = types.StringValue(r.Action)
	m.TokenPrefix = types.StringValue(emptyIfNil(r.TokenPrefix))
	if r.TokenTTL != nil {
		m.TokenTTL = types.Int64Value(int64(*r.TokenTTL))
	} else {
		m.TokenTTL = types.Int64Null()
	}
	m.ApplyTo = types.StringValue(r.ApplyTo)
}

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

var _ resource.Resource = &RiskRulesetResource{}
var _ resource.ResourceWithImportState = &RiskRulesetResource{}

// RiskRulesetResource defines the corex_risk_ruleset resource.
type RiskRulesetResource struct {
	cli *client.Client
}

func NewRiskRulesetResource() resource.Resource {
	return &RiskRulesetResource{}
}

type riskRulesetModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Slug        types.String `tfsdk:"slug"`
	Priority    types.Int64  `tfsdk:"priority"`
	RuleCount   types.Int64  `tfsdk:"rule_count"`
}

func (r *RiskRulesetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_risk_ruleset"
}

func (r *RiskRulesetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a risk ruleset in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":          idAttr(),
			"name":        stringAttr("Ruleset name.", true),
			"description": stringAttr("Ruleset description.", false),
			"enabled":     boolAttr("Whether the ruleset is enabled.", false),
			"slug":        stringAttrComputed("Ruleset slug (computed)."),
			"priority":    intAttrComputed("Ruleset priority (computed)."),
			"rule_count":  intAttrComputed("Number of rules in the ruleset (computed)."),
		},
	}
}

func (r *RiskRulesetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *RiskRulesetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan riskRulesetModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rs := plan.toAPI()
	result, err := r.cli.CreateRiskRuleset(ctx, rs)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create risk ruleset", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RiskRulesetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state riskRulesetModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rs, err := r.cli.GetRiskRuleset(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read risk ruleset", err.Error())
		return
	}

	state.fromAPI(rs)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *RiskRulesetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan riskRulesetModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rs := plan.toAPI()
	result, err := r.cli.UpdateRiskRuleset(ctx, int(plan.ID.ValueInt64()), rs)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update risk ruleset", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RiskRulesetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state riskRulesetModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteRiskRuleset(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete risk ruleset", err.Error())
		return
	}
}

func (r *RiskRulesetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	sets, err := r.cli.ListRiskRulesets(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list risk rulesets for import", err.Error())
		return
	}

	for _, rs := range sets {
		if rs.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(rs.ID)))...)
			return
		}
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Risk ruleset not found", fmt.Sprintf("No ruleset with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *riskRulesetModel) toAPI() *client.RiskRuleset {
	rs := &client.RiskRuleset{
		Name:    m.Name.ValueString(),
		Enabled: m.Enabled.ValueBool(),
	}
	if !m.Description.IsNull() {
		rs.Description = nilIfEmpty(m.Description.ValueString())
	}
	return rs
}

func (m *riskRulesetModel) fromAPI(rs *client.RiskRuleset) {
	m.ID = types.Int64Value(int64(rs.ID))
	m.Name = types.StringValue(rs.Name)
	if rs.Description != nil {
		m.Description = types.StringValue(*rs.Description)
	} else {
		m.Description = types.StringNull()
	}
	m.Enabled = types.BoolValue(rs.Enabled)
	m.Slug = types.StringValue(rs.Slug)
	m.Priority = types.Int64Value(int64(rs.Priority))
	m.RuleCount = types.Int64Value(int64(rs.RuleCount))
}

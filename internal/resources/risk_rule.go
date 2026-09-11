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

var _ resource.Resource = &RiskRuleResource{}
var _ resource.ResourceWithImportState = &RiskRuleResource{}

// RiskRuleResource defines the corex_risk_rule resource.
type RiskRuleResource struct {
	cli *client.Client
}

func NewRiskRuleResource() resource.Resource {
	return &RiskRuleResource{}
}

type riskRuleModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	ListenerIDs types.List   `tfsdk:"listener_ids"`
	Expression  types.String `tfsdk:"expression"`
	Points      types.Int64  `tfsdk:"points"`
	Category    types.String `tfsdk:"category"`
	Log         types.Bool   `tfsdk:"log"`
	RulesetID   types.Int64  `tfsdk:"ruleset_id"`
	Priority    types.Int64  `tfsdk:"priority"`
}

func (r *RiskRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_risk_rule"
}

func (r *RiskRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a risk-scoring security rule in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":           idAttr(),
			"name":         stringAttr("Rule name.", true),
			"enabled":      boolAttr("Whether the rule is enabled.", false),
			"listener_ids": listIntAttr("Listener IDs the rule applies to.", false),
			"expression":   stringAttr("Risk expression.", true),
			"points":       intAttr("Risk points (-99 to 99).", false),
			"category":     stringAttr("Rule category.", false),
			"log":          boolAttr("Whether to log matches.", false),
			"ruleset_id":   intAttr("Owning ruleset ID.", false),
			"priority":     intAttrComputed("Rule priority (computed)."),
		},
	}
}

func (r *RiskRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *RiskRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan riskRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rr := plan.toAPI(ctx)
	result, err := r.cli.CreateRiskRule(ctx, rr)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create risk rule", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RiskRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state riskRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rr, err := r.cli.GetRiskRule(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read risk rule", err.Error())
		return
	}

	state.fromAPI(ctx, rr)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *RiskRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan riskRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rr := plan.toAPI(ctx)
	result, err := r.cli.UpdateRiskRule(ctx, int(plan.ID.ValueInt64()), rr)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update risk rule", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RiskRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state riskRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteRiskRule(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete risk rule", err.Error())
		return
	}
}

func (r *RiskRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	rules, err := r.cli.ListRiskRules(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list risk rules for import", err.Error())
		return
	}

	for _, rr := range rules {
		if rr.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(rr.ID)))...)
			return
		}
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Risk rule not found", fmt.Sprintf("No rule with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *riskRuleModel) toAPI(ctx context.Context) *client.RiskRule {
	rr := &client.RiskRule{
		Name:        m.Name.ValueString(),
		Enabled:     m.Enabled.ValueBool(),
		ListenerIDs: intListToSlice(ctx, m.ListenerIDs),
		Expression:  m.Expression.ValueString(),
		Points:      int(m.Points.ValueInt64()),
		Log:         m.Log.ValueBool(),
		RulesetID:   int(m.RulesetID.ValueInt64()),
	}
	if !m.Category.IsNull() {
		rr.Category = nilIfEmpty(m.Category.ValueString())
	}
	return rr
}

func (m *riskRuleModel) fromAPI(ctx context.Context, rr *client.RiskRule) {
	m.ID = types.Int64Value(int64(rr.ID))
	m.Name = types.StringValue(rr.Name)
	m.Enabled = types.BoolValue(rr.Enabled)
	m.ListenerIDs = sliceToIntList(ctx, rr.ListenerIDs)
	m.Expression = types.StringValue(rr.Expression)
	m.Points = types.Int64Value(int64(rr.Points))
	if rr.Category != nil {
		m.Category = types.StringValue(*rr.Category)
	} else {
		m.Category = types.StringNull()
	}
	m.Log = types.BoolValue(rr.Log)
	m.RulesetID = types.Int64Value(int64(rr.RulesetID))
	m.Priority = types.Int64Value(int64(rr.Priority))
}

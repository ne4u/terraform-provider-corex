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

// Ensure the implementation satisfies the resource.Resource interface.
var _ resource.Resource = &WafRuleResource{}
var _ resource.ResourceWithImportState = &WafRuleResource{}

// WafRuleResource defines the corex_waf_rule resource.
type WafRuleResource struct {
	cli *client.Client
}

func NewWafRuleResource() resource.Resource {
	return &WafRuleResource{}
}

// wafRuleModel maps the Terraform schema to the API model.
type wafRuleModel struct {
	ID                         types.Int64  `tfsdk:"id"`
	ListenerID                 types.Int64  `tfsdk:"listener_id"`
	BackendID                  types.Int64  `tfsdk:"backend_id"`
	Name                       types.String `tfsdk:"name"`
	Enabled                    types.Bool   `tfsdk:"enabled"`
	RuleSet                    types.String `tfsdk:"rule_set"`
	RuleSetVersion             types.String `tfsdk:"rule_set_version"`
	RuleSetURL                 types.String `tfsdk:"rule_set_url"`
	RuleSetSHA256              types.String `tfsdk:"rule_set_sha256"`
	RuleSetAutoUpdate          types.Bool   `tfsdk:"rule_set_auto_update"`
	RuleSetUpdateIntervalHours types.Int64  `tfsdk:"rule_set_update_interval_hours"`
	Engine                     types.String `tfsdk:"engine"`
	ParanoiaLevel              types.Int64  `tfsdk:"paranoia_level"`
	InboundAnomalyThreshold    types.Int64  `tfsdk:"inbound_anomaly_threshold"`
	OutboundAnomalyThreshold   types.Int64  `tfsdk:"outbound_anomaly_threshold"`
	Action                     types.String `tfsdk:"action"`
	RedirectURL                types.String `tfsdk:"redirect_url"`
	StatusCode                 types.Int64  `tfsdk:"status_code"`
	PathPattern                types.String `tfsdk:"path_pattern"`
	HTTPMethods                types.List   `tfsdk:"http_methods"`
	FailOpen                   types.Bool   `tfsdk:"fail_open"`
	SiemIntegrationID          types.Int64  `tfsdk:"siem_integration_id"`
}

func (r *WafRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waf_rule"
}

func (r *WafRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a WAF rule in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":                          idAttr(),
			"listener_id":                 intAttr("Listener ID the WAF rule is attached to.", false),
			"backend_id":                  intAttr("Backend ID the WAF rule is attached to.", false),
			"name":                        stringAttr("WAF rule name.", true),
			"enabled":                     boolAttr("Whether the WAF rule is enabled.", false),
			"rule_set":                    stringAttr("Rule set name.", false),
			"rule_set_version":            stringAttr("Rule set version.", false),
			"rule_set_url":                stringAttr("Rule set download URL.", false),
			"rule_set_sha256":             stringAttr("Rule set SHA256 checksum.", false),
			"rule_set_auto_update":        boolAttr("Enable automatic rule set updates.", false),
			"rule_set_update_interval_hours": intAttr("Auto-update interval in hours.", false),
			"engine":                      stringAttr("WAF engine (modsecurity/coraza).", false),
			"paranoia_level":              intAttr("Paranoia level (1-4).", false),
			"inbound_anomaly_threshold":   intAttr("Inbound anomaly score threshold.", false),
			"outbound_anomaly_threshold":  intAttr("Outbound anomaly score threshold.", false),
			"action":                      stringAttr("Action on rule match.", false),
			"redirect_url":                stringAttr("Redirect URL for redirect action.", false),
			"status_code":                 intAttr("HTTP status code for deny action.", false),
			"path_pattern":                stringAttr("Path pattern to match.", false),
			"http_methods":                listAttr("HTTP methods to inspect.", false),
			"fail_open":                   boolAttr("Fail open if WAF engine is unavailable.", false),
			"siem_integration_id":         intAttr("SIEM integration ID.", false),
		},
	}
}

func (r *WafRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *WafRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan wafRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	w := plan.toAPI(ctx)
	result, err := r.cli.CreateWafRule(ctx, w)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create WAF rule", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *WafRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state wafRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, err := r.cli.GetWafRule(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read WAF rule", err.Error())
		return
	}

	state.fromAPI(ctx, w)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *WafRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan wafRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	w := plan.toAPI(ctx)
	result, err := r.cli.UpdateWafRule(ctx, int(plan.ID.ValueInt64()), w)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update WAF rule", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *WafRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state wafRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteWafRule(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete WAF rule", err.Error())
		return
	}
}

func (r *WafRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to import by name: list all WAF rules, find matching name.
	rules, err := r.cli.ListWafRules(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list WAF rules for import", err.Error())
		return
	}

	for _, w := range rules {
		if w.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(w.ID)))...)
			return
		}
	}

	// Try numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("WAF rule not found", fmt.Sprintf("No WAF rule with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *wafRuleModel) toAPI(ctx context.Context) *client.WafRule {
	w := &client.WafRule{
		ListenerID:                int(m.ListenerID.ValueInt64()),
		BackendID:                 int(m.BackendID.ValueInt64()),
		Name:                      m.Name.ValueString(),
		Enabled:                   m.Enabled.ValueBool(),
		RuleSet:                   m.RuleSet.ValueString(),
		RuleSetVersion:            m.RuleSetVersion.ValueString(),
		RuleSetURL:                m.RuleSetURL.ValueString(),
		RuleSetSHA256:             m.RuleSetSHA256.ValueString(),
		RuleSetAutoUpdate:         m.RuleSetAutoUpdate.ValueBool(),
		RuleSetUpdateIntervalHours: int(m.RuleSetUpdateIntervalHours.ValueInt64()),
		Engine:                    m.Engine.ValueString(),
		ParanoiaLevel:             int(m.ParanoiaLevel.ValueInt64()),
		InboundAnomalyThreshold:   int(m.InboundAnomalyThreshold.ValueInt64()),
		OutboundAnomalyThreshold:  int(m.OutboundAnomalyThreshold.ValueInt64()),
		Action:                    m.Action.ValueString(),
		RedirectURL:               m.RedirectURL.ValueString(),
		StatusCode:                int(m.StatusCode.ValueInt64()),
		PathPattern:               m.PathPattern.ValueString(),
		HTTPMethods:               stringListToSlice(ctx, m.HTTPMethods),
		FailOpen:                  m.FailOpen.ValueBool(),
	}
	if !m.SiemIntegrationID.IsNull() {
		w.SiemIntegrationID = intPtr(int(m.SiemIntegrationID.ValueInt64()))
	}
	return w
}

// fromAPI populates the Terraform model from the API model.
func (m *wafRuleModel) fromAPI(ctx context.Context, w *client.WafRule) {
	m.ID = types.Int64Value(int64(w.ID))
	m.ListenerID = types.Int64Value(int64(w.ListenerID))
	m.BackendID = types.Int64Value(int64(w.BackendID))
	m.Name = types.StringValue(w.Name)
	m.Enabled = types.BoolValue(w.Enabled)
	m.RuleSet = types.StringValue(w.RuleSet)
	m.RuleSetVersion = types.StringValue(w.RuleSetVersion)
	m.RuleSetURL = types.StringValue(w.RuleSetURL)
	m.RuleSetSHA256 = types.StringValue(w.RuleSetSHA256)
	m.RuleSetAutoUpdate = types.BoolValue(w.RuleSetAutoUpdate)
	m.RuleSetUpdateIntervalHours = types.Int64Value(int64(w.RuleSetUpdateIntervalHours))
	m.Engine = types.StringValue(w.Engine)
	m.ParanoiaLevel = types.Int64Value(int64(w.ParanoiaLevel))
	m.InboundAnomalyThreshold = types.Int64Value(int64(w.InboundAnomalyThreshold))
	m.OutboundAnomalyThreshold = types.Int64Value(int64(w.OutboundAnomalyThreshold))
	m.Action = types.StringValue(w.Action)
	m.RedirectURL = types.StringValue(w.RedirectURL)
	m.StatusCode = types.Int64Value(int64(w.StatusCode))
	m.PathPattern = types.StringValue(w.PathPattern)
	m.HTTPMethods = sliceToStringList(ctx, w.HTTPMethods)
	m.FailOpen = types.BoolValue(w.FailOpen)
	if w.SiemIntegrationID != nil {
		m.SiemIntegrationID = types.Int64Value(int64(*w.SiemIntegrationID))
	}
}

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
var _ resource.Resource = &SecurityRuleResource{}
var _ resource.ResourceWithImportState = &SecurityRuleResource{}

// SecurityRuleResource defines the corex_security_rule resource.
type SecurityRuleResource struct {
	cli *client.Client
}

func NewSecurityRuleResource() resource.Resource {
	return &SecurityRuleResource{}
}

// securityRuleModel maps the Terraform schema to the API model.
type securityRuleModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	ListenerIDs  types.List   `tfsdk:"listener_ids"`
	Expression   types.String `tfsdk:"expression"`
	Action       types.String `tfsdk:"action"`
	Log          types.Bool   `tfsdk:"log"`
	NoLog        types.Bool   `tfsdk:"no_log"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
	RedirectURL  types.String `tfsdk:"redirect_url"`
	RedirectCode types.Int64  `tfsdk:"redirect_code"`
	ErrorPageID  types.Int64  `tfsdk:"error_page_id"`
	Priority     types.Int64  `tfsdk:"priority"`
}

func (r *SecurityRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_rule"
}

func (r *SecurityRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a security rule in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":            idAttr(),
			"name":          stringAttr("Security rule name.", true),
			"enabled":       boolAttr("Whether the rule is enabled.", false),
			"listener_ids":  listIntAttr("Listener IDs the rule applies to.", false),
			"expression":    stringAttr("Match expression.", false),
			"action":        stringAttr("Action to take (allow/deny/redirect/etc.).", true),
			"log":           boolAttr("Enable logging.", false),
			"no_log":        boolAttr("Disable logging.", false),
			"status_code":   intAttr("HTTP status code for deny action.", false),
			"redirect_url":  stringAttr("Redirect URL for redirect action.", false),
			"redirect_code": intAttr("Redirect HTTP status code.", false),
			"error_page_id": intAttr("Error page ID for error action.", false),
			"priority":      intAttrComputed("Rule priority (assigned by coreX)."),
		},
	}
}

func (r *SecurityRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *SecurityRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securityRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sr := plan.toAPI(ctx)
	result, err := r.cli.CreateSecurityRule(ctx, sr)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create security rule", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SecurityRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securityRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sr, err := r.cli.GetSecurityRule(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read security rule", err.Error())
		return
	}

	state.fromAPI(ctx, sr)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *SecurityRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan securityRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sr := plan.toAPI(ctx)
	result, err := r.cli.UpdateSecurityRule(ctx, int(plan.ID.ValueInt64()), sr)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update security rule", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SecurityRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state securityRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteSecurityRule(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete security rule", err.Error())
		return
	}
}

func (r *SecurityRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to import by name: list all security rules, find matching name.
	rules, err := r.cli.ListSecurityRules(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list security rules for import", err.Error())
		return
	}

	for _, sr := range rules {
		if sr.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(sr.ID)))...)
			return
		}
	}

	// Try numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Security rule not found", fmt.Sprintf("No security rule with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *securityRuleModel) toAPI(ctx context.Context) *client.SecurityRule {
	sr := &client.SecurityRule{
		Name:        m.Name.ValueString(),
		Enabled:     m.Enabled.ValueBool(),
		ListenerIDs: intListToSlice(ctx, m.ListenerIDs),
		Expression:  m.Expression.ValueString(),
		Action:      m.Action.ValueString(),
		Log:         m.Log.ValueBool(),
		NoLog:       m.NoLog.ValueBool(),
	}
	if !m.StatusCode.IsNull() {
		sr.StatusCode = intPtr(int(m.StatusCode.ValueInt64()))
	}
	if !m.RedirectURL.IsNull() {
		sr.RedirectURL = stringPtr(m.RedirectURL.ValueString())
	}
	if !m.RedirectCode.IsNull() {
		sr.RedirectCode = intPtr(int(m.RedirectCode.ValueInt64()))
	}
	if !m.ErrorPageID.IsNull() {
		sr.ErrorPageID = intPtr(int(m.ErrorPageID.ValueInt64()))
	}
	return sr
}

// fromAPI populates the Terraform model from the API model.
func (m *securityRuleModel) fromAPI(ctx context.Context, sr *client.SecurityRule) {
	m.ID = types.Int64Value(int64(sr.ID))
	m.Name = types.StringValue(sr.Name)
	m.Enabled = types.BoolValue(sr.Enabled)
	m.ListenerIDs = sliceToIntList(ctx, sr.ListenerIDs)
	m.Expression = types.StringValue(sr.Expression)
	m.Action = types.StringValue(sr.Action)
	m.Log = types.BoolValue(sr.Log)
	m.NoLog = types.BoolValue(sr.NoLog)
	if sr.StatusCode != nil {
		m.StatusCode = types.Int64Value(int64(*sr.StatusCode))
	}
	if sr.RedirectURL != nil {
		m.RedirectURL = types.StringValue(*sr.RedirectURL)
	}
	if sr.RedirectCode != nil {
		m.RedirectCode = types.Int64Value(int64(*sr.RedirectCode))
	}
	if sr.ErrorPageID != nil {
		m.ErrorPageID = types.Int64Value(int64(*sr.ErrorPageID))
	}
	m.Priority = types.Int64Value(int64(sr.Priority))
}

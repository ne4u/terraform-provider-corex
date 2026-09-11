package resources

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

// Ensure the implementation satisfies the resource.Resource interface.
var _ resource.Resource = &WafExceptionResource{}
var _ resource.ResourceWithImportState = &WafExceptionResource{}

// WafExceptionResource defines the corex_waf_exception resource.
type WafExceptionResource struct {
	cli *client.Client
}

func NewWafExceptionResource() resource.Resource {
	return &WafExceptionResource{}
}

// wafExceptionModel maps the Terraform schema to the API model.
type wafExceptionModel struct {
	ID          types.Int64  `tfsdk:"id"`
	WafRuleID   types.Int64  `tfsdk:"waf_rule_id"`
	Name        types.String `tfsdk:"name"`
	RuleID      types.String `tfsdk:"rule_id"`
	RuleTag     types.String `tfsdk:"rule_tag"`
	RuleMsg     types.String `tfsdk:"rule_msg"`
	Zone        types.String `tfsdk:"zone"`
	Variable    types.String `tfsdk:"variable"`
	Matcher     types.String `tfsdk:"matcher"`
	Value       types.String `tfsdk:"value"`
	Description types.String `tfsdk:"description"`
	Action      types.String `tfsdk:"action"`
}

func (r *WafExceptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waf_exception"
}

func (r *WafExceptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a WAF exception in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":           idAttr(),
			"waf_rule_id":  intAttr("WAF rule ID the exception applies to.", true),
			"name":         stringAttr("Exception name.", true),
			"rule_id":      stringAttr("ModSecurity rule ID to except.", false),
			"rule_tag":     stringAttr("Rule tag to except.", false),
			"rule_msg":     stringAttr("Rule message to except.", false),
			"zone":         stringAttr("Zone (request_headers/request_body/etc.).", false),
			"variable":     stringAttr("Variable to except.", false),
			"matcher":      stringAttr("Matcher operator.", false),
			"value":        stringAttr("Match value.", false),
			"description":  stringAttr("Exception description.", false),
			"action":       stringAttr("Action (accept/pass/etc.).", false),
		},
	}
}

func (r *WafExceptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *WafExceptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan wafExceptionModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	e := plan.toAPI()
	result, err := r.cli.CreateWafException(ctx, e)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create WAF exception", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *WafExceptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state wafExceptionModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	e, err := r.cli.GetWafException(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read WAF exception", err.Error())
		return
	}

	state.fromAPI(e)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *WafExceptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan wafExceptionModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	e := plan.toAPI()
	result, err := r.cli.UpdateWafException(ctx, int(plan.ID.ValueInt64()), e)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update WAF exception", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *WafExceptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state wafExceptionModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteWafException(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete WAF exception", err.Error())
		return
	}
}

func (r *WafExceptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by numeric ID only.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "WAF exception must be imported by numeric ID")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *wafExceptionModel) toAPI() *client.WafException {
	e := &client.WafException{
		WafRuleID:   int(m.WafRuleID.ValueInt64()),
		Name:        m.Name.ValueString(),
		Zone:        m.Zone.ValueString(),
		Matcher:     m.Matcher.ValueString(),
		Value:       m.Value.ValueString(),
		Description: m.Description.ValueString(),
		Action:      m.Action.ValueString(),
	}
	if !m.RuleID.IsNull() {
		e.RuleID = stringPtr(m.RuleID.ValueString())
	}
	if !m.RuleTag.IsNull() {
		e.RuleTag = stringPtr(m.RuleTag.ValueString())
	}
	if !m.RuleMsg.IsNull() {
		e.RuleMsg = stringPtr(m.RuleMsg.ValueString())
	}
	if !m.Variable.IsNull() {
		e.Variable = stringPtr(m.Variable.ValueString())
	}
	return e
}

// fromAPI populates the Terraform model from the API model.
func (m *wafExceptionModel) fromAPI(e *client.WafException) {
	m.ID = types.Int64Value(int64(e.ID))
	m.WafRuleID = types.Int64Value(int64(e.WafRuleID))
	m.Name = types.StringValue(e.Name)
	if e.RuleID != nil {
		m.RuleID = types.StringValue(*e.RuleID)
	}
	if e.RuleTag != nil {
		m.RuleTag = types.StringValue(*e.RuleTag)
	}
	if e.RuleMsg != nil {
		m.RuleMsg = types.StringValue(*e.RuleMsg)
	}
	m.Zone = types.StringValue(e.Zone)
	if e.Variable != nil {
		m.Variable = types.StringValue(*e.Variable)
	}
	m.Matcher = types.StringValue(e.Matcher)
	m.Value = types.StringValue(e.Value)
	m.Description = types.StringValue(e.Description)
	m.Action = types.StringValue(e.Action)
}

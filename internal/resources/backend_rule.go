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
var _ resource.Resource = &BackendRuleResource{}
var _ resource.ResourceWithImportState = &BackendRuleResource{}

// BackendRuleResource defines the corex_backend_rule resource.
type BackendRuleResource struct {
	cli *client.Client
}

func NewBackendRuleResource() resource.Resource {
	return &BackendRuleResource{}
}

// backendRuleConditionModel maps a nested condition in the Terraform schema.
type backendRuleConditionModel struct {
	ConditionType types.String `tfsdk:"condition_type"`
	ConditionName types.String `tfsdk:"condition_name"`
	Operator      types.String `tfsdk:"operator"`
	Value         types.String `tfsdk:"value"`
	Join          types.String `tfsdk:"join"`
}

// backendRuleModel maps the Terraform schema to the API model.
type backendRuleModel struct {
	ID            types.Int64                  `tfsdk:"id"`
	ListenerID    types.Int64                  `tfsdk:"listener_id"`
	BackendID     types.Int64                  `tfsdk:"backend_id"`
	Name          types.String                 `tfsdk:"name"`
	Priority      types.Int64                  `tfsdk:"priority"`
	ConditionType types.String                 `tfsdk:"condition_type"`
	ConditionName types.String                 `tfsdk:"condition_name"`
	Operator      types.String                 `tfsdk:"operator"`
	Value         types.String                 `tfsdk:"value"`
	Enabled       types.Bool                   `tfsdk:"enabled"`
	Conditions    []backendRuleConditionModel  `tfsdk:"conditions"`
}

func (r *BackendRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backend_rule"
}

func (r *BackendRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a backend routing rule in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":             idAttr(),
			"listener_id":    intAttr("ID of the listener this rule applies to.", true),
			"backend_id":     intAttr("ID of the backend to route to.", true),
			"name":           stringAttr("Rule name.", false),
			"priority":       intAttr("Rule priority (lower is evaluated first).", false),
			"condition_type": stringAttr("Top-level condition type.", false),
			"condition_name": stringAttr("Top-level condition name.", false),
			"operator":       stringAttr("Top-level condition operator.", false),
			"value":          stringAttr("Top-level condition value.", false),
			"enabled":        boolAttr("Whether this rule is enabled.", false),
			"conditions": schema.ListNestedAttribute{
				Description: "List of conditions for this rule.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"condition_type": stringAttr("Condition type.", false),
						"condition_name": stringAttr("Condition name.", false),
						"operator":       stringAttr("Condition operator.", false),
						"value":          stringAttr("Condition value.", false),
						"join":           stringAttr("Join operator with the next condition (and/or).", false),
					},
				},
			},
		},
	}
}

func (r *BackendRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *BackendRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan backendRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := plan.toAPI()
	result, err := r.cli.CreateBackendRule(ctx, rule)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create backend rule", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BackendRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state backendRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.cli.GetBackendRule(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read backend rule", err.Error())
		return
	}

	// Preserve listener_id and backend_id from state since the API result may
	// not echo them reliably.
	listenerID := state.ListenerID
	backendID := state.BackendID
	state.fromAPI(rule)
	state.ListenerID = listenerID
	state.BackendID = backendID
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *BackendRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan backendRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := plan.toAPI()
	result, err := r.cli.UpdateBackendRule(ctx, int(plan.ID.ValueInt64()), rule)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update backend rule", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BackendRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state backendRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteBackendRule(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete backend rule", err.Error())
		return
	}
}

func (r *BackendRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by numeric ID only.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid backend rule ID", fmt.Sprintf("Backend rule must be imported by numeric ID, got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *backendRuleModel) toAPI() *client.BackendRule {
	rule := &client.BackendRule{
		ListenerID:    int(m.ListenerID.ValueInt64()),
		BackendID:     int(m.BackendID.ValueInt64()),
		Priority:      int(m.Priority.ValueInt64()),
		ConditionType: m.ConditionType.ValueString(),
		Operator:      m.Operator.ValueString(),
		Enabled:       m.Enabled.ValueBool(),
	}
	if !m.Name.IsNull() {
		rule.Name = stringPtr(m.Name.ValueString())
	}
	if !m.ConditionName.IsNull() {
		rule.ConditionName = stringPtr(m.ConditionName.ValueString())
	}
	if !m.Value.IsNull() {
		rule.Value = stringPtr(m.Value.ValueString())
	}
	for _, c := range m.Conditions {
		cond := client.BackendRuleCondition{
			ConditionType: c.ConditionType.ValueString(),
			Operator:      c.Operator.ValueString(),
			Join:          c.Join.ValueString(),
		}
		if !c.ConditionName.IsNull() {
			cond.ConditionName = stringPtr(c.ConditionName.ValueString())
		}
		if !c.Value.IsNull() {
			cond.Value = stringPtr(c.Value.ValueString())
		}
		rule.Conditions = append(rule.Conditions, cond)
	}
	return rule
}

// fromAPI populates the Terraform model from the API model.
func (m *backendRuleModel) fromAPI(r *client.BackendRule) {
	m.ID = types.Int64Value(int64(r.ID))
	m.ListenerID = types.Int64Value(int64(r.ListenerID))
	m.BackendID = types.Int64Value(int64(r.BackendID))
	m.Priority = types.Int64Value(int64(r.Priority))
	m.ConditionType = types.StringValue(r.ConditionType)
	m.Operator = types.StringValue(r.Operator)
	m.Enabled = types.BoolValue(r.Enabled)
	if r.Name != nil {
		m.Name = types.StringValue(*r.Name)
	}
	if r.ConditionName != nil {
		m.ConditionName = types.StringValue(*r.ConditionName)
	}
	if r.Value != nil {
		m.Value = types.StringValue(*r.Value)
	}
	m.Conditions = make([]backendRuleConditionModel, 0, len(r.Conditions))
	for _, c := range r.Conditions {
		cm := backendRuleConditionModel{
			ConditionType: types.StringValue(c.ConditionType),
			Operator:      types.StringValue(c.Operator),
			Join:          types.StringValue(c.Join),
		}
		if c.ConditionName != nil {
			cm.ConditionName = types.StringValue(*c.ConditionName)
		}
		if c.Value != nil {
			cm.Value = types.StringValue(*c.Value)
		}
		m.Conditions = append(m.Conditions, cm)
	}
}

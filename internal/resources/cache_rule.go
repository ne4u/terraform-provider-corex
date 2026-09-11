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
var _ resource.Resource = &CacheRuleResource{}
var _ resource.ResourceWithImportState = &CacheRuleResource{}

// CacheRuleResource defines the corex_cache_rule resource.
type CacheRuleResource struct {
	cli *client.Client
}

func NewCacheRuleResource() resource.Resource {
	return &CacheRuleResource{}
}

// cacheRuleModel maps the Terraform schema to the API model.
type cacheRuleModel struct {
	ID            types.Int64  `tfsdk:"id"`
	CacheConfigID types.Int64  `tfsdk:"cache_config_id"`
	MatchType     types.String `tfsdk:"match_type"`
	Pattern       types.String `tfsdk:"pattern"`
	Action        types.String `tfsdk:"action"`
	Tier          types.String `tfsdk:"tier"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Priority      types.Int64  `tfsdk:"priority"`
}

func (r *CacheRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cache_rule"
}

func (r *CacheRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a cache rule for a backend's cache config in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":               idAttr(),
			"cache_config_id":  intAttr("Cache config ID this rule belongs to.", true),
			"match_type":       stringAttr("Match type for the rule.", false),
			"pattern":          stringAttr("Pattern to match.", false),
			"action":           stringAttr("Action to take on match.", false),
			"tier":             stringAttr("Cache tier.", false),
			"enabled":          boolAttr("Whether this rule is enabled.", false),
			"priority":         intAttr("Rule priority.", false),
		},
	}
}

func (r *CacheRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *CacheRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cacheRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := plan.toAPI()
	result, err := r.cli.CreateCacheRule(ctx, int(plan.CacheConfigID.ValueInt64()), rule)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cache rule", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CacheRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cacheRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Cache rules are read by ID via the rules endpoint. We need to find the
	// rule by listing rules for the parent cache config.
	rules, err := r.cli.ListCacheRules(ctx, int(state.CacheConfigID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read cache rule", err.Error())
		return
	}

	var found *client.CacheRule
	for i := range rules {
		if rules[i].ID == int(state.ID.ValueInt64()) {
			found = &rules[i]
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.fromAPI(found)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *CacheRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cacheRuleModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := plan.toAPI()
	result, err := r.cli.UpdateCacheRule(ctx, int(plan.ID.ValueInt64()), rule)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update cache rule", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CacheRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cacheRuleModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteCacheRule(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete cache rule", err.Error())
		return
	}
}

func (r *CacheRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("No cache rule with ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *cacheRuleModel) toAPI() *client.CacheRule {
	return &client.CacheRule{
		CacheConfigID: int(m.CacheConfigID.ValueInt64()),
		MatchType:     m.MatchType.ValueString(),
		Pattern:       m.Pattern.ValueString(),
		Action:        m.Action.ValueString(),
		Tier:          m.Tier.ValueString(),
		Enabled:       m.Enabled.ValueBool(),
		Priority:      int(m.Priority.ValueInt64()),
	}
}

// fromAPI populates the Terraform model from the API model.
func (m *cacheRuleModel) fromAPI(r *client.CacheRule) {
	m.ID = types.Int64Value(int64(r.ID))
	m.CacheConfigID = types.Int64Value(int64(r.CacheConfigID))
	m.MatchType = types.StringValue(r.MatchType)
	m.Pattern = types.StringValue(r.Pattern)
	m.Action = types.StringValue(r.Action)
	m.Tier = types.StringValue(r.Tier)
	m.Enabled = types.BoolValue(r.Enabled)
	m.Priority = types.Int64Value(int64(r.Priority))
}

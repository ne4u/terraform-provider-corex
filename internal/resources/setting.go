package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

// Ensure the implementation satisfies the resource.Resource interface.
var _ resource.Resource = &SettingResource{}
var _ resource.ResourceWithImportState = &SettingResource{}

// SettingResource defines the corex_setting resource.
type SettingResource struct {
	cli *client.Client
}

func NewSettingResource() resource.Resource {
	return &SettingResource{}
}

// settingModel maps the Terraform schema to the API model.
type settingModel struct {
	ID    types.String `tfsdk:"id"`
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (r *SettingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setting"
}

func (r *SettingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a key/value setting in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Setting key (used as the state ID).",
				Computed:    true,
			},
			"key":   stringAttr("Setting key.", true),
			"value": stringAttr("Setting value.", true),
		},
	}
}

func (r *SettingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *SettingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan settingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := plan.Key.ValueString()
	value := plan.Value.ValueString()
	err := r.cli.SetSetting(ctx, key, value)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create setting", err.Error())
		return
	}

	plan.ID = types.StringValue(key)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SettingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state settingModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := state.Key.ValueString()
	s, err := r.cli.GetSetting(ctx, key)
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read setting", err.Error())
		return
	}

	state.ID = types.StringValue(s.Key)
	state.Key = types.StringValue(s.Key)
	state.Value = types.StringValue(s.Value)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *SettingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan settingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := plan.Key.ValueString()
	value := plan.Value.ValueString()
	err := r.cli.SetSetting(ctx, key, value)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update setting", err.Error())
		return
	}

	plan.ID = types.StringValue(key)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SettingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state settingModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteSetting(ctx, state.Key.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete setting", err.Error())
		return
	}
}

func (r *SettingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by key.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), types.StringValue(req.ID))...)
}

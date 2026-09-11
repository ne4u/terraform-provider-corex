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
var _ resource.Resource = &DynamicFeedResource{}
var _ resource.ResourceWithImportState = &DynamicFeedResource{}

// DynamicFeedResource defines the corex_dynamic_feed resource.
type DynamicFeedResource struct {
	cli *client.Client
}

func NewDynamicFeedResource() resource.Resource {
	return &DynamicFeedResource{}
}

// dynamicFeedModel maps the Terraform schema to the API model.
type dynamicFeedModel struct {
	ID                  types.Int64  `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	ListType            types.String `tfsdk:"list_type"`
	URL                 types.String `tfsdk:"url"`
	UpdateIntervalHours types.Int64  `tfsdk:"update_interval_hours"`
	TargetListID        types.Int64  `tfsdk:"target_list_id"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	AutoApply           types.Bool   `tfsdk:"auto_apply"`
	Description         types.String `tfsdk:"description"`
}

func (r *DynamicFeedResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dynamic_feed"
}

func (r *DynamicFeedResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a dynamic feed that populates a security list in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":                    idAttr(),
			"name":                  stringAttr("Feed name.", true),
			"list_type":             stringAttr("Target list type (network, asn, geo, ja4, pattern).", true),
			"url":                   stringAttr("Feed URL.", true),
			"update_interval_hours": intAttr("Update interval in hours.", false),
			"target_list_id":        intAttr("ID of the target security list.", false),
			"enabled":               boolAttr("Whether the feed is enabled.", false),
			"auto_apply":            boolAttr("Whether to auto-apply feed updates.", false),
			"description":           stringAttr("Optional description.", false),
		},
	}
}

func (r *DynamicFeedResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *DynamicFeedResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dynamicFeedModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI()
	result, err := r.cli.CreateDynamicFeed(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create dynamic feed", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DynamicFeedResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dynamicFeedModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	f, err := r.cli.GetDynamicFeed(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read dynamic feed", err.Error())
		return
	}

	state.fromAPI(f)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *DynamicFeedResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dynamicFeedModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI()
	result, err := r.cli.UpdateDynamicFeed(ctx, int(plan.ID.ValueInt64()), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update dynamic feed", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DynamicFeedResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dynamicFeedModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteDynamicFeed(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete dynamic feed", err.Error())
		return
	}
}

func (r *DynamicFeedResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to import by name: list all dynamic feeds, find matching name.
	feeds, err := r.cli.ListDynamicFeeds(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list dynamic feeds for import", err.Error())
		return
	}

	for _, f := range feeds {
		if f.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(f.ID)))...)
			return
		}
	}

	// Try numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Dynamic feed not found", fmt.Sprintf("No dynamic feed with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *dynamicFeedModel) toAPI() *client.DynamicFeed {
	f := &client.DynamicFeed{
		Name:                m.Name.ValueString(),
		ListType:            m.ListType.ValueString(),
		URL:                 m.URL.ValueString(),
		UpdateIntervalHours: int(m.UpdateIntervalHours.ValueInt64()),
		Enabled:             m.Enabled.ValueBool(),
		AutoApply:           m.AutoApply.ValueBool(),
	}
	if !m.TargetListID.IsNull() {
		id := int(m.TargetListID.ValueInt64())
		f.TargetListID = &id
	}
	if !m.Description.IsNull() {
		f.Description = stringPtr(m.Description.ValueString())
	}
	return f
}

// fromAPI populates the Terraform model from the API model.
func (m *dynamicFeedModel) fromAPI(f *client.DynamicFeed) {
	m.ID = types.Int64Value(int64(f.ID))
	m.Name = types.StringValue(f.Name)
	m.ListType = types.StringValue(f.ListType)
	m.URL = types.StringValue(f.URL)
	m.UpdateIntervalHours = types.Int64Value(int64(f.UpdateIntervalHours))
	if f.TargetListID != nil {
		m.TargetListID = types.Int64Value(int64(*f.TargetListID))
	}
	m.Enabled = types.BoolValue(f.Enabled)
	m.AutoApply = types.BoolValue(f.AutoApply)
	if f.Description != nil {
		m.Description = types.StringValue(*f.Description)
	}
}

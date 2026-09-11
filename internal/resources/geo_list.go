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
var _ resource.Resource = &GeoListResource{}
var _ resource.ResourceWithImportState = &GeoListResource{}

// GeoListResource defines the corex_geo_list resource.
type GeoListResource struct {
	cli *client.Client
}

func NewGeoListResource() resource.Resource {
	return &GeoListResource{}
}

// geoListModel maps the Terraform schema to the API model.
type geoListModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (r *GeoListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_geo_list"
}

func (r *GeoListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a geo security list in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":          idAttr(),
			"name":        stringAttr("List name.", true),
			"description": stringAttr("Optional description.", false),
		},
	}
}

func (r *GeoListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *GeoListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan geoListModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI()
	result, err := r.cli.CreateSecurityList(ctx, "geo", body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create geo list", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GeoListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state geoListModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	l, err := r.cli.GetSecurityList(ctx, "geo", int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read geo list", err.Error())
		return
	}

	state.fromAPI(l)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *GeoListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan geoListModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI()
	result, err := r.cli.UpdateSecurityList(ctx, "geo", int(plan.ID.ValueInt64()), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update geo list", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GeoListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state geoListModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteSecurityList(ctx, "geo", int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete geo list", err.Error())
		return
	}
}

func (r *GeoListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	lists, err := r.cli.ListSecurityLists(ctx, "geo")
	if err != nil {
		resp.Diagnostics.AddError("Failed to list geo lists for import", err.Error())
		return
	}

	for _, l := range lists {
		if l.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(l.ID)))...)
			return
		}
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Geo list not found", fmt.Sprintf("No geo list with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *geoListModel) toAPI() *client.SecurityList {
	l := &client.SecurityList{
		Name: m.Name.ValueString(),
	}
	if !m.Description.IsNull() {
		l.Description = stringPtr(m.Description.ValueString())
	}
	return l
}

// fromAPI populates the Terraform model from the API model.
func (m *geoListModel) fromAPI(l *client.SecurityList) {
	m.ID = types.Int64Value(int64(l.ID))
	m.Name = types.StringValue(l.Name)
	if l.Description != nil {
		m.Description = types.StringValue(*l.Description)
	}
}

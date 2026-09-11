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
var _ resource.Resource = &Ja4ListResource{}
var _ resource.ResourceWithImportState = &Ja4ListResource{}

// Ja4ListResource defines the corex_ja4_list resource.
type Ja4ListResource struct {
	cli *client.Client
}

func NewJa4ListResource() resource.Resource {
	return &Ja4ListResource{}
}

// ja4ListModel maps the Terraform schema to the API model.
type ja4ListModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (r *Ja4ListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ja4_list"
}

func (r *Ja4ListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JA4 fingerprint security list in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":          idAttr(),
			"name":        stringAttr("List name.", true),
			"description": stringAttr("Optional description.", false),
		},
	}
}

func (r *Ja4ListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *Ja4ListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ja4ListModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI()
	result, err := r.cli.CreateSecurityList(ctx, "ja4", body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create ja4 list", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *Ja4ListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ja4ListModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	l, err := r.cli.GetSecurityList(ctx, "ja4", int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read ja4 list", err.Error())
		return
	}

	state.fromAPI(l)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *Ja4ListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ja4ListModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI()
	result, err := r.cli.UpdateSecurityList(ctx, "ja4", int(plan.ID.ValueInt64()), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update ja4 list", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *Ja4ListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ja4ListModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteSecurityList(ctx, "ja4", int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete ja4 list", err.Error())
		return
	}
}

func (r *Ja4ListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	lists, err := r.cli.ListSecurityLists(ctx, "ja4")
	if err != nil {
		resp.Diagnostics.AddError("Failed to list ja4 lists for import", err.Error())
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
		resp.Diagnostics.AddError("Ja4 list not found", fmt.Sprintf("No ja4 list with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *ja4ListModel) toAPI() *client.SecurityList {
	l := &client.SecurityList{
		Name: m.Name.ValueString(),
	}
	if !m.Description.IsNull() {
		l.Description = stringPtr(m.Description.ValueString())
	}
	return l
}

// fromAPI populates the Terraform model from the API model.
func (m *ja4ListModel) fromAPI(l *client.SecurityList) {
	m.ID = types.Int64Value(int64(l.ID))
	m.Name = types.StringValue(l.Name)
	if l.Description != nil {
		m.Description = types.StringValue(*l.Description)
	}
}

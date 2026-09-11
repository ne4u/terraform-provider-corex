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
var _ resource.Resource = &LoggedFieldResource{}
var _ resource.ResourceWithImportState = &LoggedFieldResource{}

// LoggedFieldResource defines the corex_logged_field resource.
type LoggedFieldResource struct {
	cli *client.Client
}

func NewLoggedFieldResource() resource.Resource {
	return &LoggedFieldResource{}
}

// loggedFieldModel maps the Terraform schema to the API model.
type loggedFieldModel struct {
	ID         types.Int64  `tfsdk:"id"`
	ListenerID types.Int64  `tfsdk:"listener_id"`
	Name       types.String `tfsdk:"name"`
	Field      types.String `tfsdk:"field"`
	Enabled    types.Bool   `tfsdk:"enabled"`
}

func (r *LoggedFieldResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_logged_field"
}

func (r *LoggedFieldResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a logged field attached to a listener in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":          idAttr(),
			"listener_id": intAttr("Listener ID this logged field belongs to.", true),
			"name":        stringAttr("Logged field name.", true),
			"field":       stringAttr("Field to log.", false),
			"enabled":     boolAttr("Whether this logged field is enabled.", false),
		},
	}
}

func (r *LoggedFieldResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *LoggedFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loggedFieldModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	f := plan.toAPI()
	result, err := r.cli.CreateLoggedField(ctx, f)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create logged field", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *LoggedFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loggedFieldModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	f, err := r.cli.GetLoggedField(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read logged field", err.Error())
		return
	}

	state.fromAPI(f)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *LoggedFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan loggedFieldModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	f := plan.toAPI()
	result, err := r.cli.UpdateLoggedField(ctx, int(plan.ID.ValueInt64()), f)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update logged field", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *LoggedFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loggedFieldModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteLoggedField(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete logged field", err.Error())
		return
	}
}

func (r *LoggedFieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("No logged field with ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *loggedFieldModel) toAPI() *client.LoggedField {
	return &client.LoggedField{
		ListenerID: int(m.ListenerID.ValueInt64()),
		Name:       m.Name.ValueString(),
		Field:      m.Field.ValueString(),
		Enabled:    m.Enabled.ValueBool(),
	}
}

// fromAPI populates the Terraform model from the API model.
func (m *loggedFieldModel) fromAPI(f *client.LoggedField) {
	m.ID = types.Int64Value(int64(f.ID))
	m.ListenerID = types.Int64Value(int64(f.ListenerID))
	m.Name = types.StringValue(f.Name)
	m.Field = types.StringValue(f.Field)
	m.Enabled = types.BoolValue(f.Enabled)
}

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
var _ resource.Resource = &LogDestinationResource{}
var _ resource.ResourceWithImportState = &LogDestinationResource{}

// LogDestinationResource defines the corex_log_destination resource.
type LogDestinationResource struct {
	cli *client.Client
}

func NewLogDestinationResource() resource.Resource {
	return &LogDestinationResource{}
}

// logDestinationModel maps the Terraform schema to the API model.
type logDestinationModel struct {
	ID         types.Int64  `tfsdk:"id"`
	ListenerID types.Int64  `tfsdk:"listener_id"`
	Name       types.String `tfsdk:"name"`
	Target     types.String `tfsdk:"target"`
	Facility   types.String `tfsdk:"facility"`
	Level      types.String `tfsdk:"level"`
	Format     types.String `tfsdk:"format"`
	Enabled    types.Bool   `tfsdk:"enabled"`
}

func (r *LogDestinationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_log_destination"
}

func (r *LogDestinationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a log destination attached to a listener in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":          idAttr(),
			"listener_id": intAttr("Listener ID this log destination belongs to.", true),
			"name":        stringAttr("Log destination name.", true),
			"target":      stringAttr("Log target (e.g. syslog server).", false),
			"facility":    stringAttr("Syslog facility.", false),
			"level":       stringAttr("Log level.", false),
			"format":      stringAttr("Log format.", false),
			"enabled":     boolAttr("Whether this log destination is enabled.", false),
		},
	}
}

func (r *LogDestinationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *LogDestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan logDestinationModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	d := plan.toAPI()
	result, err := r.cli.CreateLogDestination(ctx, d)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create log destination", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *LogDestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state logDestinationModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	d, err := r.cli.GetLogDestination(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read log destination", err.Error())
		return
	}

	state.fromAPI(d)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *LogDestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan logDestinationModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	d := plan.toAPI()
	result, err := r.cli.UpdateLogDestination(ctx, int(plan.ID.ValueInt64()), d)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update log destination", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *LogDestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state logDestinationModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteLogDestination(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete log destination", err.Error())
		return
	}
}

func (r *LogDestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to import by name: list all log destinations, find matching name.
	dests, err := r.cli.ListLogDestinations(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list log destinations for import", err.Error())
		return
	}

	for _, d := range dests {
		if d.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(d.ID)))...)
			return
		}
	}

	// Try numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Log destination not found", fmt.Sprintf("No log destination with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *logDestinationModel) toAPI() *client.LogDestination {
	return &client.LogDestination{
		ListenerID: int(m.ListenerID.ValueInt64()),
		Name:       m.Name.ValueString(),
		Target:     m.Target.ValueString(),
		Facility:   m.Facility.ValueString(),
		Level:      m.Level.ValueString(),
		Format:     m.Format.ValueString(),
		Enabled:    m.Enabled.ValueBool(),
	}
}

// fromAPI populates the Terraform model from the API model.
func (m *logDestinationModel) fromAPI(d *client.LogDestination) {
	m.ID = types.Int64Value(int64(d.ID))
	m.ListenerID = types.Int64Value(int64(d.ListenerID))
	m.Name = types.StringValue(d.Name)
	m.Target = types.StringValue(d.Target)
	m.Facility = types.StringValue(d.Facility)
	m.Level = types.StringValue(d.Level)
	m.Format = types.StringValue(d.Format)
	m.Enabled = types.BoolValue(d.Enabled)
}

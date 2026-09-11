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
var _ resource.Resource = &NetworkListEntryResource{}
var _ resource.ResourceWithImportState = &NetworkListEntryResource{}

// NetworkListEntryResource defines the corex_network_list_entry resource.
type NetworkListEntryResource struct {
	cli *client.Client
}

func NewNetworkListEntryResource() resource.Resource {
	return &NetworkListEntryResource{}
}

// networkListEntryModel maps the Terraform schema to the API model.
type networkListEntryModel struct {
	ID     types.Int64  `tfsdk:"id"`
	ListID types.Int64  `tfsdk:"list_id"`
	Value  types.String `tfsdk:"value"`
	Note   types.String `tfsdk:"note"`
}

func (r *NetworkListEntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_list_entry"
}

func (r *NetworkListEntryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an entry in a network security list in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":      idAttr(),
			"list_id": intAttr("ID of the parent network list.", true),
			"value":   stringAttr("Entry value (e.g. CIDR or IP).", true),
			"note":    stringAttr("Optional note.", false),
		},
	}
}

func (r *NetworkListEntryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *NetworkListEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan networkListEntryModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI()
	result, err := r.cli.CreateSecurityListEntry(ctx, "network", int(plan.ListID.ValueInt64()), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create network list entry", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *NetworkListEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state networkListEntryModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	entries, err := r.cli.ListSecurityListEntries(ctx, "network", int(state.ListID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read network list entries", err.Error())
		return
	}

	var found *client.SecurityListEntry
	for i := range entries {
		if entries[i].ID == int(state.ID.ValueInt64()) {
			found = &entries[i]
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

func (r *NetworkListEntryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan networkListEntryModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI()
	result, err := r.cli.UpdateSecurityListEntry(ctx, "network", int(plan.ID.ValueInt64()), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update network list entry", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *NetworkListEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state networkListEntryModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteSecurityListEntry(ctx, "network", int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete network list entry", err.Error())
		return
	}
}

func (r *NetworkListEntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Network list entry ID must be numeric, got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *networkListEntryModel) toAPI() *client.SecurityListEntry {
	e := &client.SecurityListEntry{
		ListID: int(m.ListID.ValueInt64()),
		Value:  m.Value.ValueString(),
	}
	if !m.Note.IsNull() {
		e.Note = stringPtr(m.Note.ValueString())
	}
	return e
}

// fromAPI populates the Terraform model from the API model.
func (m *networkListEntryModel) fromAPI(e *client.SecurityListEntry) {
	m.ID = types.Int64Value(int64(e.ID))
	m.ListID = types.Int64Value(int64(e.ListID))
	m.Value = types.StringValue(e.Value)
	if e.Note != nil {
		m.Note = types.StringValue(*e.Note)
	}
}

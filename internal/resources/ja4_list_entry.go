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
var _ resource.Resource = &Ja4ListEntryResource{}
var _ resource.ResourceWithImportState = &Ja4ListEntryResource{}

// Ja4ListEntryResource defines the corex_ja4_list_entry resource.
type Ja4ListEntryResource struct {
	cli *client.Client
}

func NewJa4ListEntryResource() resource.Resource {
	return &Ja4ListEntryResource{}
}

// ja4ListEntryModel maps the Terraform schema to the API model.
type ja4ListEntryModel struct {
	ID     types.Int64  `tfsdk:"id"`
	ListID types.Int64  `tfsdk:"list_id"`
	Value  types.String `tfsdk:"value"`
	Note   types.String `tfsdk:"note"`
}

func (r *Ja4ListEntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ja4_list_entry"
}

func (r *Ja4ListEntryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an entry in a JA4 fingerprint security list in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":      idAttr(),
			"list_id": intAttr("ID of the parent ja4 list.", true),
			"value":   stringAttr("Entry value (JA4 fingerprint).", true),
			"note":    stringAttr("Optional note.", false),
		},
	}
}

func (r *Ja4ListEntryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *Ja4ListEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ja4ListEntryModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI()
	result, err := r.cli.CreateSecurityListEntry(ctx, "ja4", int(plan.ListID.ValueInt64()), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create ja4 list entry", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *Ja4ListEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ja4ListEntryModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	entries, err := r.cli.ListSecurityListEntries(ctx, "ja4", int(state.ListID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read ja4 list entries", err.Error())
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

func (r *Ja4ListEntryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ja4ListEntryModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI()
	result, err := r.cli.UpdateSecurityListEntry(ctx, "ja4", int(plan.ID.ValueInt64()), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update ja4 list entry", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *Ja4ListEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ja4ListEntryModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteSecurityListEntry(ctx, "ja4", int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete ja4 list entry", err.Error())
		return
	}
}

func (r *Ja4ListEntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Ja4 list entry ID must be numeric, got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *ja4ListEntryModel) toAPI() *client.SecurityListEntry {
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
func (m *ja4ListEntryModel) fromAPI(e *client.SecurityListEntry) {
	m.ID = types.Int64Value(int64(e.ID))
	m.ListID = types.Int64Value(int64(e.ListID))
	m.Value = types.StringValue(e.Value)
	if e.Note != nil {
		m.Note = types.StringValue(*e.Note)
	}
}

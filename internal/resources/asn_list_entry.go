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
var _ resource.Resource = &AsnListEntryResource{}
var _ resource.ResourceWithImportState = &AsnListEntryResource{}

// AsnListEntryResource defines the corex_asn_list_entry resource.
type AsnListEntryResource struct {
	cli *client.Client
}

func NewAsnListEntryResource() resource.Resource {
	return &AsnListEntryResource{}
}

// asnListEntryModel maps the Terraform schema to the API model.
type asnListEntryModel struct {
	ID     types.Int64  `tfsdk:"id"`
	ListID types.Int64  `tfsdk:"list_id"`
	Value  types.String `tfsdk:"value"`
	Note   types.String `tfsdk:"note"`
}

func (r *AsnListEntryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asn_list_entry"
}

func (r *AsnListEntryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an entry in an ASN security list in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":      idAttr(),
			"list_id": intAttr("ID of the parent asn list.", true),
			"value":   stringAttr("Entry value (ASN number).", true),
			"note":    stringAttr("Optional note.", false),
		},
	}
}

func (r *AsnListEntryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *AsnListEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan asnListEntryModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI()
	result, err := r.cli.CreateSecurityListEntry(ctx, "asn", int(plan.ListID.ValueInt64()), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create asn list entry", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *AsnListEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state asnListEntryModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	entries, err := r.cli.ListSecurityListEntries(ctx, "asn", int(state.ListID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read asn list entries", err.Error())
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

func (r *AsnListEntryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan asnListEntryModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := plan.toAPI()
	result, err := r.cli.UpdateSecurityListEntry(ctx, "asn", int(plan.ID.ValueInt64()), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update asn list entry", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *AsnListEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state asnListEntryModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteSecurityListEntry(ctx, "asn", int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete asn list entry", err.Error())
		return
	}
}

func (r *AsnListEntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Asn list entry ID must be numeric, got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *asnListEntryModel) toAPI() *client.SecurityListEntry {
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
func (m *asnListEntryModel) fromAPI(e *client.SecurityListEntry) {
	m.ID = types.Int64Value(int64(e.ID))
	m.ListID = types.Int64Value(int64(e.ListID))
	m.Value = types.StringValue(e.Value)
	if e.Note != nil {
		m.Note = types.StringValue(*e.Note)
	}
}

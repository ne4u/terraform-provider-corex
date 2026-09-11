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

var _ resource.Resource = &ApiArmorApiKeyListResource{}
var _ resource.ResourceWithImportState = &ApiArmorApiKeyListResource{}

// ApiArmorApiKeyListResource defines the corex_api_armor_api_key_list resource.
type ApiArmorApiKeyListResource struct {
	cli *client.Client
}

func NewApiArmorApiKeyListResource() resource.Resource {
	return &ApiArmorApiKeyListResource{}
}

type apiKeyEntryModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Value types.String `tfsdk:"value"`
	Note  types.String `tfsdk:"note"`
}

type apiArmorApiKeyListModel struct {
	ID          types.Int64        `tfsdk:"id"`
	Name        types.String       `tfsdk:"name"`
	Description types.String       `tfsdk:"description"`
	Entries     []apiKeyEntryModel `tfsdk:"entries"`
}

func (r *ApiArmorApiKeyListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_armor_api_key_list"
}

func (r *ApiArmorApiKeyListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an API Armor API key list in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":          idAttr(),
			"name":        stringAttr("Key list name.", true),
			"description": stringAttr("Description.", false),
			"entries": schema.ListNestedAttribute{
				Description: "API key entries.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":    idAttr(),
						"value": stringAttrSensitive("API key value.", true),
						"note":  stringAttr("Note for this entry.", false),
					},
				},
			},
		},
	}
}

func (r *ApiArmorApiKeyListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *ApiArmorApiKeyListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiArmorApiKeyListModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	l := plan.toAPI()
	result, err := r.cli.CreateApiArmorApiKeyList(ctx, l)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create api armor api key list", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiArmorApiKeyListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiArmorApiKeyListModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	l, err := r.cli.GetApiArmorApiKeyList(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read api armor api key list", err.Error())
		return
	}

	state.fromAPI(l)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiArmorApiKeyListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apiArmorApiKeyListModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	l := plan.toAPI()
	result, err := r.cli.UpdateApiArmorApiKeyList(ctx, int(plan.ID.ValueInt64()), l)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update api armor api key list", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiArmorApiKeyListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiArmorApiKeyListModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteApiArmorApiKeyList(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete api armor api key list", err.Error())
		return
	}
}

func (r *ApiArmorApiKeyListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	lists, err := r.cli.ListApiArmorApiKeyLists(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list api armor api key lists for import", err.Error())
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
		resp.Diagnostics.AddError("Api armor api key list not found", fmt.Sprintf("No list with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *apiArmorApiKeyListModel) toAPI() *client.ApiArmorApiKeyList {
	l := &client.ApiArmorApiKeyList{
		Name: m.Name.ValueString(),
	}
	if !m.Description.IsNull() {
		l.Description = nilIfEmpty(m.Description.ValueString())
	}
	entries := make([]client.ApiKeyEntry, 0, len(m.Entries))
	for _, e := range m.Entries {
		entry := client.ApiKeyEntry{Value: e.Value.ValueString()}
		if !e.Note.IsNull() {
			n := e.Note.ValueString()
			entry.Note = &n
		}
		entries = append(entries, entry)
	}
	l.Entries = entries
	return l
}

func (m *apiArmorApiKeyListModel) fromAPI(l *client.ApiArmorApiKeyList) {
	m.ID = types.Int64Value(int64(l.ID))
	m.Name = types.StringValue(l.Name)
	if l.Description != nil {
		m.Description = types.StringValue(*l.Description)
	} else {
		m.Description = types.StringNull()
	}
	entries := make([]apiKeyEntryModel, 0, len(l.Entries))
	for _, e := range l.Entries {
		em := apiKeyEntryModel{
			ID:    types.Int64Value(int64(e.ID)),
			Value: types.StringValue(e.Value),
		}
		if e.Note != nil {
			em.Note = types.StringValue(*e.Note)
		} else {
			em.Note = types.StringNull()
		}
		entries = append(entries, em)
	}
	m.Entries = entries
}

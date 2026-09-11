package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

var _ resource.Resource = &ApiArmorOpenApiSpecResource{}
var _ resource.ResourceWithImportState = &ApiArmorOpenApiSpecResource{}

// ApiArmorOpenApiSpecResource defines the corex_api_armor_openapi_spec resource.
// There is no PUT endpoint, so all mutable attributes use ForceNew (update =
// delete + recreate).
type ApiArmorOpenApiSpecResource struct {
	cli *client.Client
}

func NewApiArmorOpenApiSpecResource() resource.Resource {
	return &ApiArmorOpenApiSpecResource{}
}

type apiArmorOpenApiSpecModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Version     types.String `tfsdk:"version"`
	Spec        types.String `tfsdk:"spec"`
	ListenerIDs types.List   `tfsdk:"listener_ids"`
	BackendIDs  types.List   `tfsdk:"backend_ids"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	SchemaCount types.Int64  `tfsdk:"schema_count"`
}

func (r *ApiArmorOpenApiSpecResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_armor_openapi_spec"
}

func (r *ApiArmorOpenApiSpecResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an API Armor OpenAPI spec in coreX Manager. No PUT endpoint exists, so changes force replacement.",
		Attributes: map[string]schema.Attribute{
			"id":   idAttr(),
			"name": stringAttr("Spec name.", true),
			"version": stringAttrComputed("Spec version (computed by the API)."),
			"spec": schema.StringAttribute{
				Description: "Raw OpenAPI spec text (JSON or YAML).",
				Required:    true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"listener_ids": schema.ListAttribute{
				Description:  "Listener IDs.",
				Optional:     true,
				ElementType:  types.Int64Type,
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
			},
			"backend_ids": schema.ListAttribute{
				Description:  "Backend IDs.",
				Optional:     true,
				ElementType:  types.Int64Type,
				PlanModifiers: []planmodifier.List{listplanmodifier.RequiresReplace()},
			},
			"enabled":      boolAttrComputed("Whether the spec is enabled."),
			"schema_count": intAttrComputed("Number of schemas extracted from the spec."),
		},
	}
}

func (r *ApiArmorOpenApiSpecResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *ApiArmorOpenApiSpecResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiArmorOpenApiSpecModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI(ctx)
	result, err := r.cli.CreateApiArmorOpenApiSpec(ctx, s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create api armor openapi spec", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiArmorOpenApiSpecResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiArmorOpenApiSpecModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.cli.GetApiArmorOpenApiSpec(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read api armor openapi spec", err.Error())
		return
	}

	state.fromAPI(ctx, s)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update is a no-op because all mutable attributes use ForceNew; the framework
// destroys and recreates the resource instead of calling Update.
func (r *ApiArmorOpenApiSpecResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apiArmorOpenApiSpecModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiArmorOpenApiSpecResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiArmorOpenApiSpecModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteApiArmorOpenApiSpec(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete api armor openapi spec", err.Error())
		return
	}
}

func (r *ApiArmorOpenApiSpecResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	specs, err := r.cli.ListApiArmorOpenApiSpecs(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list api armor openapi specs for import", err.Error())
		return
	}

	for _, s := range specs {
		if s.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(s.ID)))...)
			return
		}
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Api armor openapi spec not found", fmt.Sprintf("No spec with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *apiArmorOpenApiSpecModel) toAPI(ctx context.Context) *client.ApiArmorOpenApiSpec {
	return &client.ApiArmorOpenApiSpec{
		Name:        m.Name.ValueString(),
		Spec:        m.Spec.ValueString(),
		ListenerIDs: intListToSlice(ctx, m.ListenerIDs),
		BackendIDs:  intListToSlice(ctx, m.BackendIDs),
	}
}

func (m *apiArmorOpenApiSpecModel) fromAPI(ctx context.Context, s *client.ApiArmorOpenApiSpec) {
	m.ID = types.Int64Value(int64(s.ID))
	m.Name = types.StringValue(s.Name)
	if s.Version != nil {
		m.Version = types.StringValue(*s.Version)
	} else {
		m.Version = types.StringNull()
	}
	m.Spec = types.StringValue(s.Spec)
	m.ListenerIDs = sliceToIntList(ctx, s.ListenerIDs)
	m.BackendIDs = sliceToIntList(ctx, s.BackendIDs)
	m.Enabled = types.BoolValue(s.Enabled)
	m.SchemaCount = types.Int64Value(int64(s.SchemaCount))
}

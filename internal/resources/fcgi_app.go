package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

// Ensure the implementation satisfies the resource.Resource interface.
var _ resource.Resource = &FcgiAppResource{}
var _ resource.ResourceWithImportState = &FcgiAppResource{}

// FcgiAppResource defines the corex_fcgi_app resource.
type FcgiAppResource struct {
	cli *client.Client
}

func NewFcgiAppResource() resource.Resource {
	return &FcgiAppResource{}
}

// fcgiAppModel maps the Terraform schema to the API model.
type fcgiAppModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Docroot     types.String `tfsdk:"docroot"`
	Index       types.String `tfsdk:"index"`
	PathInfo    types.String `tfsdk:"path_info"`
	KeepConn    types.Bool   `tfsdk:"keep_conn"`
	MpxsConns   types.Bool   `tfsdk:"mpxs_conns"`
	MaxReqs     types.Int64  `tfsdk:"max_reqs"`
	Params      types.List   `tfsdk:"params"`
}

// fcgiParamModel maps a nested param object.
type fcgiParamModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (r *FcgiAppResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fcgi_app"
}

func (r *FcgiAppResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a FastCGI application in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":          idAttr(),
			"name":        stringAttr("FastCGI application name.", true),
			"description": stringAttr("Application description.", false),
			"docroot":     stringAttr("Document root.", false),
			"index":       stringAttr("Index file.", false),
			"path_info":   stringAttr("PATH_INFO value.", false),
			"keep_conn":   boolAttr("Keep FastCGI connections alive.", false),
			"mpxs_conns":  boolAttr("Enable multiplexed connections.", false),
			"max_reqs":    intAttr("Maximum requests per connection.", false),
			"params": schema.ListNestedAttribute{
				Description: "FastCGI parameters.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":  stringAttr("Parameter name.", true),
						"value": stringAttr("Parameter value.", false),
					},
				},
			},
		},
	}
}

func (r *FcgiAppResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *FcgiAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fcgiAppModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	a := plan.toAPI(ctx)
	result, err := r.cli.CreateFcgiApp(ctx, a)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create FastCGI app", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *FcgiAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fcgiAppModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	a, err := r.cli.GetFcgiApp(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read FastCGI app", err.Error())
		return
	}

	state.fromAPI(ctx, a)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *FcgiAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan fcgiAppModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	a := plan.toAPI(ctx)
	result, err := r.cli.UpdateFcgiApp(ctx, int(plan.ID.ValueInt64()), a)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update FastCGI app", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *FcgiAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state fcgiAppModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteFcgiApp(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete FastCGI app", err.Error())
		return
	}
}

func (r *FcgiAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to import by name: list all FastCGI apps, find matching name.
	apps, err := r.cli.ListFcgiApps(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list FastCGI apps for import", err.Error())
		return
	}

	for _, a := range apps {
		if a.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(a.ID)))...)
			return
		}
	}

	// Try numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("FastCGI app not found", fmt.Sprintf("No FastCGI app with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *fcgiAppModel) toAPI(ctx context.Context) *client.FcgiApp {
	return &client.FcgiApp{
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
		Docroot:     m.Docroot.ValueString(),
		Index:       m.Index.ValueString(),
		PathInfo:    m.PathInfo.ValueString(),
		KeepConn:    m.KeepConn.ValueBool(),
		MpxsConns:   m.MpxsConns.ValueBool(),
		MaxReqs:     int(m.MaxReqs.ValueInt64()),
		Params:      fcgiParamsToAPI(ctx, m.Params),
	}
}

// fromAPI populates the Terraform model from the API model.
func (m *fcgiAppModel) fromAPI(ctx context.Context, a *client.FcgiApp) {
	m.ID = types.Int64Value(int64(a.ID))
	m.Name = types.StringValue(a.Name)
	m.Description = types.StringValue(a.Description)
	m.Docroot = types.StringValue(a.Docroot)
	m.Index = types.StringValue(a.Index)
	m.PathInfo = types.StringValue(a.PathInfo)
	m.KeepConn = types.BoolValue(a.KeepConn)
	m.MpxsConns = types.BoolValue(a.MpxsConns)
	m.MaxReqs = types.Int64Value(int64(a.MaxReqs))
	m.Params = fcgiParamsFromAPI(ctx, a.Params)
}

// fcgiParamsToAPI converts a types.List of nested param objects to []client.FcgiParam.
func fcgiParamsToAPI(ctx context.Context, l types.List) []client.FcgiParam {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	params := make([]client.FcgiParam, 0, len(l.Elements()))
	for _, v := range l.Elements() {
		obj, ok := v.(types.Object)
		if !ok {
			continue
		}
		var pm fcgiParamModel
		diags := pm.fromObject(ctx, obj)
		_ = diags
		params = append(params, client.FcgiParam{
			Name:  pm.Name.ValueString(),
			Value: pm.Value.ValueString(),
		})
	}
	return params
}

// fcgiParamsFromAPI converts []client.FcgiParam to a types.List of nested param objects.
func fcgiParamsFromAPI(ctx context.Context, params []client.FcgiParam) types.List {
	if params == nil {
		return types.ListNull(paramsObjectType)
	}
	elems := make([]attr.Value, 0, len(params))
	for _, p := range params {
		objVal, diags := types.ObjectValue(
			map[string]attr.Type{
				"name":  types.StringType,
				"value": types.StringType,
			},
			map[string]attr.Value{
				"name":  types.StringValue(p.Name),
				"value": types.StringValue(p.Value),
			},
		)
		_ = diags
		elems = append(elems, objVal)
	}
	l, diags := types.ListValue(paramsObjectType, elems)
	_ = diags
	return l
}

// paramsObjectType is the types.Object type for a FastCGI param.
var paramsObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name":  types.StringType,
		"value": types.StringType,
	},
}

// fromObject populates the param model from a types.Object.
func (m *fcgiParamModel) fromObject(ctx context.Context, obj types.Object) diag.Diagnostics {
	var diags diag.Diagnostics
	// Extract attributes from the object.
	attrs := obj.Attributes()
	if v, ok := attrs["name"].(types.String); ok {
		m.Name = v
	}
	if v, ok := attrs["value"].(types.String); ok {
		m.Value = v
	}
	return diags
}

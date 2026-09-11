package resources

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

// Ensure the implementation satisfies the resource.Resource interface.
var _ resource.Resource = &ResponseTransformResource{}
var _ resource.ResourceWithImportState = &ResponseTransformResource{}

// ResponseTransformResource defines the corex_response_transform resource.
type ResponseTransformResource struct {
	cli *client.Client
}

func NewResponseTransformResource() resource.Resource {
	return &ResponseTransformResource{}
}

// responseTransformModel maps the Terraform schema to the API model.
type responseTransformModel struct {
	ID              types.Int64  `tfsdk:"id"`
	BackendID       types.Int64  `tfsdk:"backend_id"`
	BackendIDs      types.List   `tfsdk:"backend_ids"`
	Priority        types.Int64  `tfsdk:"priority"`
	Name            types.String `tfsdk:"name"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	TransformType   types.String `tfsdk:"transform_type"`
	FindRegex       types.String `tfsdk:"find_regex"`
	ReplaceString   types.String `tfsdk:"replace_string"`
	InjectHeader    types.String `tfsdk:"inject_header"`
	InjectValue     types.String `tfsdk:"inject_value"`
	MaskPattern     types.String `tfsdk:"mask_pattern"`
	MaskReplacement types.String `tfsdk:"mask_replacement"`
}

func (r *ResponseTransformResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_response_transform"
}

func (r *ResponseTransformResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a response transformation rule in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":               idAttr(),
			"backend_id":       intAttr("Backend ID the transform is attached to.", false),
			"backend_ids":      listIntAttr("Backend IDs the transform applies to.", false),
			"priority":         intAttr("Transform priority.", false),
			"name":             stringAttr("Transform name.", true),
			"enabled":          boolAttr("Whether the transform is enabled.", false),
			"transform_type":   stringAttr("Transform type (find_replace/inject/mask).", true),
			"find_regex":       stringAttr("Regex to find.", false),
			"replace_string":   stringAttr("Replacement string.", false),
			"inject_header":    stringAttr("Header to inject.", false),
			"inject_value":     stringAttr("Header value to inject.", false),
			"mask_pattern":     stringAttr("Pattern to mask.", false),
			"mask_replacement": stringAttr("Mask replacement string.", false),
		},
	}
}

func (r *ResponseTransformResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *ResponseTransformResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan responseTransformModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	t := plan.toAPI(ctx)
	result, err := r.cli.CreateResponseTransform(ctx, t)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create response transform", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ResponseTransformResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state responseTransformModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	t, err := r.cli.GetResponseTransform(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read response transform", err.Error())
		return
	}

	state.fromAPI(ctx, t)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ResponseTransformResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan responseTransformModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	t := plan.toAPI(ctx)
	result, err := r.cli.UpdateResponseTransform(ctx, int(plan.ID.ValueInt64()), t)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update response transform", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ResponseTransformResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state responseTransformModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteResponseTransform(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete response transform", err.Error())
		return
	}
}

func (r *ResponseTransformResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by numeric ID only.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "Response transform must be imported by numeric ID")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *responseTransformModel) toAPI(ctx context.Context) *client.ResponseTransform {
	return &client.ResponseTransform{
		BackendID:       int(m.BackendID.ValueInt64()),
		BackendIDs:      intListToSlice(ctx, m.BackendIDs),
		Priority:        int(m.Priority.ValueInt64()),
		Name:            m.Name.ValueString(),
		Enabled:         m.Enabled.ValueBool(),
		TransformType:   m.TransformType.ValueString(),
		FindRegex:       m.FindRegex.ValueString(),
		ReplaceString:   m.ReplaceString.ValueString(),
		InjectHeader:    m.InjectHeader.ValueString(),
		InjectValue:     m.InjectValue.ValueString(),
		MaskPattern:     m.MaskPattern.ValueString(),
		MaskReplacement: m.MaskReplacement.ValueString(),
	}
}

// fromAPI populates the Terraform model from the API model.
func (m *responseTransformModel) fromAPI(ctx context.Context, t *client.ResponseTransform) {
	m.ID = types.Int64Value(int64(t.ID))
	m.BackendID = types.Int64Value(int64(t.BackendID))
	m.BackendIDs = sliceToIntList(ctx, t.BackendIDs)
	m.Priority = types.Int64Value(int64(t.Priority))
	m.Name = types.StringValue(t.Name)
	m.Enabled = types.BoolValue(t.Enabled)
	m.TransformType = types.StringValue(t.TransformType)
	m.FindRegex = types.StringValue(t.FindRegex)
	m.ReplaceString = types.StringValue(t.ReplaceString)
	m.InjectHeader = types.StringValue(t.InjectHeader)
	m.InjectValue = types.StringValue(t.InjectValue)
	m.MaskPattern = types.StringValue(t.MaskPattern)
	m.MaskReplacement = types.StringValue(t.MaskReplacement)
}

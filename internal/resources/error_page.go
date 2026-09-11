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
var _ resource.Resource = &ErrorPageResource{}
var _ resource.ResourceWithImportState = &ErrorPageResource{}

// ErrorPageResource defines the corex_error_page resource.
type ErrorPageResource struct {
	cli *client.Client
}

func NewErrorPageResource() resource.Resource {
	return &ErrorPageResource{}
}

// errorPageModel maps the Terraform schema to the API model.
type errorPageModel struct {
	ID          types.Int64  `tfsdk:"id"`
	ListenerID  types.Int64  `tfsdk:"listener_id"`
	ListenerIDs types.List   `tfsdk:"listener_ids"`
	Code        types.Int64  `tfsdk:"code"`
	ContentType types.String `tfsdk:"content_type"`
	Content     types.String `tfsdk:"content"`
}

func (r *ErrorPageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_error_page"
}

func (r *ErrorPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom error page in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":           idAttr(),
			"listener_id":  intAttr("Listener ID the error page is attached to.", false),
			"listener_ids": listIntAttr("Listener IDs the error page applies to.", false),
			"code":         intAttr("HTTP status code for the error page.", true),
			"content_type": stringAttr("Content-Type header value.", false),
			"content":      stringAttr("Error page content.", true),
		},
	}
}

func (r *ErrorPageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *ErrorPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan errorPageModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := plan.toAPI(ctx)
	result, err := r.cli.CreateErrorPage(ctx, p)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create error page", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ErrorPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state errorPageModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := r.cli.GetErrorPage(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read error page", err.Error())
		return
	}

	state.fromAPI(ctx, p)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ErrorPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan errorPageModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := plan.toAPI(ctx)
	result, err := r.cli.UpdateErrorPage(ctx, int(plan.ID.ValueInt64()), p)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update error page", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ErrorPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state errorPageModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteErrorPage(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete error page", err.Error())
		return
	}
}

func (r *ErrorPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by numeric ID only.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "Error page must be imported by numeric ID")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *errorPageModel) toAPI(ctx context.Context) *client.ErrorPage {
	return &client.ErrorPage{
		ListenerID:  int(m.ListenerID.ValueInt64()),
		ListenerIDs: intListToSlice(ctx, m.ListenerIDs),
		Code:        int(m.Code.ValueInt64()),
		ContentType: m.ContentType.ValueString(),
		Content:     m.Content.ValueString(),
	}
}

// fromAPI populates the Terraform model from the API model.
func (m *errorPageModel) fromAPI(ctx context.Context, p *client.ErrorPage) {
	m.ID = types.Int64Value(int64(p.ID))
	m.ListenerID = types.Int64Value(int64(p.ListenerID))
	m.ListenerIDs = sliceToIntList(ctx, p.ListenerIDs)
	m.Code = types.Int64Value(int64(p.Code))
	m.ContentType = types.StringValue(p.ContentType)
	m.Content = types.StringValue(p.Content)
}

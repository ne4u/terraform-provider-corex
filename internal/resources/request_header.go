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
var _ resource.Resource = &RequestHeaderResource{}
var _ resource.ResourceWithImportState = &RequestHeaderResource{}

// RequestHeaderResource defines the corex_request_header resource.
type RequestHeaderResource struct {
	cli *client.Client
}

func NewRequestHeaderResource() resource.Resource {
	return &RequestHeaderResource{}
}

// requestHeaderModel maps the Terraform schema to the API model.
type requestHeaderModel struct {
	ID         types.Int64  `tfsdk:"id"`
	BackendID  types.Int64  `tfsdk:"backend_id"`
	BackendIDs types.List   `tfsdk:"backend_ids"`
	Header     types.String `tfsdk:"header"`
	Value      types.String `tfsdk:"value"`
	Action     types.String `tfsdk:"action"`
	Condition  types.String `tfsdk:"condition"`
}

func (r *RequestHeaderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request_header"
}

func (r *RequestHeaderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a request header manipulation rule in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":          idAttr(),
			"backend_id":  intAttr("Backend ID the rule is attached to.", false),
			"backend_ids": listIntAttr("Backend IDs the rule applies to.", false),
			"header":      stringAttr("Header name.", true),
			"value":       stringAttr("Header value.", false),
			"action":      stringAttr("Action (set/add/del/etc.).", true),
			"condition":   stringAttr("Condition expression.", false),
		},
	}
}

func (r *RequestHeaderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *RequestHeaderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan requestHeaderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	h := plan.toAPI(ctx)
	result, err := r.cli.CreateRequestHeader(ctx, h)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create request header", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RequestHeaderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state requestHeaderModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	h, err := r.cli.GetRequestHeader(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read request header", err.Error())
		return
	}

	state.fromAPI(ctx, h)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *RequestHeaderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan requestHeaderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	h := plan.toAPI(ctx)
	result, err := r.cli.UpdateRequestHeader(ctx, int(plan.ID.ValueInt64()), h)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update request header", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RequestHeaderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state requestHeaderModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteRequestHeader(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete request header", err.Error())
		return
	}
}

func (r *RequestHeaderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by numeric ID only.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "Request header must be imported by numeric ID")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *requestHeaderModel) toAPI(ctx context.Context) *client.RequestHeader {
	return &client.RequestHeader{
		BackendID:  int(m.BackendID.ValueInt64()),
		BackendIDs: intListToSlice(ctx, m.BackendIDs),
		Header:     m.Header.ValueString(),
		Value:      m.Value.ValueString(),
		Action:     m.Action.ValueString(),
		Condition:  m.Condition.ValueString(),
	}
}

// fromAPI populates the Terraform model from the API model.
func (m *requestHeaderModel) fromAPI(ctx context.Context, h *client.RequestHeader) {
	m.ID = types.Int64Value(int64(h.ID))
	m.BackendID = types.Int64Value(int64(h.BackendID))
	m.BackendIDs = sliceToIntList(ctx, h.BackendIDs)
	m.Header = types.StringValue(h.Header)
	m.Value = types.StringValue(h.Value)
	m.Action = types.StringValue(h.Action)
	m.Condition = types.StringValue(h.Condition)
}

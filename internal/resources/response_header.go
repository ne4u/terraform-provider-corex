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
var _ resource.Resource = &ResponseHeaderResource{}
var _ resource.ResourceWithImportState = &ResponseHeaderResource{}

// ResponseHeaderResource defines the corex_response_header resource.
type ResponseHeaderResource struct {
	cli *client.Client
}

func NewResponseHeaderResource() resource.Resource {
	return &ResponseHeaderResource{}
}

// responseHeaderModel maps the Terraform schema to the API model.
type responseHeaderModel struct {
	ID          types.Int64  `tfsdk:"id"`
	ListenerID  types.Int64  `tfsdk:"listener_id"`
	ListenerIDs types.List   `tfsdk:"listener_ids"`
	Header      types.String `tfsdk:"header"`
	Value       types.String `tfsdk:"value"`
	Action      types.String `tfsdk:"action"`
	Condition   types.String `tfsdk:"condition"`
}

func (r *ResponseHeaderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_response_header"
}

func (r *ResponseHeaderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a response header manipulation rule in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":           idAttr(),
			"listener_id":  intAttr("Listener ID the rule is attached to.", false),
			"listener_ids": listIntAttr("Listener IDs the rule applies to.", false),
			"header":       stringAttr("Header name.", true),
			"value":        stringAttr("Header value.", false),
			"action":       stringAttr("Action (set/add/del/etc.).", true),
			"condition":    stringAttr("Condition expression.", false),
		},
	}
}

func (r *ResponseHeaderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *ResponseHeaderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan responseHeaderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	h := plan.toAPI(ctx)
	result, err := r.cli.CreateResponseHeader(ctx, h)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create response header", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ResponseHeaderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state responseHeaderModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	h, err := r.cli.GetResponseHeader(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read response header", err.Error())
		return
	}

	state.fromAPI(ctx, h)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ResponseHeaderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan responseHeaderModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	h := plan.toAPI(ctx)
	result, err := r.cli.UpdateResponseHeader(ctx, int(plan.ID.ValueInt64()), h)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update response header", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ResponseHeaderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state responseHeaderModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteResponseHeader(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete response header", err.Error())
		return
	}
}

func (r *ResponseHeaderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by numeric ID only.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", "Response header must be imported by numeric ID")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *responseHeaderModel) toAPI(ctx context.Context) *client.ResponseHeader {
	return &client.ResponseHeader{
		ListenerID:  int(m.ListenerID.ValueInt64()),
		ListenerIDs: intListToSlice(ctx, m.ListenerIDs),
		Header:      m.Header.ValueString(),
		Value:       m.Value.ValueString(),
		Action:      m.Action.ValueString(),
		Condition:   m.Condition.ValueString(),
	}
}

// fromAPI populates the Terraform model from the API model.
func (m *responseHeaderModel) fromAPI(ctx context.Context, h *client.ResponseHeader) {
	m.ID = types.Int64Value(int64(h.ID))
	m.ListenerID = types.Int64Value(int64(h.ListenerID))
	m.ListenerIDs = sliceToIntList(ctx, h.ListenerIDs)
	m.Header = types.StringValue(h.Header)
	m.Value = types.StringValue(h.Value)
	m.Action = types.StringValue(h.Action)
	m.Condition = types.StringValue(h.Condition)
}

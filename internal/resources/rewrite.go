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
var _ resource.Resource = &RewriteResource{}
var _ resource.ResourceWithImportState = &RewriteResource{}

// RewriteResource defines the corex_rewrite resource.
type RewriteResource struct {
	cli *client.Client
}

func NewRewriteResource() resource.Resource {
	return &RewriteResource{}
}

// rewriteModel maps the Terraform schema to the API model.
type rewriteModel struct {
	ID          types.Int64  `tfsdk:"id"`
	ListenerID  types.Int64  `tfsdk:"listener_id"`
	ListenerIDs types.List   `tfsdk:"listener_ids"`
	Priority    types.Int64  `tfsdk:"priority"`
	Name        types.String `tfsdk:"name"`
	HostMatch   types.String `tfsdk:"host_match"`
	SourceRegex types.String `tfsdk:"source_regex"`
	Target      types.String `tfsdk:"target"`
	Type        types.String `tfsdk:"type"`
}

func (r *RewriteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rewrite"
}

func (r *RewriteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a URL rewrite rule in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":           idAttr(),
			"listener_id":  intAttr("Listener ID the rewrite is attached to.", false),
			"listener_ids": listIntAttr("Listener IDs the rewrite applies to.", false),
			"priority":     intAttr("Rewrite priority.", false),
			"name":         stringAttr("Rewrite name.", true),
			"host_match":   stringAttr("Host match pattern.", false),
			"source_regex": stringAttr("Source regex pattern.", false),
			"target":       stringAttr("Replacement target.", false),
			"type":         stringAttr("Rewrite type (path/query/host).", false),
		},
	}
}

func (r *RewriteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *RewriteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan rewriteModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rw := plan.toAPI(ctx)
	result, err := r.cli.CreateRewrite(ctx, rw)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create rewrite", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RewriteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state rewriteModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rw, err := r.cli.GetRewrite(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read rewrite", err.Error())
		return
	}

	state.fromAPI(ctx, rw)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *RewriteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan rewriteModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rw := plan.toAPI(ctx)
	result, err := r.cli.UpdateRewrite(ctx, int(plan.ID.ValueInt64()), rw)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update rewrite", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RewriteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state rewriteModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteRewrite(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete rewrite", err.Error())
		return
	}
}

func (r *RewriteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to import by name: list all rewrites, find matching name.
	rewrites, err := r.cli.ListRewrites(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list rewrites for import", err.Error())
		return
	}

	for _, rw := range rewrites {
		if rw.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(rw.ID)))...)
			return
		}
	}

	// Try numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Rewrite not found", fmt.Sprintf("No rewrite with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *rewriteModel) toAPI(ctx context.Context) *client.Rewrite {
	return &client.Rewrite{
		ListenerID:  int(m.ListenerID.ValueInt64()),
		ListenerIDs: intListToSlice(ctx, m.ListenerIDs),
		Priority:    int(m.Priority.ValueInt64()),
		Name:        m.Name.ValueString(),
		HostMatch:   m.HostMatch.ValueString(),
		SourceRegex: m.SourceRegex.ValueString(),
		Target:      m.Target.ValueString(),
		Type:        m.Type.ValueString(),
	}
}

// fromAPI populates the Terraform model from the API model.
func (m *rewriteModel) fromAPI(ctx context.Context, rw *client.Rewrite) {
	m.ID = types.Int64Value(int64(rw.ID))
	m.ListenerID = types.Int64Value(int64(rw.ListenerID))
	m.ListenerIDs = sliceToIntList(ctx, rw.ListenerIDs)
	m.Priority = types.Int64Value(int64(rw.Priority))
	m.Name = types.StringValue(rw.Name)
	m.HostMatch = types.StringValue(rw.HostMatch)
	m.SourceRegex = types.StringValue(rw.SourceRegex)
	m.Target = types.StringValue(rw.Target)
	m.Type = types.StringValue(rw.Type)
}

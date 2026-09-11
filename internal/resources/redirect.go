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
var _ resource.Resource = &RedirectResource{}
var _ resource.ResourceWithImportState = &RedirectResource{}

// RedirectResource defines the corex_redirect resource.
type RedirectResource struct {
	cli *client.Client
}

func NewRedirectResource() resource.Resource {
	return &RedirectResource{}
}

// redirectModel maps the Terraform schema to the API model.
type redirectModel struct {
	ID            types.Int64  `tfsdk:"id"`
	ListenerID    types.Int64  `tfsdk:"listener_id"`
	ListenerIDs   types.List   `tfsdk:"listener_ids"`
	Priority      types.Int64  `tfsdk:"priority"`
	Name          types.String `tfsdk:"name"`
	Source        types.String `tfsdk:"source"`
	Target        types.String `tfsdk:"target"`
	Type          types.String `tfsdk:"type"`
	Code          types.Int64  `tfsdk:"code"`
	PreserveQuery types.Bool   `tfsdk:"preserve_query"`
}

func (r *RedirectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redirect"
}

func (r *RedirectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a redirect rule in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":             idAttr(),
			"listener_id":    intAttr("Listener ID the redirect is attached to.", false),
			"listener_ids":   listIntAttr("Listener IDs the redirect applies to.", false),
			"priority":       intAttr("Redirect priority.", false),
			"name":           stringAttr("Redirect name.", true),
			"source":         stringAttr("Source pattern.", false),
			"target":         stringAttr("Target URL.", false),
			"type":           stringAttr("Redirect type (exact/prefix/regex).", false),
			"code":           intAttr("HTTP redirect status code.", false),
			"preserve_query": boolAttr("Preserve query string.", false),
		},
	}
}

func (r *RedirectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *RedirectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan redirectModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rd := plan.toAPI(ctx)
	result, err := r.cli.CreateRedirect(ctx, rd)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create redirect", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RedirectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state redirectModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rd, err := r.cli.GetRedirect(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read redirect", err.Error())
		return
	}

	state.fromAPI(ctx, rd)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *RedirectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan redirectModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rd := plan.toAPI(ctx)
	result, err := r.cli.UpdateRedirect(ctx, int(plan.ID.ValueInt64()), rd)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update redirect", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RedirectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state redirectModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteRedirect(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete redirect", err.Error())
		return
	}
}

func (r *RedirectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to import by name: list all redirects, find matching name.
	redirects, err := r.cli.ListRedirects(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list redirects for import", err.Error())
		return
	}

	for _, rd := range redirects {
		if rd.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(rd.ID)))...)
			return
		}
	}

	// Try numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Redirect not found", fmt.Sprintf("No redirect with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *redirectModel) toAPI(ctx context.Context) *client.Redirect {
	return &client.Redirect{
		ListenerID:    int(m.ListenerID.ValueInt64()),
		ListenerIDs:   intListToSlice(ctx, m.ListenerIDs),
		Priority:      int(m.Priority.ValueInt64()),
		Name:          m.Name.ValueString(),
		Source:        m.Source.ValueString(),
		Target:        m.Target.ValueString(),
		Type:          m.Type.ValueString(),
		Code:          int(m.Code.ValueInt64()),
		PreserveQuery: m.PreserveQuery.ValueBool(),
	}
}

// fromAPI populates the Terraform model from the API model.
func (m *redirectModel) fromAPI(ctx context.Context, rd *client.Redirect) {
	m.ID = types.Int64Value(int64(rd.ID))
	m.ListenerID = types.Int64Value(int64(rd.ListenerID))
	m.ListenerIDs = sliceToIntList(ctx, rd.ListenerIDs)
	m.Priority = types.Int64Value(int64(rd.Priority))
	m.Name = types.StringValue(rd.Name)
	m.Source = types.StringValue(rd.Source)
	m.Target = types.StringValue(rd.Target)
	m.Type = types.StringValue(rd.Type)
	m.Code = types.Int64Value(int64(rd.Code))
	m.PreserveQuery = types.BoolValue(rd.PreserveQuery)
}

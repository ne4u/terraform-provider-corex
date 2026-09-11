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

var _ resource.Resource = &PageProtectScriptResource{}
var _ resource.ResourceWithImportState = &PageProtectScriptResource{}

// PageProtectScriptResource defines the corex_page_protect_script resource.
type PageProtectScriptResource struct {
	cli *client.Client
}

func NewPageProtectScriptResource() resource.Resource {
	return &PageProtectScriptResource{}
}

type pageProtectScriptModel struct {
	ID           types.Int64  `tfsdk:"id"`
	URL          types.String `tfsdk:"url"`
	ResourceType types.String `tfsdk:"resource_type"`
	Notes        types.String `tfsdk:"notes"`
	FetchMethod  types.String `tfsdk:"fetch_method"`
	Ignored      types.Bool   `tfsdk:"ignored"`
}

func (r *PageProtectScriptResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_page_protect_script"
}

func (r *PageProtectScriptResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a page protection script resource in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":             idAttr(),
			"url":            stringAttr("Script URL.", true),
			"resource_type":  stringAttr("Resource type.", false),
			"notes":          stringAttr("Notes.", false),
			"fetch_method":   stringAttr("Fetch method.", false),
			"ignored":        boolAttr("Whether this script is ignored.", false),
		},
	}
}

func (r *PageProtectScriptResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *PageProtectScriptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pageProtectScriptModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI()
	result, err := r.cli.CreatePageProtectScript(ctx, s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create page protect script", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PageProtectScriptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pageProtectScriptModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.cli.GetPageProtectScript(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read page protect script", err.Error())
		return
	}

	state.fromAPI(s)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *PageProtectScriptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pageProtectScriptModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI()
	result, err := r.cli.UpdatePageProtectScript(ctx, int(plan.ID.ValueInt64()), s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update page protect script", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PageProtectScriptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pageProtectScriptModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeletePageProtectScript(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete page protect script", err.Error())
		return
	}
}

func (r *PageProtectScriptResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	scripts, err := r.cli.ListPageProtectScripts(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list page protect scripts for import", err.Error())
		return
	}

	for _, s := range scripts {
		if s.URL == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(s.ID)))...)
			return
		}
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Page protect script not found", fmt.Sprintf("No script with URL or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *pageProtectScriptModel) toAPI() *client.PageProtectScript {
	return &client.PageProtectScript{
		URL:          m.URL.ValueString(),
		ResourceType: m.ResourceType.ValueString(),
		Notes:        m.Notes.ValueString(),
		FetchMethod:  m.FetchMethod.ValueString(),
		Ignored:      m.Ignored.ValueBool(),
	}
}

func (m *pageProtectScriptModel) fromAPI(s *client.PageProtectScript) {
	m.ID = types.Int64Value(int64(s.ID))
	m.URL = types.StringValue(s.URL)
	m.ResourceType = types.StringValue(s.ResourceType)
	m.Notes = types.StringValue(s.Notes)
	m.FetchMethod = types.StringValue(s.FetchMethod)
	m.Ignored = types.BoolValue(s.Ignored)
}

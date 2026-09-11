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

var _ resource.Resource = &McpTeamResource{}
var _ resource.ResourceWithImportState = &McpTeamResource{}

type McpTeamResource struct {
	cli *client.Client
}

func NewMcpTeamResource() resource.Resource {
	return &McpTeamResource{}
}

type mcpTeamModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Description types.String `tfsdk:"description"`
}

func (r *McpTeamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_team"
}

func (r *McpTeamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an MCP gateway team in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":          idAttr(),
			"name":        stringAttr("Team name.", true),
			"slug":        stringAttr("Team slug (lowercase, hyphens only).", true),
			"description": stringAttr("Team description.", false),
		},
	}
}

func (r *McpTeamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *McpTeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mcpTeamModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	t := plan.toAPI()
	result, err := r.cli.CreateMcpTeam(ctx, t)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create MCP team", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpTeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mcpTeamModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	t, err := r.cli.GetMcpTeam(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read MCP team", err.Error())
		return
	}

	state.fromAPI(t)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *McpTeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan mcpTeamModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	t := plan.toAPI()
	result, err := r.cli.UpdateMcpTeam(ctx, int(plan.ID.ValueInt64()), t)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update MCP team", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpTeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mcpTeamModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteMcpTeam(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete MCP team", err.Error())
		return
	}
}

func (r *McpTeamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	teams, err := r.cli.ListMcpTeams(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list MCP teams for import", err.Error())
		return
	}

	for _, t := range teams {
		if t.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(t.ID)))...)
			return
		}
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("MCP team not found", fmt.Sprintf("No MCP team with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *mcpTeamModel) toAPI() *client.McpTeam {
	return &client.McpTeam{
		Name:        m.Name.ValueString(),
		Slug:        m.Slug.ValueString(),
		Description: m.Description.ValueString(),
	}
}

func (m *mcpTeamModel) fromAPI(t *client.McpTeam) {
	m.ID = types.Int64Value(int64(t.ID))
	m.Name = types.StringValue(t.Name)
	m.Slug = types.StringValue(t.Slug)
	m.Description = types.StringValue(t.Description)
}

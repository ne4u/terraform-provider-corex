package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

var _ resource.Resource = &McpTeamMemberResource{}
var _ resource.ResourceWithImportState = &McpTeamMemberResource{}

type McpTeamMemberResource struct {
	cli *client.Client
}

func NewMcpTeamMemberResource() resource.Resource {
	return &McpTeamMemberResource{}
}

type mcpTeamMemberModel struct {
	ID     types.String `tfsdk:"id"`
	TeamID types.Int64  `tfsdk:"team_id"`
	UserID types.Int64  `tfsdk:"user_id"`
}

func (r *McpTeamMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_team_member"
}

func (r *McpTeamMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	teamIDAttr := intAttr("Team ID.", true)
	teamIDAttr.PlanModifiers = []planmodifier.Int64{int64planmodifier.RequiresReplace()}
	userIDAttr := intAttr("User ID to add to the team.", true)
	userIDAttr.PlanModifiers = []planmodifier.Int64{int64planmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Description: "Manages an MCP team membership in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID in the format {team_id}:{user_id}.",
				Computed:    true,
			},
			"team_id": teamIDAttr,
			"user_id": userIDAttr,
		},
	}
}

func (r *McpTeamMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *McpTeamMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mcpTeamMemberModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := int(plan.TeamID.ValueInt64())
	userID := int(plan.UserID.ValueInt64())

	_, err := r.cli.AddMcpTeamMember(ctx, teamID, userID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to add MCP team member", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d:%d", teamID, userID))
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpTeamMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mcpTeamMemberModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := int(state.TeamID.ValueInt64())
	userID := int(state.UserID.ValueInt64())

	members, err := r.cli.ListMcpTeamMembers(ctx, teamID)
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to list MCP team members", err.Error())
		return
	}

	found := false
	for _, m := range members {
		if m.UserID == userID {
			found = true
			break
		}
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(fmt.Sprintf("%d:%d", teamID, userID))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *McpTeamMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update — team_id and user_id are ForceNew.
}

func (r *McpTeamMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mcpTeamMemberModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	teamID := int(state.TeamID.ValueInt64())
	userID := int(state.UserID.ValueInt64())

	err := r.cli.RemoveMcpTeamMember(ctx, teamID, userID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to remove MCP team member", err.Error())
		return
	}
}

func (r *McpTeamMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format {team_id}:{user_id}")
		return
	}
	teamID, err := strconv.Atoi(parts[0])
	if err != nil {
		resp.Diagnostics.AddError("Invalid team_id in import ID", err.Error())
		return
	}
	userID, err := strconv.Atoi(parts[1])
	if err != nil {
		resp.Diagnostics.AddError("Invalid user_id in import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("team_id"), types.Int64Value(int64(teamID)))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), types.Int64Value(int64(userID)))...)
}

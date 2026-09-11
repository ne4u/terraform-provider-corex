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

var _ resource.Resource = &McpSkillResource{}
var _ resource.ResourceWithImportState = &McpSkillResource{}

type McpSkillResource struct {
	cli *client.Client
}

func NewMcpSkillResource() resource.Resource {
	return &McpSkillResource{}
}

type mcpSkillModel struct {
	ID                 types.Int64  `tfsdk:"id"`
	TeamID             types.Int64  `tfsdk:"team_id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Enabled            types.Bool   `tfsdk:"enabled"`
	EnableWhen         types.String `tfsdk:"enable_when"`
	EnableWhenAST      types.String `tfsdk:"enable_when_ast"`
	Tags               types.List   `tfsdk:"tags"`
	PublishedVersionID types.Int64  `tfsdk:"published_version_id"`
}

func (r *McpSkillResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_skill"
}

func (r *McpSkillResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an MCP gateway skill in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":                   idAttr(),
			"team_id":              intAttr("Team ID that owns this skill.", true),
			"name":                 stringAttr("Skill name (lowercase, hyphens only).", true),
			"description":          stringAttr("Skill description.", false),
			"enabled":              boolAttr("Whether the skill is enabled.", false),
			"enable_when":          stringAttr("Condition expression for when to enable the skill.", false),
			"enable_when_ast":      stringAttrComputed("JSON-encoded AST of the enable_when expression."),
			"tags":                 listAttr("Skill tags.", false),
			"published_version_id": intAttrComputed("ID of the published skill version."),
		},
	}
}

func (r *McpSkillResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *McpSkillResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mcpSkillModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI(ctx)
	result, err := r.cli.CreateMcpSkill(ctx, s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create MCP skill", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpSkillResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mcpSkillModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.cli.GetMcpSkill(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read MCP skill", err.Error())
		return
	}

	state.fromAPI(ctx, s)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *McpSkillResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan mcpSkillModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI(ctx)
	result, err := r.cli.UpdateMcpSkill(ctx, int(plan.ID.ValueInt64()), s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update MCP skill", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpSkillResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mcpSkillModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteMcpSkill(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete MCP skill", err.Error())
		return
	}
}

func (r *McpSkillResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	skills, err := r.cli.ListMcpSkills(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list MCP skills for import", err.Error())
		return
	}

	for _, s := range skills {
		if s.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(s.ID)))...)
			return
		}
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("MCP skill not found", fmt.Sprintf("No MCP skill with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *mcpSkillModel) toAPI(ctx context.Context) *client.McpSkill {
	s := &client.McpSkill{
		TeamID: int(m.TeamID.ValueInt64()),
		Name:   m.Name.ValueString(),
		Enabled: m.Enabled.ValueBool(),
	}
	if !m.Description.IsNull() {
		s.Description = stringPtr(m.Description.ValueString())
	}
	if !m.EnableWhen.IsNull() {
		s.EnableWhen = stringPtr(m.EnableWhen.ValueString())
	}
	if !m.Tags.IsNull() {
		s.Tags = stringListToSlice(ctx, m.Tags)
	}
	return s
}

func (m *mcpSkillModel) fromAPI(ctx context.Context, s *client.McpSkill) {
	m.ID = types.Int64Value(int64(s.ID))
	m.TeamID = types.Int64Value(int64(s.TeamID))
	m.Name = types.StringValue(s.Name)
	m.Description = types.StringValue(emptyIfNil(s.Description))
	m.Enabled = types.BoolValue(s.Enabled)
	m.EnableWhen = types.StringValue(emptyIfNil(s.EnableWhen))
	m.EnableWhenAST = types.StringValue(goMapToJSON(s.EnableWhenAST))
	m.Tags = sliceToStringList(ctx, s.Tags)
	if s.PublishedVersionID != nil {
		m.PublishedVersionID = types.Int64Value(int64(*s.PublishedVersionID))
	} else {
		m.PublishedVersionID = types.Int64Null()
	}
}

package resources

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

var _ resource.Resource = &McpSkillVersionResource{}
var _ resource.ResourceWithImportState = &McpSkillVersionResource{}

type McpSkillVersionResource struct {
	cli *client.Client
}

func NewMcpSkillVersionResource() resource.Resource {
	return &McpSkillVersionResource{}
}

type mcpSkillVersionModel struct {
	ID          types.Int64  `tfsdk:"id"`
	SkillID     types.Int64  `tfsdk:"skill_id"`
	Version     types.Int64  `tfsdk:"version"`
	Frontmatter types.String `tfsdk:"frontmatter"`
	Body        types.String `tfsdk:"body"`
	Files       types.List   `tfsdk:"files"`
}

type mcpSkillFileModel struct {
	Path       types.String `tfsdk:"path"`
	MediaType  types.String `tfsdk:"media_type"`
	ContentB64 types.String `tfsdk:"content_b64"`
}

func (r *McpSkillVersionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_skill_version"
}

func (r *McpSkillVersionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an MCP skill version in coreX Manager. Skill versions are immutable.",
		Attributes: map[string]schema.Attribute{
			"id":          idAttr(),
			"skill_id":    intAttr("Parent skill ID.", true),
			"version":     intAttrComputed("Version number (assigned by server)."),
			"frontmatter": stringAttr("JSON-encoded frontmatter metadata.", false),
			"body":        stringAttr("Skill body content (Markdown).", true),
			"files": schema.ListNestedAttribute{
				Description: "Files attached to the skill version.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path":        stringAttr("File path.", false),
						"media_type":  stringAttr("File media type.", false),
						"content_b64": stringAttr("File content (base64-encoded).", false),
					},
				},
			},
		},
	}
}

func (r *McpSkillVersionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *McpSkillVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mcpSkillVersionModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	v := plan.toAPI(ctx)
	result, err := r.cli.CreateMcpSkillVersion(ctx, int(plan.SkillID.ValueInt64()), v)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create MCP skill version", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpSkillVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mcpSkillVersionModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	versions, err := r.cli.ListMcpSkillVersions(ctx, int(state.SkillID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to list MCP skill versions", err.Error())
		return
	}

	var found *client.McpSkillVersion
	for i := range versions {
		if versions[i].ID == int(state.ID.ValueInt64()) {
			found = &versions[i]
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.fromAPI(ctx, found)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update is a no-op — skill versions are immutable.
func (r *McpSkillVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete is a no-op — skill versions are immutable.
func (r *McpSkillVersionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *McpSkillVersionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected a numeric skill version ID")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *mcpSkillVersionModel) toAPI(ctx context.Context) *client.McpSkillVersion {
	v := &client.McpSkillVersion{
		Body: m.Body.ValueString(),
	}
	if !m.Frontmatter.IsNull() {
		v.Frontmatter = jsonMapToGo(m.Frontmatter.ValueString())
	}
	if !m.Files.IsNull() && !m.Files.IsUnknown() {
		var fileModels []mcpSkillFileModel
		diags := m.Files.ElementsAs(ctx, &fileModels, false)
		_ = diags
		for _, fm := range fileModels {
			v.Files = append(v.Files, client.McpSkillFile{
				Path:       fm.Path.ValueString(),
				MediaType:  fm.MediaType.ValueString(),
				ContentB64: fm.ContentB64.ValueString(),
			})
		}
	}
	return v
}

func (m *mcpSkillVersionModel) fromAPI(ctx context.Context, v *client.McpSkillVersion) {
	m.ID = types.Int64Value(int64(v.ID))
	m.SkillID = types.Int64Value(int64(v.SkillID))
	m.Version = types.Int64Value(int64(v.Version))
	m.Frontmatter = types.StringValue(goMapToJSON(v.Frontmatter))
	m.Body = types.StringValue(v.Body)

	if v.Files == nil {
		m.Files = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"path":        types.StringType,
				"media_type":  types.StringType,
				"content_b64": types.StringType,
			},
		})
	} else {
		elems := make([]attr.Value, 0, len(v.Files))
		objType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"path":        types.StringType,
				"media_type":  types.StringType,
				"content_b64": types.StringType,
			},
		}
		for _, f := range v.Files {
			obj, diags := types.ObjectValue(objType.AttrTypes, map[string]attr.Value{
				"path":        types.StringValue(f.Path),
				"media_type":  types.StringValue(f.MediaType),
				"content_b64": types.StringValue(f.ContentB64),
			})
			_ = diags
			elems = append(elems, obj)
		}
		l, diags := types.ListValue(objType, elems)
		_ = diags
		m.Files = l
	}
}

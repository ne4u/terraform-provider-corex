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

var _ resource.Resource = &McpServerReplicaResource{}
var _ resource.ResourceWithImportState = &McpServerReplicaResource{}

type McpServerReplicaResource struct {
	cli *client.Client
}

func NewMcpServerReplicaResource() resource.Resource {
	return &McpServerReplicaResource{}
}

type mcpServerReplicaModel struct {
	ID        types.Int64  `tfsdk:"id"`
	ServerID  types.Int64  `tfsdk:"server_id"`
	URL       types.String `tfsdk:"url"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	VerifyTLS types.Bool   `tfsdk:"verify_tls"`
}

func (r *McpServerReplicaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_server_replica"
}

func (r *McpServerReplicaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an MCP server replica URL in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":         idAttr(),
			"server_id":  intAttr("Parent MCP server ID.", true),
			"url":        stringAttr("Replica upstream URL.", true),
			"enabled":    boolAttr("Whether the replica is enabled.", false),
			"verify_tls": boolAttr("Verify upstream TLS certificate.", false),
		},
	}
}

func (r *McpServerReplicaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *McpServerReplicaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mcpServerReplicaModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rep := plan.toAPI()
	result, err := r.cli.CreateMcpServerReplica(ctx, int(plan.ServerID.ValueInt64()), rep)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create MCP server replica", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpServerReplicaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mcpServerReplicaModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	replicas, err := r.cli.ListMcpServerReplicas(ctx, int(state.ServerID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to list MCP server replicas", err.Error())
		return
	}

	var found *client.McpServerReplica
	for i := range replicas {
		if replicas[i].ID == int(state.ID.ValueInt64()) {
			found = &replicas[i]
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.fromAPI(found)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *McpServerReplicaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan mcpServerReplicaModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rep := plan.toAPI()
	result, err := r.cli.UpdateMcpServerReplica(ctx, int(plan.ServerID.ValueInt64()), int(plan.ID.ValueInt64()), rep)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update MCP server replica", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpServerReplicaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mcpServerReplicaModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteMcpServerReplica(ctx, int(state.ServerID.ValueInt64()), int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete MCP server replica", err.Error())
		return
	}
}

func (r *McpServerReplicaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected a numeric replica ID")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *mcpServerReplicaModel) toAPI() *client.McpServerReplica {
	return &client.McpServerReplica{
		URL:       m.URL.ValueString(),
		Enabled:   m.Enabled.ValueBool(),
		VerifyTLS: m.VerifyTLS.ValueBool(),
	}
}

func (m *mcpServerReplicaModel) fromAPI(r *client.McpServerReplica) {
	m.ID = types.Int64Value(int64(r.ID))
	m.ServerID = types.Int64Value(int64(r.ServerID))
	m.URL = types.StringValue(r.URL)
	m.Enabled = types.BoolValue(r.Enabled)
	m.VerifyTLS = types.BoolValue(r.VerifyTLS)
}

package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

const mcpAlertConfigID = "mcp-alerts"

var _ resource.Resource = &McpAlertConfigResource{}
var _ resource.ResourceWithImportState = &McpAlertConfigResource{}

type McpAlertConfigResource struct {
	cli *client.Client
}

func NewMcpAlertConfigResource() resource.Resource {
	return &McpAlertConfigResource{}
}

type mcpAlertConfigModel struct {
	ID          types.String `tfsdk:"id"`
	WebhookURL  types.String `tfsdk:"webhook_url"`
	Thresholds  types.Map    `tfsdk:"thresholds"`
}

func (r *McpAlertConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_alert_config"
}

func (r *McpAlertConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the MCP alert configuration (singleton) in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Singleton identifier, always \"mcp-alerts\".",
				Computed:    true,
			},
			"webhook_url": stringAttr("Webhook URL for alert notifications.", false),
			"thresholds":  mapIntAttr("Alert thresholds (event_type → count).", false),
		},
	}
}

func (r *McpAlertConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *McpAlertConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mcpAlertConfigModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := plan.toAPI(ctx)
	result, err := r.cli.UpdateMcpAlertConfig(ctx, cfg)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create MCP alert config", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	plan.ID = types.StringValue(mcpAlertConfigID)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *McpAlertConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mcpAlertConfigModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.cli.GetMcpAlertConfig(ctx)
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read MCP alert config", err.Error())
		return
	}

	state.fromAPI(ctx, cfg)
	state.ID = types.StringValue(mcpAlertConfigID)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *McpAlertConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan mcpAlertConfigModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := plan.toAPI(ctx)
	result, err := r.cli.UpdateMcpAlertConfig(ctx, cfg)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update MCP alert config", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	plan.ID = types.StringValue(mcpAlertConfigID)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

// Delete is a no-op — the alert config is a singleton and cannot be deleted.
func (r *McpAlertConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *McpAlertConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(mcpAlertConfigID))...)
}

func (m *mcpAlertConfigModel) toAPI(ctx context.Context) *client.McpAlertConfig {
	cfg := &client.McpAlertConfig{
		Thresholds: map[string]int{},
	}
	if !m.WebhookURL.IsNull() {
		cfg.WebhookURL = stringPtr(m.WebhookURL.ValueString())
	}
	if !m.Thresholds.IsNull() {
		cfg.Thresholds = mapIntToGo(ctx, m.Thresholds)
	}
	return cfg
}

func (m *mcpAlertConfigModel) fromAPI(ctx context.Context, cfg *client.McpAlertConfig) {
	m.WebhookURL = types.StringValue(emptyIfNil(cfg.WebhookURL))
	m.Thresholds = goMapIntToMap(ctx, cfg.Thresholds)
}

package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

// getClient extracts the client from the datasource Configure response.
func getClient(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *client.Client {
	if req.ProviderData == nil {
		return nil
	}
	cli, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return nil
	}
	return cli
}

// rawGet performs a GET request and returns the response marshalled to a JSON
// string. This is used for data sources that expose a single `xxx_json`
// attribute containing the raw API response.
func rawGet(ctx context.Context, cli *client.Client, path string) (string, error) {
	var raw interface{}
	if err := cli.Get(ctx, path, &raw); err != nil {
		return "", err
	}
	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}
	return string(jsonBytes), nil
}

// limitOrDefault returns the configured limit or 100 when unset/null.
func limitOrDefault(l types.Int64) int {
	if l.IsNull() || l.IsUnknown() {
		return 100
	}
	return int(l.ValueInt64())
}

// ===========================================================================
// 1. corex_config_status  - GET /config/status
// ===========================================================================

type configStatusDataSource struct {
	cli *client.Client
}

func NewConfigStatusDataSource() datasource.DataSource {
	return &configStatusDataSource{}
}

type configStatusModel struct {
	Status        types.String `tfsdk:"status"`
	PendingChanges types.Bool  `tfsdk:"pending_changes"`
	LastAppliedAt types.String `tfsdk:"last_applied_at"`
}

func (d *configStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_status"
}

func (d *configStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns the current coreX configuration status.",
		Attributes: map[string]schema.Attribute{
			"status":          schema.StringAttribute{Computed: true, Description: "Configuration status."},
			"pending_changes": schema.BoolAttribute{Computed: true, Description: "Whether there are pending changes."},
			"last_applied_at": schema.StringAttribute{Computed: true, Description: "Timestamp of last applied configuration."},
		},
	}
}

func (d *configStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *configStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data configStatusModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result struct {
		Status        string `json:"status"`
		PendingChanges bool   `json:"pending_changes"`
		LastAppliedAt string `json:"last_applied_at"`
	}
	if err := d.cli.Get(ctx, "/config/status", &result); err != nil {
		resp.Diagnostics.AddError("Failed to read config status", err.Error())
		return
	}

	data.Status = types.StringValue(result.Status)
	data.PendingChanges = types.BoolValue(result.PendingChanges)
	if result.LastAppliedAt != "" {
		data.LastAppliedAt = types.StringValue(result.LastAppliedAt)
	} else {
		data.LastAppliedAt = types.StringNull()
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 2. corex_config_preview - GET /config/preview
// ===========================================================================

type configPreviewDataSource struct {
	cli *client.Client
}

func NewConfigPreviewDataSource() datasource.DataSource {
	return &configPreviewDataSource{}
}

type configPreviewModel struct {
	Config types.String `tfsdk:"config"`
}

func (d *configPreviewDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_preview"
}

func (d *configPreviewDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns the preview of the generated coreX configuration.",
		Attributes: map[string]schema.Attribute{
			"config": schema.StringAttribute{Computed: true, Description: "Preview of the generated configuration."},
		},
	}
}

func (d *configPreviewDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *configPreviewDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data configPreviewModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result struct {
		Config string `json:"config"`
	}
	if err := d.cli.Get(ctx, "/config/preview", &result); err != nil {
		resp.Diagnostics.AddError("Failed to read config preview", err.Error())
		return
	}
	data.Config = types.StringValue(result.Config)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 3. corex_config_diff - GET /config/diff
// ===========================================================================

type configDiffDataSource struct {
	cli *client.Client
}

func NewConfigDiffDataSource() datasource.DataSource {
	return &configDiffDataSource{}
}

type configDiffModel struct {
	Diff types.String `tfsdk:"diff"`
}

func (d *configDiffDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_diff"
}

func (d *configDiffDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns the diff between the current and pending coreX configuration.",
		Attributes: map[string]schema.Attribute{
			"diff": schema.StringAttribute{Computed: true, Description: "Configuration diff."},
		},
	}
}

func (d *configDiffDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *configDiffDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data configDiffModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result struct {
		Diff string `json:"diff"`
	}
	if err := d.cli.Get(ctx, "/config/diff", &result); err != nil {
		resp.Diagnostics.AddError("Failed to read config diff", err.Error())
		return
	}
	data.Diff = types.StringValue(result.Diff)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 4. corex_config_snapshots - GET /config/snapshots
// ===========================================================================

type configSnapshotsDataSource struct {
	cli *client.Client
}

func NewConfigSnapshotsDataSource() datasource.DataSource {
	return &configSnapshotsDataSource{}
}

type configSnapshotModel struct {
	ID          types.Int64  `tfsdk:"id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
}

type configSnapshotsModel struct {
	Snapshots []configSnapshotModel `tfsdk:"snapshots"`
}

func (d *configSnapshotsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_snapshots"
}

func (d *configSnapshotsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns the list of coreX configuration snapshots.",
		Attributes: map[string]schema.Attribute{
			"snapshots": schema.ListAttribute{
				Computed:    true,
				Description: "List of configuration snapshots.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":          types.Int64Type,
						"created_at":  types.StringType,
						"description": types.StringType,
						"status":      types.StringType,
					},
				},
			},
		},
	}
}

func (d *configSnapshotsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *configSnapshotsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data configSnapshotsModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result []struct {
		ID          int    `json:"id"`
		CreatedAt   string `json:"created_at"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := d.cli.Get(ctx, "/config/snapshots", &result); err != nil {
		resp.Diagnostics.AddError("Failed to read config snapshots", err.Error())
		return
	}

	data.Snapshots = make([]configSnapshotModel, 0, len(result))
	for _, s := range result {
		data.Snapshots = append(data.Snapshots, configSnapshotModel{
			ID:          types.Int64Value(int64(s.ID)),
			CreatedAt:   types.StringValue(s.CreatedAt),
			Description: types.StringValue(s.Description),
			Status:      types.StringValue(s.Status),
		})
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 5. corex_backend - GET /backends (find by name)
// ===========================================================================

type backendDataSource struct {
	cli *client.Client
}

func NewBackendDataSource() datasource.DataSource {
	return &backendDataSource{}
}

type backendDSModel struct {
	Name      types.String `tfsdk:"name"`
	ID        types.Int64  `tfsdk:"id"`
	Mode      types.String `tfsdk:"mode"`
	Protocol  types.String `tfsdk:"protocol"`
	Algorithm types.String `tfsdk:"algorithm"`
}

func (d *backendDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backend"
}

func (d *backendDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a coreX backend by name.",
		Attributes: map[string]schema.Attribute{
			"name":      schema.StringAttribute{Required: true, Description: "Backend name to look up."},
			"id":        schema.Int64Attribute{Computed: true, Description: "Numeric ID of the backend."},
			"mode":      schema.StringAttribute{Computed: true, Description: "HAProxy mode."},
			"protocol":  schema.StringAttribute{Computed: true, Description: "Backend protocol."},
			"algorithm": schema.StringAttribute{Computed: true, Description: "Load balancing algorithm."},
		},
	}
}

func (d *backendDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *backendDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data backendDSModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var backends []struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Mode      string `json:"mode"`
		Protocol  string `json:"protocol"`
		Algorithm string `json:"algorithm"`
	}
	if err := d.cli.Get(ctx, "/backends", &backends); err != nil {
		resp.Diagnostics.AddError("Failed to list backends", err.Error())
		return
	}

	found := false
	for _, b := range backends {
		if b.Name == data.Name.ValueString() {
			data.ID = types.Int64Value(int64(b.ID))
			data.Mode = types.StringValue(b.Mode)
			data.Protocol = types.StringValue(b.Protocol)
			data.Algorithm = types.StringValue(b.Algorithm)
			found = true
			break
		}
	}
	if !found {
		resp.Diagnostics.AddError("Backend not found", fmt.Sprintf("No backend with name %q", data.Name.ValueString()))
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 6. corex_listener - GET /listeners (find by name)
// ===========================================================================

type listenerDataSource struct {
	cli *client.Client
}

func NewListenerDataSource() datasource.DataSource {
	return &listenerDataSource{}
}

type listenerDSModel struct {
	Name            types.String `tfsdk:"name"`
	ID              types.Int64  `tfsdk:"id"`
	BindAddress     types.String `tfsdk:"bind_address"`
	BindPort        types.Int64  `tfsdk:"bind_port"`
	Mode            types.String `tfsdk:"mode"`
	Protocol        types.String `tfsdk:"protocol"`
	SslEnabled      types.Bool   `tfsdk:"ssl_enabled"`
	DefaultBackendID types.Int64 `tfsdk:"default_backend_id"`
}

func (d *listenerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_listener"
}

func (d *listenerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a coreX listener by name.",
		Attributes: map[string]schema.Attribute{
			"name":               schema.StringAttribute{Required: true, Description: "Listener name to look up."},
			"id":                 schema.Int64Attribute{Computed: true, Description: "Numeric ID of the listener."},
			"bind_address":       schema.StringAttribute{Computed: true, Description: "Bind address."},
			"bind_port":          schema.Int64Attribute{Computed: true, Description: "Bind port."},
			"mode":               schema.StringAttribute{Computed: true, Description: "HAProxy mode."},
			"protocol":           schema.StringAttribute{Computed: true, Description: "Listener protocol."},
			"ssl_enabled":        schema.BoolAttribute{Computed: true, Description: "Whether SSL is enabled."},
			"default_backend_id": schema.Int64Attribute{Computed: true, Description: "Default backend ID."},
		},
	}
}

func (d *listenerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *listenerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data listenerDSModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var listeners []struct {
		ID              int    `json:"id"`
		Name            string `json:"name"`
		BindAddress     string `json:"bind_address"`
		BindPort        int    `json:"bind_port"`
		Mode            string `json:"mode"`
		Protocol        string `json:"protocol"`
		SslEnabled      bool   `json:"ssl_enabled"`
		DefaultBackendID int   `json:"default_backend_id"`
	}
	if err := d.cli.Get(ctx, "/listeners", &listeners); err != nil {
		resp.Diagnostics.AddError("Failed to list listeners", err.Error())
		return
	}

	found := false
	for _, l := range listeners {
		if l.Name == data.Name.ValueString() {
			data.ID = types.Int64Value(int64(l.ID))
			data.BindAddress = types.StringValue(l.BindAddress)
			data.BindPort = types.Int64Value(int64(l.BindPort))
			data.Mode = types.StringValue(l.Mode)
			data.Protocol = types.StringValue(l.Protocol)
			data.SslEnabled = types.BoolValue(l.SslEnabled)
			data.DefaultBackendID = types.Int64Value(int64(l.DefaultBackendID))
			found = true
			break
		}
	}
	if !found {
		resp.Diagnostics.AddError("Listener not found", fmt.Sprintf("No listener with name %q", data.Name.ValueString()))
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 7. corex_system_stats - GET /stats
// ===========================================================================

type systemStatsDataSource struct {
	cli *client.Client
}

func NewSystemStatsDataSource() datasource.DataSource {
	return &systemStatsDataSource{}
}

type systemStatsModel struct {
	StatsJSON types.String `tfsdk:"stats_json"`
}

func (d *systemStatsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_stats"
}

func (d *systemStatsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns coreX system statistics as raw JSON.",
		Attributes: map[string]schema.Attribute{
			"stats_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the system stats response."},
		},
	}
}

func (d *systemStatsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *systemStatsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data systemStatsModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	raw, err := rawGet(ctx, d.cli, "/stats")
	if err != nil {
		resp.Diagnostics.AddError("Failed to read system stats", err.Error())
		return
	}
	data.StatsJSON = types.StringValue(raw)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 8. corex_haproxy_stats - GET /haproxy-stats
// ===========================================================================

type haproxyStatsDataSource struct {
	cli *client.Client
}

func NewHaproxyStatsDataSource() datasource.DataSource {
	return &haproxyStatsDataSource{}
}

type haproxyStatsModel struct {
	StatsJSON types.String `tfsdk:"stats_json"`
}

func (d *haproxyStatsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_haproxy_stats"
}

func (d *haproxyStatsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns HAProxy statistics as raw JSON.",
		Attributes: map[string]schema.Attribute{
			"stats_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the HAProxy stats response."},
		},
	}
}

func (d *haproxyStatsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *haproxyStatsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data haproxyStatsModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	raw, err := rawGet(ctx, d.cli, "/haproxy-stats")
	if err != nil {
		resp.Diagnostics.AddError("Failed to read HAProxy stats", err.Error())
		return
	}
	data.StatsJSON = types.StringValue(raw)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 9. corex_health - GET /system/health
// ===========================================================================

type healthDataSource struct {
	cli *client.Client
}

func NewHealthDataSource() datasource.DataSource {
	return &healthDataSource{}
}

type healthModel struct {
	Status     types.String `tfsdk:"status"`
	HealthJSON types.String `tfsdk:"health_json"`
}

func (d *healthDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_health"
}

func (d *healthDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns coreX system health status.",
		Attributes: map[string]schema.Attribute{
			"status":      schema.StringAttribute{Computed: true, Description: "Health status string."},
			"health_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the health response."},
		},
	}
}

func (d *healthDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *healthDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data healthModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var raw map[string]interface{}
	if err := d.cli.Get(ctx, "/system/health", &raw); err != nil {
		resp.Diagnostics.AddError("Failed to read system health", err.Error())
		return
	}

	if s, ok := raw["status"].(string); ok {
		data.Status = types.StringValue(s)
	} else {
		data.Status = types.StringNull()
	}

	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal health response", err.Error())
		return
	}
	data.HealthJSON = types.StringValue(string(jsonBytes))

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 10. corex_audit_events - GET /audit-events?limit={limit}
// ===========================================================================

type auditEventsDataSource struct {
	cli *client.Client
}

func NewAuditEventsDataSource() datasource.DataSource {
	return &auditEventsDataSource{}
}

type auditEventsModel struct {
	Limit      types.Int64  `tfsdk:"limit"`
	EventsJSON types.String `tfsdk:"events_json"`
}

func (d *auditEventsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_events"
}

func (d *auditEventsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns recent coreX audit events as raw JSON.",
		Attributes: map[string]schema.Attribute{
			"limit":       schema.Int64Attribute{Optional: true, Description: "Maximum number of events to return. Defaults to 100."},
			"events_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the audit events response."},
		},
	}
}

func (d *auditEventsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *auditEventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data auditEventsModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/audit-events?limit=" + strconv.Itoa(limitOrDefault(data.Limit))
	raw, err := rawGet(ctx, d.cli, path)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read audit events", err.Error())
		return
	}
	data.EventsJSON = types.StringValue(raw)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 11. corex_recent_logs - GET /logs/recent?limit={limit}
// ===========================================================================

type recentLogsDataSource struct {
	cli *client.Client
}

func NewRecentLogsDataSource() datasource.DataSource {
	return &recentLogsDataSource{}
}

type recentLogsModel struct {
	Limit    types.Int64  `tfsdk:"limit"`
	LogsJSON types.String `tfsdk:"logs_json"`
}

func (d *recentLogsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_recent_logs"
}

func (d *recentLogsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns recent coreX logs as raw JSON.",
		Attributes: map[string]schema.Attribute{
			"limit":     schema.Int64Attribute{Optional: true, Description: "Maximum number of log entries to return. Defaults to 100."},
			"logs_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the recent logs response."},
		},
	}
}

func (d *recentLogsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *recentLogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data recentLogsModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/logs/recent?limit=" + strconv.Itoa(limitOrDefault(data.Limit))
	raw, err := rawGet(ctx, d.cli, path)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read recent logs", err.Error())
		return
	}
	data.LogsJSON = types.StringValue(raw)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 12. corex_stick_tables - GET /haproxy/tables
// ===========================================================================

type stickTablesDataSource struct {
	cli *client.Client
}

func NewStickTablesDataSource() datasource.DataSource {
	return &stickTablesDataSource{}
}

type stickTablesModel struct {
	TablesJSON types.String `tfsdk:"tables_json"`
}

func (d *stickTablesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stick_tables"
}

func (d *stickTablesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns the list of HAProxy stick tables as raw JSON.",
		Attributes: map[string]schema.Attribute{
			"tables_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the stick tables response."},
		},
	}
}

func (d *stickTablesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *stickTablesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data stickTablesModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	raw, err := rawGet(ctx, d.cli, "/haproxy/tables")
	if err != nil {
		resp.Diagnostics.AddError("Failed to read stick tables", err.Error())
		return
	}
	data.TablesJSON = types.StringValue(raw)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 13. corex_stick_table - GET /haproxy/tables/{name}?limit={limit}
// ===========================================================================

type stickTableDataSource struct {
	cli *client.Client
}

func NewStickTableDataSource() datasource.DataSource {
	return &stickTableDataSource{}
}

type stickTableModel struct {
	Name      types.String `tfsdk:"name"`
	Limit     types.Int64  `tfsdk:"limit"`
	TableJSON types.String `tfsdk:"table_json"`
}

func (d *stickTableDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stick_table"
}

func (d *stickTableDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns a single HAProxy stick table by name as raw JSON.",
		Attributes: map[string]schema.Attribute{
			"name":       schema.StringAttribute{Required: true, Description: "Stick table name."},
			"limit":      schema.Int64Attribute{Optional: true, Description: "Maximum number of entries to return. Defaults to 100."},
			"table_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the stick table response."},
		},
	}
}

func (d *stickTableDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *stickTableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data stickTableModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/haproxy/tables/" + data.Name.ValueString() + "?limit=" + strconv.Itoa(limitOrDefault(data.Limit))
	raw, err := rawGet(ctx, d.cli, path)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read stick table", err.Error())
		return
	}
	data.TableJSON = types.StringValue(raw)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 14. corex_valkey_info - GET /valkey/info
// ===========================================================================

type valkeyInfoDataSource struct {
	cli *client.Client
}

func NewValkeyInfoDataSource() datasource.DataSource {
	return &valkeyInfoDataSource{}
}

type valkeyInfoModel struct {
	InfoJSON types.String `tfsdk:"info_json"`
}

func (d *valkeyInfoDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_valkey_info"
}

func (d *valkeyInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns Valkey server info as raw JSON.",
		Attributes: map[string]schema.Attribute{
			"info_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the Valkey info response."},
		},
	}
}

func (d *valkeyInfoDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *valkeyInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data valkeyInfoModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	raw, err := rawGet(ctx, d.cli, "/valkey/info")
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Valkey info", err.Error())
		return
	}
	data.InfoJSON = types.StringValue(raw)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 15. corex_valkey_namespaces - GET /valkey/namespaces
// ===========================================================================

type valkeyNamespacesDataSource struct {
	cli *client.Client
}

func NewValkeyNamespacesDataSource() datasource.DataSource {
	return &valkeyNamespacesDataSource{}
}

type valkeyNamespacesModel struct {
	NamespacesJSON types.String `tfsdk:"namespaces_json"`
}

func (d *valkeyNamespacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_valkey_namespaces"
}

func (d *valkeyNamespacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns the list of Valkey namespaces as raw JSON.",
		Attributes: map[string]schema.Attribute{
			"namespaces_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the Valkey namespaces response."},
		},
	}
}

func (d *valkeyNamespacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *valkeyNamespacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data valkeyNamespacesModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	raw, err := rawGet(ctx, d.cli, "/valkey/namespaces")
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Valkey namespaces", err.Error())
		return
	}
	data.NamespacesJSON = types.StringValue(raw)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 16. corex_geoip_status - GET /settings/geoip/status
// ===========================================================================

type geoipStatusDataSource struct {
	cli *client.Client
}

func NewGeoipStatusDataSource() datasource.DataSource {
	return &geoipStatusDataSource{}
}

type geoipStatusModel struct {
	StatusJSON types.String `tfsdk:"status_json"`
}

func (d *geoipStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_geoip_status"
}

func (d *geoipStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns the GeoIP settings status as raw JSON.",
		Attributes: map[string]schema.Attribute{
			"status_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the GeoIP status response."},
		},
	}
}

func (d *geoipStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *geoipStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data geoipStatusModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	raw, err := rawGet(ctx, d.cli, "/settings/geoip/status")
	if err != nil {
		resp.Diagnostics.AddError("Failed to read GeoIP status", err.Error())
		return
	}
	data.StatusJSON = types.StringValue(raw)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 17. corex_asn_lookup - GET /geoip/asn?ip={ip}
// ===========================================================================

type asnLookupDataSource struct {
	cli *client.Client
}

func NewAsnLookupDataSource() datasource.DataSource {
	return &asnLookupDataSource{}
}

type asnLookupModel struct {
	IP         types.String `tfsdk:"ip"`
	ResultJSON types.String `tfsdk:"result_json"`
}

func (d *asnLookupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asn_lookup"
}

func (d *asnLookupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up ASN information for an IP address.",
		Attributes: map[string]schema.Attribute{
			"ip":          schema.StringAttribute{Required: true, Description: "IP address to look up."},
			"result_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the ASN lookup result."},
		},
	}
}

func (d *asnLookupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *asnLookupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data asnLookupModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/geoip/asn?ip=" + data.IP.ValueString()
	raw, err := rawGet(ctx, d.cli, path)
	if err != nil {
		resp.Diagnostics.AddError("Failed to perform ASN lookup", err.Error())
		return
	}
	data.ResultJSON = types.StringValue(raw)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 18. corex_ssl_labs_scans - GET /certificates/{cert_id}/ssllabs/scans
// ===========================================================================

type sslLabsScansDataSource struct {
	cli *client.Client
}

func NewSslLabsScansDataSource() datasource.DataSource {
	return &sslLabsScansDataSource{}
}

type sslLabsScansModel struct {
	CertID    types.Int64  `tfsdk:"cert_id"`
	ScansJSON types.String `tfsdk:"scans_json"`
}

func (d *sslLabsScansDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssl_labs_scans"
}

func (d *sslLabsScansDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns SSL Labs scans for a certificate as raw JSON.",
		Attributes: map[string]schema.Attribute{
			"cert_id":    schema.Int64Attribute{Required: true, Description: "Certificate ID."},
			"scans_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the SSL Labs scans response."},
		},
	}
}

func (d *sslLabsScansDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *sslLabsScansDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data sslLabsScansModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/certificates/" + strconv.FormatInt(data.CertID.ValueInt64(), 10) + "/ssllabs/scans"
	raw, err := rawGet(ctx, d.cli, path)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read SSL Labs scans", err.Error())
		return
	}
	data.ScansJSON = types.StringValue(raw)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 19. corex_mcp_gateway_status - GET /mcp/gateway/status
// ===========================================================================

type mcpGatewayStatusDataSource struct {
	cli *client.Client
}

func NewMcpGatewayStatusDataSource() datasource.DataSource {
	return &mcpGatewayStatusDataSource{}
}

type mcpGatewayStatusModel struct {
	StatusJSON types.String `tfsdk:"status_json"`
}

func (d *mcpGatewayStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_gateway_status"
}

func (d *mcpGatewayStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns the MCP gateway status as raw JSON.",
		Attributes: map[string]schema.Attribute{
			"status_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the MCP gateway status response."},
		},
	}
}

func (d *mcpGatewayStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *mcpGatewayStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data mcpGatewayStatusModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	raw, err := rawGet(ctx, d.cli, "/mcp/gateway/status")
	if err != nil {
		resp.Diagnostics.AddError("Failed to read MCP gateway status", err.Error())
		return
	}
	data.StatusJSON = types.StringValue(raw)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 20. corex_mcp_config_status - GET /mcp/config/status
// ===========================================================================

type mcpConfigStatusDataSource struct {
	cli *client.Client
}

func NewMcpConfigStatusDataSource() datasource.DataSource {
	return &mcpConfigStatusDataSource{}
}

type mcpConfigStatusModel struct {
	LastGenerated types.String `tfsdk:"last_generated"`
	BundleSize    types.Int64  `tfsdk:"bundle_size"`
	ConfigPath    types.String `tfsdk:"config_path"`
}

func (d *mcpConfigStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_config_status"
}

func (d *mcpConfigStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns the MCP configuration status.",
		Attributes: map[string]schema.Attribute{
			"last_generated": schema.StringAttribute{Computed: true, Description: "Timestamp of last MCP config generation."},
			"bundle_size":    schema.Int64Attribute{Computed: true, Description: "Size of the generated MCP bundle in bytes."},
			"config_path":    schema.StringAttribute{Computed: true, Description: "Path to the MCP config file."},
		},
	}
}

func (d *mcpConfigStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *mcpConfigStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data mcpConfigStatusModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result struct {
		LastGenerated string `json:"last_generated"`
		BundleSize    int64  `json:"bundle_size"`
		ConfigPath    string `json:"config_path"`
	}
	if err := d.cli.Get(ctx, "/mcp/config/status", &result); err != nil {
		resp.Diagnostics.AddError("Failed to read MCP config status", err.Error())
		return
	}

	if result.LastGenerated != "" {
		data.LastGenerated = types.StringValue(result.LastGenerated)
	} else {
		data.LastGenerated = types.StringNull()
	}
	data.BundleSize = types.Int64Value(result.BundleSize)
	if result.ConfigPath != "" {
		data.ConfigPath = types.StringValue(result.ConfigPath)
	} else {
		data.ConfigPath = types.StringNull()
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 21. corex_mcp_events - GET /mcp/events?limit={limit}
// ===========================================================================

type mcpEventsDataSource struct {
	cli *client.Client
}

func NewMcpEventsDataSource() datasource.DataSource {
	return &mcpEventsDataSource{}
}

type mcpEventsModel struct {
	Limit      types.Int64  `tfsdk:"limit"`
	EventsJSON types.String `tfsdk:"events_json"`
}

func (d *mcpEventsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_events"
}

func (d *mcpEventsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns recent MCP events as raw JSON.",
		Attributes: map[string]schema.Attribute{
			"limit":       schema.Int64Attribute{Optional: true, Description: "Maximum number of events to return. Defaults to 100."},
			"events_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the MCP events response."},
		},
	}
}

func (d *mcpEventsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *mcpEventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data mcpEventsModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/mcp/events?limit=" + strconv.Itoa(limitOrDefault(data.Limit))
	raw, err := rawGet(ctx, d.cli, path)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read MCP events", err.Error())
		return
	}
	data.EventsJSON = types.StringValue(raw)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// ===========================================================================
// 22. corex_mcp_marketplace_search - GET /mcp/marketplace/search?q={query}
// ===========================================================================

type mcpMarketplaceSearchDataSource struct {
	cli *client.Client
}

func NewMcpMarketplaceSearchDataSource() datasource.DataSource {
	return &mcpMarketplaceSearchDataSource{}
}

type mcpMarketplaceSearchModel struct {
	Query       types.String `tfsdk:"query"`
	ResultsJSON types.String `tfsdk:"results_json"`
}

func (d *mcpMarketplaceSearchDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mcp_marketplace_search"
}

func (d *mcpMarketplaceSearchDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Searches the MCP marketplace.",
		Attributes: map[string]schema.Attribute{
			"query":        schema.StringAttribute{Required: true, Description: "Search query."},
			"results_json": schema.StringAttribute{Computed: true, Description: "Raw JSON of the marketplace search results."},
		},
	}
}

func (d *mcpMarketplaceSearchDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.cli = getClient(req, resp)
}

func (d *mcpMarketplaceSearchDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data mcpMarketplaceSearchModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/mcp/marketplace/search?q=" + data.Query.ValueString()
	raw, err := rawGet(ctx, d.cli, path)
	if err != nil {
		resp.Diagnostics.AddError("Failed to search MCP marketplace", err.Error())
		return
	}
	data.ResultsJSON = types.StringValue(raw)

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

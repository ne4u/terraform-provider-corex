package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

const pageProtectSettingsID = "pp-settings"

var _ resource.Resource = &PageProtectSettingsResource{}

// PageProtectSettingsResource defines the corex_page_protect_settings resource (singleton).
type PageProtectSettingsResource struct {
	cli *client.Client
}

func NewPageProtectSettingsResource() resource.Resource {
	return &PageProtectSettingsResource{}
}

type pageProtectSettingsModel struct {
	ID                          types.String `tfsdk:"id"`
	MonitoringEnabled           types.Bool   `tfsdk:"monitoring_enabled"`
	ChangeDetectionEnabled      types.Bool   `tfsdk:"change_detection_enabled"`
	ChangeDetectionIntervalHours types.Int64 `tfsdk:"change_detection_interval_hours"`
	ReportRetentionDays         types.Int64  `tfsdk:"report_retention_days"`
	ReportPath                  types.String `tfsdk:"report_path"`
	BeaconInjectionEnabled      types.Bool   `tfsdk:"beacon_injection_enabled"`
	BeaconTrustEnabled          types.Bool   `tfsdk:"beacon_trust_enabled"`
	BeaconPaths                 types.List   `tfsdk:"beacon_paths"`
	BeaconContentTypes          types.List   `tfsdk:"beacon_content_types"`
	BeaconPatterns              types.List   `tfsdk:"beacon_patterns"`
	BackendIDs                  types.List   `tfsdk:"backend_ids"`
	AutoPruneStaleDays          types.Int64  `tfsdk:"auto_prune_stale_days"`
}

func (r *PageProtectSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_page_protect_settings"
}

func (r *PageProtectSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the page protection singleton settings in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Fixed singleton identifier.",
				Computed:    true,
			},
			"monitoring_enabled":              boolAttr("Enable monitoring.", false),
			"change_detection_enabled":        boolAttr("Enable change detection.", false),
			"change_detection_interval_hours": intAttr("Change detection interval in hours.", false),
			"report_retention_days":           intAttr("Report retention in days.", false),
			"report_path":                     stringAttr("Report path.", false),
			"beacon_injection_enabled":        boolAttr("Enable beacon injection.", false),
			"beacon_trust_enabled":            boolAttr("Enable beacon trust.", false),
			"beacon_paths":                    listAttr("Beacon paths.", false),
			"beacon_content_types":            listAttr("Beacon content types.", false),
			"beacon_patterns":                 listAttr("Beacon patterns.", false),
			"backend_ids":                     listIntAttr("Backend IDs.", false),
			"auto_prune_stale_days":           intAttr("Auto-prune stale days.", false),
		},
	}
}

func (r *PageProtectSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *PageProtectSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pageProtectSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Singleton: try GET first; if it exists, PUT, otherwise PUT to create.
	s := plan.toAPI(ctx)
	result, err := r.cli.UpdatePageProtectSettings(ctx, s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to set page protect settings", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	plan.ID = types.StringValue(pageProtectSettingsID)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PageProtectSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pageProtectSettingsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.cli.GetPageProtectSettings(ctx)
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read page protect settings", err.Error())
		return
	}

	state.fromAPI(ctx, s)
	state.ID = types.StringValue(pageProtectSettingsID)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *PageProtectSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pageProtectSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI(ctx)
	result, err := r.cli.UpdatePageProtectSettings(ctx, s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update page protect settings", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	plan.ID = types.StringValue(pageProtectSettingsID)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PageProtectSettingsResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton: delete is a no-op.
	_ = ctx
}

func (m *pageProtectSettingsModel) toAPI(ctx context.Context) *client.PageProtectSettings {
	return &client.PageProtectSettings{
		MonitoringEnabled:             m.MonitoringEnabled.ValueBool(),
		ChangeDetectionEnabled:        m.ChangeDetectionEnabled.ValueBool(),
		ChangeDetectionIntervalHours:  int(m.ChangeDetectionIntervalHours.ValueInt64()),
		ReportRetentionDays:           int(m.ReportRetentionDays.ValueInt64()),
		ReportPath:                    m.ReportPath.ValueString(),
		BeaconInjectionEnabled:        m.BeaconInjectionEnabled.ValueBool(),
		BeaconTrustEnabled:            m.BeaconTrustEnabled.ValueBool(),
		BeaconPaths:                   stringListToSlice(ctx, m.BeaconPaths),
		BeaconContentTypes:            stringListToSlice(ctx, m.BeaconContentTypes),
		BeaconPatterns:                stringListToSlice(ctx, m.BeaconPatterns),
		BackendIDs:                    intListToSlice(ctx, m.BackendIDs),
		AutoPruneStaleDays:            int(m.AutoPruneStaleDays.ValueInt64()),
	}
}

func (m *pageProtectSettingsModel) fromAPI(ctx context.Context, s *client.PageProtectSettings) {
	m.MonitoringEnabled = types.BoolValue(s.MonitoringEnabled)
	m.ChangeDetectionEnabled = types.BoolValue(s.ChangeDetectionEnabled)
	m.ChangeDetectionIntervalHours = types.Int64Value(int64(s.ChangeDetectionIntervalHours))
	m.ReportRetentionDays = types.Int64Value(int64(s.ReportRetentionDays))
	m.ReportPath = types.StringValue(s.ReportPath)
	m.BeaconInjectionEnabled = types.BoolValue(s.BeaconInjectionEnabled)
	m.BeaconTrustEnabled = types.BoolValue(s.BeaconTrustEnabled)
	m.BeaconPaths = sliceToStringList(ctx, s.BeaconPaths)
	m.BeaconContentTypes = sliceToStringList(ctx, s.BeaconContentTypes)
	m.BeaconPatterns = sliceToStringList(ctx, s.BeaconPatterns)
	m.BackendIDs = sliceToIntList(ctx, s.BackendIDs)
	m.AutoPruneStaleDays = types.Int64Value(int64(s.AutoPruneStaleDays))
}

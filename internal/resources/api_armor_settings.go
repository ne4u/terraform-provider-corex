package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

const apiArmorSettingsID = "api-armor-settings"

var _ resource.Resource = &ApiArmorSettingsResource{}

// ApiArmorSettingsResource defines the corex_api_armor_settings resource (singleton).
type ApiArmorSettingsResource struct {
	cli *client.Client
}

func NewApiArmorSettingsResource() resource.Resource {
	return &ApiArmorSettingsResource{}
}

type apiArmorSettingsModel struct {
	ID                              types.String `tfsdk:"id"`
	ApiArmorEnabled                 types.Bool   `tfsdk:"api_armor_enabled"`
	ApiArmorMaxBodyBytes            types.Int64  `tfsdk:"api_armor_max_body_bytes"`
	ApiArmorModuleEnabled           types.Bool   `tfsdk:"api_armor_module_enabled"`
	ApiArmorSchemaLearningEnabled   types.Bool   `tfsdk:"api_armor_schema_learning_enabled"`
	ApiArmorProfilingLearningEnabled types.Bool  `tfsdk:"api_armor_profiling_learning_enabled"`
	ApiArmorProfileRetentionDays    types.Int64  `tfsdk:"api_armor_profile_retention_days"`
	ApiArmorScope                   types.String `tfsdk:"api_armor_scope"`
	ApiArmorBackendIDs              types.List   `tfsdk:"api_armor_backend_ids"`
	ApiArmorPathPatterns            types.List   `tfsdk:"api_armor_path_patterns"`
}

func (r *ApiArmorSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_armor_settings"
}

func (r *ApiArmorSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the API Armor singleton settings in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Fixed singleton identifier.",
				Computed:    true,
			},
			"api_armor_enabled":                  boolAttr("Master toggle for API Armor.", false),
			"api_armor_max_body_bytes":           intAttr("Max request body size for inspection (bytes).", false),
			"api_armor_module_enabled":           boolAttr("Use Rust Lua module (true) vs Lua fallback (false).", false),
			"api_armor_schema_learning_enabled":  boolAttr("Enable learned schema inference.", false),
			"api_armor_profiling_learning_enabled": boolAttr("Enable behavioral profile learning.", false),
			"api_armor_profile_retention_days":   intAttr("Retention for behavioral profiles and anomalies (days).", false),
			"api_armor_scope":                    stringAttr("API Armor scope: listener, backend, or path.", false),
			"api_armor_backend_ids":              listIntAttr("Backend IDs to protect when scope is backend/path.", false),
			"api_armor_path_patterns":            listAttr("Path regex patterns to protect when scope is path.", false),
		},
	}
}

func (r *ApiArmorSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *ApiArmorSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiArmorSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI(ctx)
	result, err := r.cli.UpdateApiArmorSettings(ctx, s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to set api armor settings", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	plan.ID = types.StringValue(apiArmorSettingsID)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiArmorSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiArmorSettingsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.cli.GetApiArmorSettings(ctx)
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read api armor settings", err.Error())
		return
	}

	state.fromAPI(ctx, s)
	state.ID = types.StringValue(apiArmorSettingsID)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiArmorSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan apiArmorSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI(ctx)
	result, err := r.cli.UpdateApiArmorSettings(ctx, s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update api armor settings", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	plan.ID = types.StringValue(apiArmorSettingsID)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApiArmorSettingsResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton: delete is a no-op.
	_ = ctx
}

func (m *apiArmorSettingsModel) toAPI(ctx context.Context) *client.ApiArmorSettings {
	return &client.ApiArmorSettings{
		ApiArmorEnabled:                  m.ApiArmorEnabled.ValueBool(),
		ApiArmorMaxBodyBytes:             int(m.ApiArmorMaxBodyBytes.ValueInt64()),
		ApiArmorModuleEnabled:            m.ApiArmorModuleEnabled.ValueBool(),
		ApiArmorSchemaLearningEnabled:    m.ApiArmorSchemaLearningEnabled.ValueBool(),
		ApiArmorProfilingLearningEnabled: m.ApiArmorProfilingLearningEnabled.ValueBool(),
		ApiArmorProfileRetentionDays:     int(m.ApiArmorProfileRetentionDays.ValueInt64()),
		ApiArmorScope:                    m.ApiArmorScope.ValueString(),
		ApiArmorBackendIDs:               intListToSlice(ctx, m.ApiArmorBackendIDs),
		ApiArmorPathPatterns:             stringListToSlice(ctx, m.ApiArmorPathPatterns),
	}
}

func (m *apiArmorSettingsModel) fromAPI(ctx context.Context, s *client.ApiArmorSettings) {
	m.ApiArmorEnabled = types.BoolValue(s.ApiArmorEnabled)
	m.ApiArmorMaxBodyBytes = types.Int64Value(int64(s.ApiArmorMaxBodyBytes))
	m.ApiArmorModuleEnabled = types.BoolValue(s.ApiArmorModuleEnabled)
	m.ApiArmorSchemaLearningEnabled = types.BoolValue(s.ApiArmorSchemaLearningEnabled)
	m.ApiArmorProfilingLearningEnabled = types.BoolValue(s.ApiArmorProfilingLearningEnabled)
	m.ApiArmorProfileRetentionDays = types.Int64Value(int64(s.ApiArmorProfileRetentionDays))
	m.ApiArmorScope = types.StringValue(s.ApiArmorScope)
	m.ApiArmorBackendIDs = sliceToIntList(ctx, s.ApiArmorBackendIDs)
	m.ApiArmorPathPatterns = sliceToStringList(ctx, s.ApiArmorPathPatterns)
}

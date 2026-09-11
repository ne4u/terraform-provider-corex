package resources

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

var _ resource.Resource = &SslLabsSettingsResource{}

// SslLabsSettingsResource defines the corex_ssl_labs_settings resource
// (singleton per certificate).
type SslLabsSettingsResource struct {
	cli *client.Client
}

func NewSslLabsSettingsResource() resource.Resource {
	return &SslLabsSettingsResource{}
}

type sslLabsSettingsModel struct {
	ID              types.String `tfsdk:"id"`
	CertID          types.Int64  `tfsdk:"cert_id"`
	MaxScansPerHost types.Int64  `tfsdk:"max_scans_per_host"`
}

func (r *SslLabsSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssl_labs_settings"
}

func (r *SslLabsSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages SSL Labs scan settings for a certificate in coreX Manager (singleton per cert).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Certificate ID (as string) identifying this singleton.",
				Computed:    true,
			},
			"cert_id": schema.Int64Attribute{
				Description: "Certificate ID these settings apply to.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"max_scans_per_host": intAttr("Maximum concurrent SSL Labs scans per host.", false),
		},
	}
}

func (r *SslLabsSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *SslLabsSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan sslLabsSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	certID := int(plan.CertID.ValueInt64())
	s := plan.toAPI()
	result, err := r.cli.UpdateSslLabsSettings(ctx, certID, s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to set ssl labs settings", err.Error())
		return
	}

	plan.fromAPI(result)
	plan.ID = types.StringValue(strconv.Itoa(certID))
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SslLabsSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state sslLabsSettingsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	certID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		certID = int(state.CertID.ValueInt64())
	}

	s, err := r.cli.GetSslLabsSettings(ctx, certID)
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read ssl labs settings", err.Error())
		return
	}

	state.CertID = types.Int64Value(int64(certID))
	state.fromAPI(s)
	state.ID = types.StringValue(strconv.Itoa(certID))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *SslLabsSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan sslLabsSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	certID := int(plan.CertID.ValueInt64())
	s := plan.toAPI()
	result, err := r.cli.UpdateSslLabsSettings(ctx, certID, s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update ssl labs settings", err.Error())
		return
	}

	plan.fromAPI(result)
	plan.ID = types.StringValue(strconv.Itoa(certID))
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SslLabsSettingsResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton per cert: delete is a no-op (settings revert to defaults server-side).
	_ = ctx
}

func (m *sslLabsSettingsModel) toAPI() *client.SslLabsSettings {
	return &client.SslLabsSettings{
		MaxScansPerHost: int(m.MaxScansPerHost.ValueInt64()),
	}
}

func (m *sslLabsSettingsModel) fromAPI(s *client.SslLabsSettings) {
	m.MaxScansPerHost = types.Int64Value(int64(s.MaxScansPerHost))
}

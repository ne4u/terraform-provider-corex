package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

var _ resource.Resource = &CaptchaKeyResource{}
var _ resource.ResourceWithImportState = &CaptchaKeyResource{}

// CaptchaKeyResource defines the corex_captcha_key resource.
type CaptchaKeyResource struct {
	cli *client.Client
}

func NewCaptchaKeyResource() resource.Resource {
	return &CaptchaKeyResource{}
}

type captchaKeyModel struct {
	ID       types.String `tfsdk:"id"`
	SiteKey  types.String `tfsdk:"site_key"`
	Provider types.String `tfsdk:"provider"`
	Config   types.Map    `tfsdk:"config"`
}

func (r *CaptchaKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_captcha_key"
}

func (r *CaptchaKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a CAPTCHA site key in coreX Manager (via the Cap service proxy).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Site key (used as the resource identifier).",
				Computed:    true,
			},
			"site_key": stringAttrComputed("Cap site key."),
			"provider": stringAttr("Captcha provider.", false),
			"config":   mapAttr("Provider-specific key configuration.", false),
		},
	}
}

func (r *CaptchaKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *CaptchaKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan captchaKeyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	k := plan.toAPI(ctx)
	result, err := r.cli.CreateCaptchaKey(ctx, k)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create captcha key", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CaptchaKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state captchaKeyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteKey := state.ID.ValueString()
	if siteKey == "" {
		siteKey = state.SiteKey.ValueString()
	}
	k, err := r.cli.GetCaptchaKey(ctx, siteKey)
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read captcha key", err.Error())
		return
	}

	state.fromAPI(ctx, k)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *CaptchaKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan captchaKeyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteKey := plan.ID.ValueString()
	if siteKey == "" {
		siteKey = plan.SiteKey.ValueString()
	}
	k := plan.toAPI(ctx)
	result, err := r.cli.UpdateCaptchaKey(ctx, siteKey, k)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update captcha key", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CaptchaKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state captchaKeyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	siteKey := state.ID.ValueString()
	if siteKey == "" {
		siteKey = state.SiteKey.ValueString()
	}
	err := r.cli.DeleteCaptchaKey(ctx, siteKey)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete captcha key", err.Error())
		return
	}
}

func (r *CaptchaKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by site_key directly (the Cap service identifier).
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("site_key"), types.StringValue(req.ID))...)
}

func (m *captchaKeyModel) toAPI(ctx context.Context) *client.CaptchaKey {
	return &client.CaptchaKey{
		SiteKey:  m.SiteKey.ValueString(),
		Provider: m.Provider.ValueString(),
		Config:   stringMapToGoInterface(ctx, m.Config),
	}
}

func (m *captchaKeyModel) fromAPI(ctx context.Context, k *client.CaptchaKey) {
	id := k.SiteKey
	if id == "" {
		id = k.ID
	}
	m.ID = types.StringValue(id)
	m.SiteKey = types.StringValue(k.SiteKey)
	m.Provider = types.StringValue(k.Provider)
	m.Config = interfaceMapToString(ctx, k.Config)
}

package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

const captchaSettingsID = "captcha-settings"

var _ resource.Resource = &CaptchaSettingsResource{}

// CaptchaSettingsResource defines the corex_captcha_settings resource (singleton).
type CaptchaSettingsResource struct {
	cli *client.Client
}

func NewCaptchaSettingsResource() resource.Resource {
	return &CaptchaSettingsResource{}
}

type captchaSettingsModel struct {
	ID                  types.String `tfsdk:"id"`
	CaptchaProvider     types.String `tfsdk:"captcha_provider"`
	CapSiteKey          types.String `tfsdk:"cap_site_key"`
	CapSecret           types.String `tfsdk:"cap_secret"`
	RecaptchaSiteKey    types.String `tfsdk:"recaptcha_site_key"`
	RecaptchaSecret     types.String `tfsdk:"recaptcha_secret"`
	TurnstileSiteKey    types.String `tfsdk:"turnstile_site_key"`
	TurnstileSecret     types.String `tfsdk:"turnstile_secret"`
	CaptchaValidSeconds types.Int64  `tfsdk:"captcha_valid_seconds"`
	ChallengeURL        types.String `tfsdk:"challenge_url"`
	ProxyPath           types.String `tfsdk:"proxy_path"`
}

func (r *CaptchaSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_captcha_settings"
}

func (r *CaptchaSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the CAPTCHA singleton settings in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Fixed singleton identifier.",
				Computed:    true,
			},
			"captcha_provider":      stringAttr("Captcha provider (cap, recaptcha, turnstile).", false),
			"cap_site_key":          stringAttr("Cap site key.", false),
			"cap_secret":            stringAttrSensitive("Cap secret.", false),
			"recaptcha_site_key":    stringAttr("reCAPTCHA site key.", false),
			"recaptcha_secret":      stringAttrSensitive("reCAPTCHA secret.", false),
			"turnstile_site_key":    stringAttr("Turnstile site key.", false),
			"turnstile_secret":      stringAttrSensitive("Turnstile secret.", false),
			"captcha_valid_seconds": intAttr("Captcha validity in seconds.", false),
			"challenge_url":         stringAttr("Challenge URL.", false),
			"proxy_path":            stringAttr("Proxy path.", false),
		},
	}
}

func (r *CaptchaSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *CaptchaSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan captchaSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI()
	result, err := r.cli.UpdateCaptchaSettings(ctx, s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to set captcha settings", err.Error())
		return
	}

	plan.fromAPI(result)
	plan.ID = types.StringValue(captchaSettingsID)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CaptchaSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state captchaSettingsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.cli.GetCaptchaSettings(ctx)
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read captcha settings", err.Error())
		return
	}

	state.fromAPI(s)
	state.ID = types.StringValue(captchaSettingsID)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *CaptchaSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan captchaSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI()
	result, err := r.cli.UpdateCaptchaSettings(ctx, s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update captcha settings", err.Error())
		return
	}

	plan.fromAPI(result)
	plan.ID = types.StringValue(captchaSettingsID)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CaptchaSettingsResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton: delete is a no-op.
	_ = ctx
}

func (m *captchaSettingsModel) toAPI() *client.CaptchaSettings {
	s := &client.CaptchaSettings{
		CaptchaProvider:     m.CaptchaProvider.ValueString(),
		CaptchaValidSeconds: int(m.CaptchaValidSeconds.ValueInt64()),
		ChallengeURL:        m.ChallengeURL.ValueString(),
		ProxyPath:           m.ProxyPath.ValueString(),
	}
	if !m.CapSiteKey.IsNull() {
		s.CapSiteKey = nilIfEmpty(m.CapSiteKey.ValueString())
	}
	if !m.CapSecret.IsNull() {
		s.CapSecret = nilIfEmpty(m.CapSecret.ValueString())
	}
	if !m.RecaptchaSiteKey.IsNull() {
		s.RecaptchaSiteKey = nilIfEmpty(m.RecaptchaSiteKey.ValueString())
	}
	if !m.RecaptchaSecret.IsNull() {
		s.RecaptchaSecret = nilIfEmpty(m.RecaptchaSecret.ValueString())
	}
	if !m.TurnstileSiteKey.IsNull() {
		s.TurnstileSiteKey = nilIfEmpty(m.TurnstileSiteKey.ValueString())
	}
	if !m.TurnstileSecret.IsNull() {
		s.TurnstileSecret = nilIfEmpty(m.TurnstileSecret.ValueString())
	}
	return s
}

func (m *captchaSettingsModel) fromAPI(s *client.CaptchaSettings) {
	m.CaptchaProvider = types.StringValue(s.CaptchaProvider)
	m.CaptchaValidSeconds = types.Int64Value(int64(s.CaptchaValidSeconds))
	m.ChallengeURL = types.StringValue(s.ChallengeURL)
	m.ProxyPath = types.StringValue(s.ProxyPath)
	if s.CapSiteKey != nil {
		m.CapSiteKey = types.StringValue(*s.CapSiteKey)
	} else {
		m.CapSiteKey = types.StringNull()
	}
	if s.CapSecret != nil {
		m.CapSecret = types.StringValue(*s.CapSecret)
	} else {
		m.CapSecret = types.StringNull()
	}
	if s.RecaptchaSiteKey != nil {
		m.RecaptchaSiteKey = types.StringValue(*s.RecaptchaSiteKey)
	} else {
		m.RecaptchaSiteKey = types.StringNull()
	}
	if s.RecaptchaSecret != nil {
		m.RecaptchaSecret = types.StringValue(*s.RecaptchaSecret)
	} else {
		m.RecaptchaSecret = types.StringNull()
	}
	if s.TurnstileSiteKey != nil {
		m.TurnstileSiteKey = types.StringValue(*s.TurnstileSiteKey)
	} else {
		m.TurnstileSiteKey = types.StringNull()
	}
	if s.TurnstileSecret != nil {
		m.TurnstileSecret = types.StringValue(*s.TurnstileSecret)
	} else {
		m.TurnstileSecret = types.StringNull()
	}
}

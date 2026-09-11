package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

// maxmindKeyID is the fixed singleton ID for the MaxMind license key resource.
const maxmindKeyID = "maxmind-key"

// Ensure the implementation satisfies the resource.Resource interface.
var _ resource.Resource = &MaxmindLicenseKeyResource{}
var _ resource.ResourceWithImportState = &MaxmindLicenseKeyResource{}

// MaxmindLicenseKeyResource defines the corex_maxmind_license_key resource.
type MaxmindLicenseKeyResource struct {
	cli *client.Client
}

func NewMaxmindLicenseKeyResource() resource.Resource {
	return &MaxmindLicenseKeyResource{}
}

// maxmindLicenseKeyModel maps the Terraform schema to the API model.
type maxmindLicenseKeyModel struct {
	ID    types.String `tfsdk:"id"`
	Value types.String `tfsdk:"value"`
}

func (r *MaxmindLicenseKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maxmind_license_key"
}

func (r *MaxmindLicenseKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the MaxMind license key setting in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Fixed singleton ID for the MaxMind license key.",
				Computed:    true,
			},
			"value": stringAttrSensitive("MaxMind license key.", true),
		},
	}
}

func (r *MaxmindLicenseKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *MaxmindLicenseKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan maxmindLicenseKeyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	value := plan.Value.ValueString()
	err := r.cli.SetMaxmindLicenseKey(ctx, value)
	if err != nil {
		resp.Diagnostics.AddError("Failed to set MaxMind license key", err.Error())
		return
	}

	plan.ID = types.StringValue(maxmindKeyID)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MaxmindLicenseKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state maxmindLicenseKeyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mk, err := r.cli.GetMaxmindLicenseKey(ctx)
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read MaxMind license key", err.Error())
		return
	}

	state.ID = types.StringValue(maxmindKeyID)
	state.Value = types.StringValue(mk.Value)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *MaxmindLicenseKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan maxmindLicenseKeyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	value := plan.Value.ValueString()
	err := r.cli.SetMaxmindLicenseKey(ctx, value)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update MaxMind license key", err.Error())
		return
	}

	plan.ID = types.StringValue(maxmindKeyID)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MaxmindLicenseKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state maxmindLicenseKeyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete is a no-op: clear the license key by setting an empty value.
	_ = r.cli.SetMaxmindLicenseKey(ctx, "")
}

func (r *MaxmindLicenseKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by the fixed singleton ID.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(maxmindKeyID))...)
}

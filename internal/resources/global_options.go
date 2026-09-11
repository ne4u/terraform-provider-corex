package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

const globalOptionsID = "global-opts"

var _ resource.Resource = &GlobalOptionsResource{}

// GlobalOptionsResource defines the corex_global_options resource (singleton).
// The resource models the full list of global HAProxy options.
type GlobalOptionsResource struct {
	cli *client.Client
}

func NewGlobalOptionsResource() resource.Resource {
	return &GlobalOptionsResource{}
}

type haproxyOptionModel struct {
	Target    types.String `tfsdk:"target"`
	Directive types.String `tfsdk:"directive"`
	Value     types.String `tfsdk:"value"`
	Enabled   types.Bool   `tfsdk:"enabled"`
}

type globalOptionsModel struct {
	ID      types.String         `tfsdk:"id"`
	Options []haproxyOptionModel `tfsdk:"options"`
}

func (r *GlobalOptionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_options"
}

func (r *GlobalOptionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the global HAProxy options list in coreX Manager (singleton).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Fixed singleton identifier.",
				Computed:    true,
			},
			"options": schema.ListNestedAttribute{
				Description: "Global HAProxy options.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"target":    stringAttr("Option target (global).", false),
						"directive": stringAttr("HAProxy directive.", false),
						"value":     stringAttr("Directive value.", false),
						"enabled":   boolAttr("Whether this option is enabled.", false),
					},
				},
			},
		},
	}
}

func (r *GlobalOptionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *GlobalOptionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan globalOptionsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := plan.toAPI()
	result, err := r.cli.UpdateGlobalOptions(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to set global options", err.Error())
		return
	}

	plan.fromAPI(result)
	plan.ID = types.StringValue(globalOptionsID)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GlobalOptionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state globalOptionsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts, err := r.cli.GetGlobalOptions(ctx)
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read global options", err.Error())
		return
	}

	state.fromAPI(opts)
	state.ID = types.StringValue(globalOptionsID)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *GlobalOptionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan globalOptionsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := plan.toAPI()
	result, err := r.cli.UpdateGlobalOptions(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update global options", err.Error())
		return
	}

	plan.fromAPI(result)
	plan.ID = types.StringValue(globalOptionsID)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GlobalOptionsResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton: delete is a no-op.
	_ = ctx
}

func (m *globalOptionsModel) toAPI() []client.HaproxyOption {
	opts := make([]client.HaproxyOption, 0, len(m.Options))
	for _, o := range m.Options {
		opts = append(opts, client.HaproxyOption{
			Target:    o.Target.ValueString(),
			Directive: o.Directive.ValueString(),
			Value:     o.Value.ValueString(),
			Enabled:   o.Enabled.ValueBool(),
		})
	}
	return opts
}

func (m *globalOptionsModel) fromAPI(opts []client.HaproxyOption) {
	out := make([]haproxyOptionModel, 0, len(opts))
	for _, o := range opts {
		out = append(out, haproxyOptionModel{
			Target:    types.StringValue(o.Target),
			Directive: types.StringValue(o.Directive),
			Value:     types.StringValue(o.Value),
			Enabled:   types.BoolValue(o.Enabled),
		})
	}
	m.Options = out
}

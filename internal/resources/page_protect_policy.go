package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

var _ resource.Resource = &PageProtectPolicyResource{}
var _ resource.ResourceWithImportState = &PageProtectPolicyResource{}

// PageProtectPolicyResource defines the corex_page_protect_policy resource.
type PageProtectPolicyResource struct {
	cli *client.Client
}

func NewPageProtectPolicyResource() resource.Resource {
	return &PageProtectPolicyResource{}
}

type pageProtectPolicyModel struct {
	ID                  types.Int64  `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Enabled             types.Bool   `tfsdk:"enabled"`
	BackendIDs          types.List   `tfsdk:"backend_ids"`
	Mode                types.String `tfsdk:"mode"`
	SampleRatePercent   types.Int64  `tfsdk:"sample_rate_percent"`
	ReportPath          types.String `tfsdk:"report_path"`
	Directives          types.Map    `tfsdk:"directives"`
}

func (r *PageProtectPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_page_protect_policy"
}

func (r *PageProtectPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a page protection policy in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":                    idAttr(),
			"name":                  stringAttr("Policy name.", true),
			"enabled":               boolAttr("Whether the policy is enabled.", false),
			"backend_ids":           listIntAttr("Backend IDs this policy applies to.", false),
			"mode":                  stringAttr("Protection mode.", false),
			"sample_rate_percent":   intAttr("Sample rate percentage.", false),
			"report_path":           stringAttr("Report path.", false),
			"directives": schema.MapAttribute{
				Description:  "Directives map (key -> list of values).",
				Optional:     true,
				ElementType:  types.ListType{},
			},
		},
	}
}

func (r *PageProtectPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *PageProtectPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pageProtectPolicyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := plan.toAPI(ctx)
	result, err := r.cli.CreatePageProtectPolicy(ctx, p)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create page protect policy", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PageProtectPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pageProtectPolicyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := r.cli.GetPageProtectPolicy(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read page protect policy", err.Error())
		return
	}

	state.fromAPI(ctx, p)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *PageProtectPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pageProtectPolicyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := plan.toAPI(ctx)
	result, err := r.cli.UpdatePageProtectPolicy(ctx, int(plan.ID.ValueInt64()), p)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update page protect policy", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PageProtectPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pageProtectPolicyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeletePageProtectPolicy(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete page protect policy", err.Error())
		return
	}
}

func (r *PageProtectPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	policies, err := r.cli.ListPageProtectPolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list page protect policies for import", err.Error())
		return
	}

	for _, p := range policies {
		if p.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(p.ID)))...)
			return
		}
	}

	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Page protect policy not found", fmt.Sprintf("No policy with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

func (m *pageProtectPolicyModel) toAPI(ctx context.Context) *client.PageProtectPolicy {
	return &client.PageProtectPolicy{
		Name:              m.Name.ValueString(),
		Enabled:           m.Enabled.ValueBool(),
		BackendIDs:        intListToSlice(ctx, m.BackendIDs),
		Mode:              m.Mode.ValueString(),
		SampleRatePercent: int(m.SampleRatePercent.ValueInt64()),
		ReportPath:        m.ReportPath.ValueString(),
		Directives:        stringListMapToGo(ctx, m.Directives),
	}
}

func (m *pageProtectPolicyModel) fromAPI(ctx context.Context, p *client.PageProtectPolicy) {
	m.ID = types.Int64Value(int64(p.ID))
	m.Name = types.StringValue(p.Name)
	m.Enabled = types.BoolValue(p.Enabled)
	m.BackendIDs = sliceToIntList(ctx, p.BackendIDs)
	m.Mode = types.StringValue(p.Mode)
	m.SampleRatePercent = types.Int64Value(int64(p.SampleRatePercent))
	m.ReportPath = types.StringValue(p.ReportPath)
	m.Directives = goMapToStringListMap(ctx, p.Directives)
}

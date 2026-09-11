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

// Ensure the implementation satisfies the resource.Resource interface.
var _ resource.Resource = &RateLimitResource{}
var _ resource.ResourceWithImportState = &RateLimitResource{}

// RateLimitResource defines the corex_rate_limit resource.
type RateLimitResource struct {
	cli *client.Client
}

func NewRateLimitResource() resource.Resource {
	return &RateLimitResource{}
}

// rateLimitModel maps the Terraform schema to the API model.
type rateLimitModel struct {
	ID              types.Int64  `tfsdk:"id"`
	ListenerID      types.Int64  `tfsdk:"listener_id"`
	Name            types.String `tfsdk:"name"`
	LimitType       types.String `tfsdk:"limit_type"`
	Events          types.Int64  `tfsdk:"events"`
	WindowSeconds   types.Int64  `tfsdk:"window_seconds"`
	Burst           types.Int64  `tfsdk:"burst"`
	Action          types.String `tfsdk:"action"`
	DurationSeconds types.Int64  `tfsdk:"duration_seconds"`
	Expression      types.String `tfsdk:"expression"`
	RateKey         types.String `tfsdk:"rate_key"`
}

func (r *RateLimitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rate_limit"
}

func (r *RateLimitResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rate limit in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":               idAttr(),
			"listener_id":      intAttr("Listener ID the rate limit applies to.", true),
			"name":             stringAttr("Rate limit name.", true),
			"limit_type":       stringAttr("Limit type (sliding/fixed).", false),
			"events":           intAttr("Number of events allowed.", false),
			"window_seconds":   intAttr("Time window in seconds.", false),
			"burst":            intAttr("Burst allowance.", false),
			"action":           stringAttr("Action on limit exceeded.", false),
			"duration_seconds": intAttr("Duration of the action in seconds.", false),
			"expression":       stringAttr("Match expression.", false),
			"rate_key":         stringAttr("Rate key (source_ip/etc.).", false),
		},
	}
}

func (r *RateLimitResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *RateLimitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan rateLimitModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rl := plan.toAPI()
	result, err := r.cli.CreateRateLimit(ctx, rl)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create rate limit", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RateLimitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state rateLimitModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rl, err := r.cli.GetRateLimit(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read rate limit", err.Error())
		return
	}

	state.fromAPI(rl)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *RateLimitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan rateLimitModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	rl := plan.toAPI()
	result, err := r.cli.UpdateRateLimit(ctx, int(plan.ID.ValueInt64()), rl)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update rate limit", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RateLimitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state rateLimitModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteRateLimit(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete rate limit", err.Error())
		return
	}
}

func (r *RateLimitResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to import by name: list all rate limits, find matching name.
	limits, err := r.cli.ListRateLimits(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list rate limits for import", err.Error())
		return
	}

	for _, rl := range limits {
		if rl.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(rl.ID)))...)
			return
		}
	}

	// Try numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Rate limit not found", fmt.Sprintf("No rate limit with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *rateLimitModel) toAPI() *client.RateLimit {
	return &client.RateLimit{
		ListenerID:      int(m.ListenerID.ValueInt64()),
		Name:            m.Name.ValueString(),
		LimitType:       m.LimitType.ValueString(),
		Events:          int(m.Events.ValueInt64()),
		WindowSeconds:   int(m.WindowSeconds.ValueInt64()),
		Burst:           int(m.Burst.ValueInt64()),
		Action:          m.Action.ValueString(),
		DurationSeconds: int(m.DurationSeconds.ValueInt64()),
		Expression:      m.Expression.ValueString(),
		RateKey:         m.RateKey.ValueString(),
	}
}

// fromAPI populates the Terraform model from the API model.
func (m *rateLimitModel) fromAPI(rl *client.RateLimit) {
	m.ID = types.Int64Value(int64(rl.ID))
	m.ListenerID = types.Int64Value(int64(rl.ListenerID))
	m.Name = types.StringValue(rl.Name)
	m.LimitType = types.StringValue(rl.LimitType)
	m.Events = types.Int64Value(int64(rl.Events))
	m.WindowSeconds = types.Int64Value(int64(rl.WindowSeconds))
	m.Burst = types.Int64Value(int64(rl.Burst))
	m.Action = types.StringValue(rl.Action)
	m.DurationSeconds = types.Int64Value(int64(rl.DurationSeconds))
	m.Expression = types.StringValue(rl.Expression)
	m.RateKey = types.StringValue(rl.RateKey)
}

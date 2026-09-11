package resources

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

// Ensure the implementation satisfies the resource.Resource interface.
var _ resource.Resource = &CacheConfigResource{}
var _ resource.ResourceWithImportState = &CacheConfigResource{}

// CacheConfigResource defines the corex_cache_config resource.
type CacheConfigResource struct {
	cli *client.Client
}

func NewCacheConfigResource() resource.Resource {
	return &CacheConfigResource{}
}

// cacheConfigModel maps the Terraform schema to the API model.
type cacheConfigModel struct {
	ID                 types.Int64  `tfsdk:"id"`
	BackendID          types.Int64  `tfsdk:"backend_id"`
	HaproxyEnabled     types.Bool   `tfsdk:"haproxy_enabled"`
	HaproxyCacheSize   types.Int64  `tfsdk:"haproxy_cache_size"`
	HaproxyCacheMaxAge types.Int64  `tfsdk:"haproxy_cache_max_age"`
	HaproxyCacheVary   types.List   `tfsdk:"haproxy_cache_vary"`
	DiskCacheEnabled   types.Bool   `tfsdk:"disk_cache_enabled"`
	DiskCachePath      types.String `tfsdk:"disk_cache_path"`
	DiskCacheMaxSize   types.Int64  `tfsdk:"disk_cache_max_size"`
	DiskCacheMaxAge    types.Int64  `tfsdk:"disk_cache_max_age"`
	RFC7234Compliance  types.Bool   `tfsdk:"rfc7234_compliance"`
}

func (r *CacheConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cache_config"
}

func (r *CacheConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a cache configuration for a backend in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":                    idAttr(),
			"backend_id":            intAttr("Backend ID this cache config belongs to.", true),
			"haproxy_enabled":       boolAttr("Enable HAProxy cache.", false),
			"haproxy_cache_size":    intAttr("HAProxy cache size in bytes.", false),
			"haproxy_cache_max_age": intAttr("HAProxy cache max age in seconds.", false),
			"haproxy_cache_vary":    listAttr("HAProxy cache vary headers.", false),
			"disk_cache_enabled":    boolAttr("Enable disk cache.", false),
			"disk_cache_path":       stringAttr("Disk cache path.", false),
			"disk_cache_max_size":   intAttr("Disk cache max size in bytes.", false),
			"disk_cache_max_age":    intAttr("Disk cache max age in seconds.", false),
			"rfc7234_compliance":    boolAttr("Enable RFC 7234 compliance.", false),
		},
	}
}

func (r *CacheConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *CacheConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cacheConfigModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cc := plan.toAPI(ctx)
	result, err := r.cli.CreateCacheConfig(ctx, cc)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cache config", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	// Keyed by backend_id: use backend_id as the state ID.
	plan.ID = types.Int64Value(int64(result.BackendID))
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CacheConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cacheConfigModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cc, err := r.cli.GetCacheConfig(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read cache config", err.Error())
		return
	}

	state.fromAPI(ctx, cc)
	state.ID = types.Int64Value(int64(cc.BackendID))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *CacheConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cacheConfigModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cc := plan.toAPI(ctx)
	result, err := r.cli.UpdateCacheConfig(ctx, int(plan.BackendID.ValueInt64()), cc)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update cache config", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	plan.ID = types.Int64Value(int64(result.BackendID))
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CacheConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cacheConfigModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteCacheConfig(ctx, int(state.BackendID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete cache config", err.Error())
		return
	}
}

func (r *CacheConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Cache config is keyed by backend_id; import by backend_id.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected a numeric backend ID")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("backend_id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *cacheConfigModel) toAPI(ctx context.Context) *client.CacheConfig {
	return &client.CacheConfig{
		BackendID:           int(m.BackendID.ValueInt64()),
		HaproxyEnabled:      m.HaproxyEnabled.ValueBool(),
		HaproxyCacheSize:    int(m.HaproxyCacheSize.ValueInt64()),
		HaproxyCacheMaxAge:  int(m.HaproxyCacheMaxAge.ValueInt64()),
		HaproxyCacheVary:    stringListToSlice(ctx, m.HaproxyCacheVary),
		DiskCacheEnabled:    m.DiskCacheEnabled.ValueBool(),
		DiskCachePath:       m.DiskCachePath.ValueString(),
		DiskCacheMaxSize:    int(m.DiskCacheMaxSize.ValueInt64()),
		DiskCacheMaxAge:     int(m.DiskCacheMaxAge.ValueInt64()),
		RFC7234Compliance:   m.RFC7234Compliance.ValueBool(),
	}
}

// fromAPI populates the Terraform model from the API model.
func (m *cacheConfigModel) fromAPI(ctx context.Context, cc *client.CacheConfig) {
	m.ID = types.Int64Value(int64(cc.ID))
	m.BackendID = types.Int64Value(int64(cc.BackendID))
	m.HaproxyEnabled = types.BoolValue(cc.HaproxyEnabled)
	m.HaproxyCacheSize = types.Int64Value(int64(cc.HaproxyCacheSize))
	m.HaproxyCacheMaxAge = types.Int64Value(int64(cc.HaproxyCacheMaxAge))
	m.HaproxyCacheVary = sliceToStringList(ctx, cc.HaproxyCacheVary)
	m.DiskCacheEnabled = types.BoolValue(cc.DiskCacheEnabled)
	m.DiskCachePath = types.StringValue(cc.DiskCachePath)
	m.DiskCacheMaxSize = types.Int64Value(int64(cc.DiskCacheMaxSize))
	m.DiskCacheMaxAge = types.Int64Value(int64(cc.DiskCacheMaxAge))
	m.RFC7234Compliance = types.BoolValue(cc.RFC7234Compliance)
}

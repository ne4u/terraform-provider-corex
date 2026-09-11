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
var _ resource.Resource = &BackendResource{}
var _ resource.ResourceWithImportState = &BackendResource{}

// BackendResource defines the corex_backend resource.
type BackendResource struct {
	cli *client.Client
}

func NewBackendResource() resource.Resource {
	return &BackendResource{}
}

// backendModel maps the Terraform schema to the API model.
type backendModel struct {
	ID                      types.Int64  `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Mode                    types.String `tfsdk:"mode"`
	Protocol                types.String `tfsdk:"protocol"`
	Algorithm               types.String `tfsdk:"algorithm"`
	StickySessions          types.Bool   `tfsdk:"sticky_sessions"`
	CookieName              types.String `tfsdk:"cookie_name"`
	BalanceArgs             types.String `tfsdk:"balance_args"`
	HealthCheckEnabled      types.Bool   `tfsdk:"health_check_enabled"`
	HealthCheckInterval     types.Int64  `tfsdk:"health_check_interval"`
	HealthCheckURI          types.String `tfsdk:"health_check_uri"`
	HealthCheckMethod       types.String `tfsdk:"health_check_method"`
	Retries                 types.Int64  `tfsdk:"retries"`
	Redispatch              types.Bool   `tfsdk:"redispatch"`
	StickTable              types.Bool   `tfsdk:"stick_table"`
	StickTableSize          types.String `tfsdk:"stick_table_size"`
	StickTableExpire        types.String `tfsdk:"stick_table_expire"`
	StickTableType          types.String `tfsdk:"stick_table_type"`
	Resolvers               types.String `tfsdk:"resolvers"`
	HostHeader              types.String `tfsdk:"host_header"`
	RestoreClientIP         types.Bool   `tfsdk:"restore_client_ip"`
	ClientIPHeader          types.String `tfsdk:"client_ip_header"`
	HaproxyOptions          types.List   `tfsdk:"haproxy_options"`
}

func (r *BackendResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backend"
}

func (r *BackendResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a backend pool in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":                    idAttr(),
			"name":                  stringAttr("Backend name.", true),
			"mode":                  stringAttr("HAProxy mode (tcp/http).", true),
			"protocol":              stringAttr("Backend protocol.", false),
			"algorithm":             stringAttr("Load balancing algorithm.", false),
			"sticky_sessions":       boolAttr("Enable sticky sessions.", false),
			"cookie_name":           stringAttr("Sticky session cookie name.", false),
			"balance_args":          stringAttr("Balance algorithm arguments.", false),
			"health_check_enabled":  boolAttr("Enable health checks.", false),
			"health_check_interval": intAttr("Health check interval in ms.", false),
			"health_check_uri":      stringAttr("Health check URI.", false),
			"health_check_method":   stringAttr("Health check HTTP method.", false),
			"retries":               intAttr("Number of retries on connection failure.", false),
			"redispatch":            boolAttr("Enable redispatch to another server on failure.", false),
			"stick_table":           boolAttr("Enable stick table.", false),
			"stick_table_size":      stringAttr("Stick table size.", false),
			"stick_table_expire":    stringAttr("Stick table expiration.", false),
			"stick_table_type":      stringAttr("Stick table type.", false),
			"resolvers":             stringAttr("DNS resolvers to use.", false),
			"host_header":           stringAttr("Host header to send to backend.", false),
			"restore_client_ip":     boolAttr("Restore client IP (PROXY protocol).", false),
			"client_ip_header":      stringAttr("Client IP header name.", false),
			"haproxy_options": schema.ListNestedAttribute{
				Description: "Raw HAProxy options.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"target":    stringAttr("Option target (frontend/backend/global).", false),
						"directive": stringAttr("HAProxy directive.", false),
						"value":     stringAttr("Directive value.", false),
						"enabled":   boolAttr("Whether this option is enabled.", false),
					},
				},
			},
		},
	}
}

func (r *BackendResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *BackendResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan backendModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	b := plan.toAPI()
	result, err := r.cli.CreateBackend(ctx, b)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create backend", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BackendResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state backendModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	b, err := r.cli.GetBackend(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read backend", err.Error())
		return
	}

	state.fromAPI(b)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *BackendResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan backendModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	b := plan.toAPI()
	result, err := r.cli.UpdateBackend(ctx, int(plan.ID.ValueInt64()), b)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update backend", err.Error())
		return
	}

	plan.fromAPI(result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BackendResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state backendModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteBackend(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete backend", err.Error())
		return
	}
}

func (r *BackendResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to import by name: list all backends, find matching name.
	backends, err := r.cli.ListBackends(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list backends for import", err.Error())
		return
	}

	for _, b := range backends {
		if b.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(b.ID)))...)
			return
		}
	}

	// Try numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Backend not found", fmt.Sprintf("No backend with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *backendModel) toAPI() *client.Backend {
	b := &client.Backend{
		Name:                  m.Name.ValueString(),
		Mode:                  m.Mode.ValueString(),
		Protocol:              m.Protocol.ValueString(),
		Algorithm:             m.Algorithm.ValueString(),
		StickySessions:        m.StickySessions.ValueBool(),
		HealthCheckEnabled:    m.HealthCheckEnabled.ValueBool(),
		HealthCheckInterval:   int(m.HealthCheckInterval.ValueInt64()),
		HealthCheckURI:        m.HealthCheckURI.ValueString(),
		HealthCheckMethod:     m.HealthCheckMethod.ValueString(),
		Retries:               int(m.Retries.ValueInt64()),
		Redispatch:            m.Redispatch.ValueBool(),
		StickTable:            m.StickTable.ValueBool(),
		StickTableSize:        m.StickTableSize.ValueString(),
		StickTableExpire:      m.StickTableExpire.ValueString(),
		StickTableType:        m.StickTableType.ValueString(),
		RestoreClientIP:       m.RestoreClientIP.ValueBool(),
		ClientIPHeader:        m.ClientIPHeader.ValueString(),
	}
	if !m.CookieName.IsNull() {
		b.CookieName = stringPtr(m.CookieName.ValueString())
	}
	if !m.BalanceArgs.IsNull() {
		b.BalanceArgs = stringPtr(m.BalanceArgs.ValueString())
	}
	if !m.Resolvers.IsNull() {
		b.Resolvers = stringPtr(m.Resolvers.ValueString())
	}
	if !m.HostHeader.IsNull() {
		b.HostHeader = stringPtr(m.HostHeader.ValueString())
	}
	// HaproxyOptions would be parsed from the nested list here.
	return b
}

// fromAPI populates the Terraform model from the API model.
func (m *backendModel) fromAPI(b *client.Backend) {
	m.ID = types.Int64Value(int64(b.ID))
	m.Name = types.StringValue(b.Name)
	m.Mode = types.StringValue(b.Mode)
	m.Protocol = types.StringValue(b.Protocol)
	m.Algorithm = types.StringValue(b.Algorithm)
	m.StickySessions = types.BoolValue(b.StickySessions)
	if b.CookieName != nil {
		m.CookieName = types.StringValue(*b.CookieName)
	}
	if b.BalanceArgs != nil {
		m.BalanceArgs = types.StringValue(*b.BalanceArgs)
	}
	m.HealthCheckEnabled = types.BoolValue(b.HealthCheckEnabled)
	m.HealthCheckInterval = types.Int64Value(int64(b.HealthCheckInterval))
	m.HealthCheckURI = types.StringValue(b.HealthCheckURI)
	m.HealthCheckMethod = types.StringValue(b.HealthCheckMethod)
	m.Retries = types.Int64Value(int64(b.Retries))
	m.Redispatch = types.BoolValue(b.Redispatch)
	m.StickTable = types.BoolValue(b.StickTable)
	m.StickTableSize = types.StringValue(b.StickTableSize)
	m.StickTableExpire = types.StringValue(b.StickTableExpire)
	m.StickTableType = types.StringValue(b.StickTableType)
	if b.Resolvers != nil {
		m.Resolvers = types.StringValue(*b.Resolvers)
	}
	if b.HostHeader != nil {
		m.HostHeader = types.StringValue(*b.HostHeader)
	}
	m.RestoreClientIP = types.BoolValue(b.RestoreClientIP)
	m.ClientIPHeader = types.StringValue(b.ClientIPHeader)
}

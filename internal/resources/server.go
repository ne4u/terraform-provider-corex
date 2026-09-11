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
var _ resource.Resource = &ServerResource{}
var _ resource.ResourceWithImportState = &ServerResource{}

// ServerResource defines the corex_server resource.
type ServerResource struct {
	cli *client.Client
}

func NewServerResource() resource.Resource {
	return &ServerResource{}
}

// serverModel maps the Terraform schema to the API model.
type serverModel struct {
	ID           types.Int64  `tfsdk:"id"`
	BackendID    types.Int64  `tfsdk:"backend_id"`
	Name         types.String `tfsdk:"name"`
	Address      types.String `tfsdk:"address"`
	Port         types.Int64  `tfsdk:"port"`
	Weight       types.Int64  `tfsdk:"weight"`
	Maxconn      types.Int64  `tfsdk:"maxconn"`
	Check        types.Bool   `tfsdk:"check"`
	Backup       types.Bool   `tfsdk:"backup"`
	Inter        types.Int64  `tfsdk:"inter"`
	Rise         types.Int64  `tfsdk:"rise"`
	Fall         types.Int64  `tfsdk:"fall"`
	Slowstart    types.Int64  `tfsdk:"slowstart"`
	Maxqueue     types.Int64  `tfsdk:"maxqueue"`
	SSL          types.Bool   `tfsdk:"ssl"`
	Verify       types.String `tfsdk:"verify"`
	Verifyhost   types.String `tfsdk:"verifyhost"`
	Ciphers      types.String `tfsdk:"ciphers"`
	ALPN         types.String `tfsdk:"alpn"`
	SNI          types.String `tfsdk:"sni"`
	CheckSSL     types.Bool   `tfsdk:"check_ssl"`
	CheckSNI     types.String `tfsdk:"check_sni"`
	CheckPort    types.Int64  `tfsdk:"check_port"`
	SendProxy    types.Bool   `tfsdk:"send_proxy"`
	SendProxyV2  types.Bool   `tfsdk:"send_proxy_v2"`
	Resolve      types.Bool   `tfsdk:"resolve"`
	InitAddr     types.String `tfsdk:"init_addr"`
	AgentCheck   types.Bool   `tfsdk:"agent_check"`
	AgentPort    types.Int64  `tfsdk:"agent_port"`
	Track        types.String `tfsdk:"track"`
	Protocol     types.String `tfsdk:"protocol"`
}

func (r *ServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a backend server in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":             idAttr(),
			"backend_id":     intAttr("ID of the backend this server belongs to.", true),
			"name":           stringAttr("Server name.", true),
			"address":        stringAttr("Server address (IP or hostname).", true),
			"port":           intAttr("Server port.", true),
			"weight":         intAttr("Server weight for load balancing.", false),
			"maxconn":        intAttr("Maximum connections to this server.", false),
			"check":          boolAttr("Enable health checks on this server.", false),
			"backup":         boolAttr("Mark this server as a backup.", false),
			"inter":          intAttr("Health check interval in ms.", false),
			"rise":           intAttr("Number of successful checks before server is considered up.", false),
			"fall":           intAttr("Number of failed checks before server is considered down.", false),
			"slowstart":      intAttr("Slow start time in ms.", false),
			"maxqueue":       intAttr("Maximum queue length.", false),
			"ssl":            boolAttr("Enable SSL to backend.", false),
			"verify":         stringAttr("SSL verification level (none/required).", false),
			"verifyhost":     stringAttr("Hostname to verify against SSL certificate.", false),
			"ciphers":        stringAttr("SSL ciphers to use.", false),
			"alpn":           stringAttr("ALPN protocols to negotiate.", false),
			"sni":            stringAttr("SNI hostname for SSL.", false),
			"check_ssl":      boolAttr("Enable SSL for health checks.", false),
			"check_sni":      stringAttr("SNI hostname for health checks.", false),
			"check_port":     intAttr("Port to use for health checks.", false),
			"send_proxy":     boolAttr("Send PROXY protocol v1 to backend.", false),
			"send_proxy_v2":  boolAttr("Send PROXY protocol v2 to backend.", false),
			"resolve":        boolAttr("Resolve server address via DNS.", false),
			"init_addr":      stringAttr("Initial address before DNS resolution.", false),
			"agent_check":    boolAttr("Enable agent health checks.", false),
			"agent_port":     intAttr("Port for agent health checks.", false),
			"track":          stringAttr("Track another server's state.", false),
			"protocol":       stringAttr("Server protocol.", false),
		},
	}
}

func (r *ServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *ServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI()
	result, err := r.cli.AddServer(ctx, int(plan.BackendID.ValueInt64()), s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create server", err.Error())
		return
	}

	plan.fromAPI(result)
	// Preserve backend_id from the plan since the API result may not echo it.
	plan.BackendID = types.Int64Value(int64(s.BackendID))
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.cli.GetServer(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read server", err.Error())
		return
	}

	// Preserve backend_id from state since the API result may not echo it.
	backendID := state.BackendID
	state.fromAPI(s)
	state.BackendID = backendID
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s := plan.toAPI()
	result, err := r.cli.UpdateServer(ctx, int(plan.ID.ValueInt64()), s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update server", err.Error())
		return
	}

	plan.fromAPI(result)
	plan.BackendID = types.Int64Value(int64(s.BackendID))
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteServer(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete server", err.Error())
		return
	}
}

func (r *ServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by numeric ID only.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid server ID", fmt.Sprintf("Server must be imported by numeric ID, got %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *serverModel) toAPI() *client.Server {
	s := &client.Server{
		BackendID:    int(m.BackendID.ValueInt64()),
		Name:         m.Name.ValueString(),
		Address:      m.Address.ValueString(),
		Port:         int(m.Port.ValueInt64()),
		Weight:       int(m.Weight.ValueInt64()),
		Maxconn:      int(m.Maxconn.ValueInt64()),
		Check:        m.Check.ValueBool(),
		Backup:       m.Backup.ValueBool(),
		SSL:          m.SSL.ValueBool(),
		CheckSSL:     m.CheckSSL.ValueBool(),
		SendProxy:    m.SendProxy.ValueBool(),
		SendProxyV2:  m.SendProxyV2.ValueBool(),
		Resolve:      m.Resolve.ValueBool(),
		AgentCheck:   m.AgentCheck.ValueBool(),
		Protocol:     m.Protocol.ValueString(),
	}
	if !m.Inter.IsNull() {
		v := int(m.Inter.ValueInt64())
		s.Inter = &v
	}
	if !m.Rise.IsNull() {
		v := int(m.Rise.ValueInt64())
		s.Rise = &v
	}
	if !m.Fall.IsNull() {
		v := int(m.Fall.ValueInt64())
		s.Fall = &v
	}
	if !m.Slowstart.IsNull() {
		v := int(m.Slowstart.ValueInt64())
		s.Slowstart = &v
	}
	if !m.Maxqueue.IsNull() {
		v := int(m.Maxqueue.ValueInt64())
		s.Maxqueue = &v
	}
	if !m.Verify.IsNull() {
		s.Verify = stringPtr(m.Verify.ValueString())
	}
	if !m.Verifyhost.IsNull() {
		s.Verifyhost = stringPtr(m.Verifyhost.ValueString())
	}
	if !m.Ciphers.IsNull() {
		s.Ciphers = stringPtr(m.Ciphers.ValueString())
	}
	if !m.ALPN.IsNull() {
		s.ALPN = stringPtr(m.ALPN.ValueString())
	}
	if !m.SNI.IsNull() {
		s.SNI = stringPtr(m.SNI.ValueString())
	}
	if !m.CheckSNI.IsNull() {
		s.CheckSNI = stringPtr(m.CheckSNI.ValueString())
	}
	if !m.CheckPort.IsNull() {
		v := int(m.CheckPort.ValueInt64())
		s.CheckPort = &v
	}
	if !m.InitAddr.IsNull() {
		s.InitAddr = stringPtr(m.InitAddr.ValueString())
	}
	if !m.AgentPort.IsNull() {
		v := int(m.AgentPort.ValueInt64())
		s.AgentPort = &v
	}
	if !m.Track.IsNull() {
		s.Track = stringPtr(m.Track.ValueString())
	}
	return s
}

// fromAPI populates the Terraform model from the API model.
func (m *serverModel) fromAPI(s *client.Server) {
	m.ID = types.Int64Value(int64(s.ID))
	m.Name = types.StringValue(s.Name)
	m.Address = types.StringValue(s.Address)
	m.Port = types.Int64Value(int64(s.Port))
	m.Weight = types.Int64Value(int64(s.Weight))
	m.Maxconn = types.Int64Value(int64(s.Maxconn))
	m.Check = types.BoolValue(s.Check)
	m.Backup = types.BoolValue(s.Backup)
	m.SSL = types.BoolValue(s.SSL)
	m.CheckSSL = types.BoolValue(s.CheckSSL)
	m.SendProxy = types.BoolValue(s.SendProxy)
	m.SendProxyV2 = types.BoolValue(s.SendProxyV2)
	m.Resolve = types.BoolValue(s.Resolve)
	m.AgentCheck = types.BoolValue(s.AgentCheck)
	m.Protocol = types.StringValue(s.Protocol)
	if s.Inter != nil {
		m.Inter = types.Int64Value(int64(*s.Inter))
	}
	if s.Rise != nil {
		m.Rise = types.Int64Value(int64(*s.Rise))
	}
	if s.Fall != nil {
		m.Fall = types.Int64Value(int64(*s.Fall))
	}
	if s.Slowstart != nil {
		m.Slowstart = types.Int64Value(int64(*s.Slowstart))
	}
	if s.Maxqueue != nil {
		m.Maxqueue = types.Int64Value(int64(*s.Maxqueue))
	}
	if s.Verify != nil {
		m.Verify = types.StringValue(*s.Verify)
	}
	if s.Verifyhost != nil {
		m.Verifyhost = types.StringValue(*s.Verifyhost)
	}
	if s.Ciphers != nil {
		m.Ciphers = types.StringValue(*s.Ciphers)
	}
	if s.ALPN != nil {
		m.ALPN = types.StringValue(*s.ALPN)
	}
	if s.SNI != nil {
		m.SNI = types.StringValue(*s.SNI)
	}
	if s.CheckSNI != nil {
		m.CheckSNI = types.StringValue(*s.CheckSNI)
	}
	if s.CheckPort != nil {
		m.CheckPort = types.Int64Value(int64(*s.CheckPort))
	}
	if s.InitAddr != nil {
		m.InitAddr = types.StringValue(*s.InitAddr)
	}
	if s.AgentPort != nil {
		m.AgentPort = types.Int64Value(int64(*s.AgentPort))
	}
	if s.Track != nil {
		m.Track = types.StringValue(*s.Track)
	}
}

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
var _ resource.Resource = &ListenerResource{}
var _ resource.ResourceWithImportState = &ListenerResource{}

// ListenerResource defines the corex_listener resource.
type ListenerResource struct {
	cli *client.Client
}

func NewListenerResource() resource.Resource {
	return &ListenerResource{}
}

// listenerModel maps the Terraform schema to the API model.
type listenerModel struct {
	ID               types.Int64  `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	BindAddress      types.String `tfsdk:"bind_address"`
	BindPort         types.Int64  `tfsdk:"bind_port"`
	Mode             types.String `tfsdk:"mode"`
	Protocol         types.String `tfsdk:"protocol"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	SslEnabled       types.Bool   `tfsdk:"ssl_enabled"`
	CertificateID    types.Int64  `tfsdk:"certificate_id"`
	CertificateIDs   types.List   `tfsdk:"certificate_ids"`
	HTTP2            types.Bool   `tfsdk:"http2"`
	Quic             types.Bool   `tfsdk:"quic"`
	ALPN             types.String `tfsdk:"alpn"`
	ProxyProtocol    types.Bool   `tfsdk:"proxy_protocol"`
	ForceHTTPS       types.Bool   `tfsdk:"force_https"`
	DefaultBackendID types.Int64  `tfsdk:"default_backend_id"`
}

func (r *ListenerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_listener"
}

func (r *ListenerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a listener (frontend bind) in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id":                 idAttr(),
			"name":               stringAttr("Listener name.", true),
			"bind_address":       stringAttr("Address to bind to.", true),
			"bind_port":          intAttr("Port to bind to.", true),
			"mode":               stringAttr("HAProxy mode (tcp/http).", true),
			"protocol":           stringAttr("Listener protocol.", false),
			"enabled":            boolAttr("Whether this listener is enabled.", false),
			"ssl_enabled":        boolAttr("Enable SSL/TLS on this listener.", false),
			"certificate_id":     intAttr("ID of the primary certificate to use.", false),
			"certificate_ids":    listIntAttr("List of certificate IDs to use (SNI).", false),
			"http2":              boolAttr("Enable HTTP/2.", false),
			"quic":               boolAttr("Enable QUIC/HTTP3.", false),
			"alpn":               stringAttr("ALPN protocols to negotiate.", false),
			"proxy_protocol":     boolAttr("Enable PROXY protocol on the listener.", false),
			"force_https":        boolAttr("Force redirect to HTTPS.", false),
			"default_backend_id": intAttr("ID of the default backend for this listener.", false),
		},
	}
}

func (r *ListenerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *ListenerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan listenerModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	l := plan.toAPI(ctx)
	result, err := r.cli.CreateListener(ctx, l)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create listener", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ListenerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state listenerModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	l, err := r.cli.GetListener(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read listener", err.Error())
		return
	}

	state.fromAPI(ctx, l)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ListenerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan listenerModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	l := plan.toAPI(ctx)
	result, err := r.cli.UpdateListener(ctx, int(plan.ID.ValueInt64()), l)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update listener", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ListenerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state listenerModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.cli.DeleteListener(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete listener", err.Error())
		return
	}
}

func (r *ListenerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Try to import by name: list all listeners, find matching name.
	listeners, err := r.cli.ListListeners(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list listeners for import", err.Error())
		return
	}

	for _, l := range listeners {
		if l.Name == req.ID {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(l.ID)))...)
			return
		}
	}

	// Try numeric ID.
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Listener not found", fmt.Sprintf("No listener with name or ID %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.Int64Value(int64(id)))...)
}

// toAPI converts the Terraform model to the API model.
func (m *listenerModel) toAPI(ctx context.Context) *client.Listener {
	l := &client.Listener{
		Name:          m.Name.ValueString(),
		BindAddress:   m.BindAddress.ValueString(),
		BindPort:      int(m.BindPort.ValueInt64()),
		Mode:          m.Mode.ValueString(),
		Protocol:      m.Protocol.ValueString(),
		Enabled:       m.Enabled.ValueBool(),
		SslEnabled:    m.SslEnabled.ValueBool(),
		HTTP2:         m.HTTP2.ValueBool(),
		Quic:          m.Quic.ValueBool(),
		ProxyProtocol: m.ProxyProtocol.ValueBool(),
		ForceHTTPS:    m.ForceHTTPS.ValueBool(),
		CertificateIDs: intListToSlice(ctx, m.CertificateIDs),
	}
	if !m.CertificateID.IsNull() {
		v := int(m.CertificateID.ValueInt64())
		l.CertificateID = &v
	}
	if !m.ALPN.IsNull() {
		l.ALPN = stringPtr(m.ALPN.ValueString())
	}
	if !m.DefaultBackendID.IsNull() {
		v := int(m.DefaultBackendID.ValueInt64())
		l.DefaultBackendID = &v
	}
	return l
}

// fromAPI populates the Terraform model from the API model.
func (m *listenerModel) fromAPI(ctx context.Context, l *client.Listener) {
	m.ID = types.Int64Value(int64(l.ID))
	m.Name = types.StringValue(l.Name)
	m.BindAddress = types.StringValue(l.BindAddress)
	m.BindPort = types.Int64Value(int64(l.BindPort))
	m.Mode = types.StringValue(l.Mode)
	m.Protocol = types.StringValue(l.Protocol)
	m.Enabled = types.BoolValue(l.Enabled)
	m.SslEnabled = types.BoolValue(l.SslEnabled)
	m.HTTP2 = types.BoolValue(l.HTTP2)
	m.Quic = types.BoolValue(l.Quic)
	m.ProxyProtocol = types.BoolValue(l.ProxyProtocol)
	m.ForceHTTPS = types.BoolValue(l.ForceHTTPS)
	if l.CertificateID != nil {
		m.CertificateID = types.Int64Value(int64(*l.CertificateID))
	}
	m.CertificateIDs = sliceToIntList(ctx, l.CertificateIDs)
	if l.ALPN != nil {
		m.ALPN = types.StringValue(*l.ALPN)
	}
	if l.DefaultBackendID != nil {
		m.DefaultBackendID = types.Int64Value(int64(*l.DefaultBackendID))
	}
}

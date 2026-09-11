package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/ne4u/terraform-provider-corex/internal/client"
)

const haConfigID = "ha"

var _ resource.Resource = &HaConfigResource{}

// HaConfigResource defines the corex_ha_config resource (singleton).
type HaConfigResource struct {
	cli *client.Client
}

func NewHaConfigResource() resource.Resource {
	return &HaConfigResource{}
}

type haproxyInstanceModel struct {
	Name     types.String `tfsdk:"name"`
	URL      types.String `tfsdk:"url"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
}

type keepalivedModel struct {
	Vip             types.String `tfsdk:"vip"`
	VirtualRouterID types.Int64  `tfsdk:"virtual_router_id"`
	Priority        types.Int64  `tfsdk:"priority"`
	Interface       types.String `tfsdk:"interface"`
	AuthPassword    types.String `tfsdk:"auth_password"`
	PeerAddresses   types.List   `tfsdk:"peer_addresses"`
	AdvertInt       types.Int64  `tfsdk:"advert_int"`
	Preempt         types.Bool   `tfsdk:"preempt"`
	TrackScript     types.String `tfsdk:"track_script"`
}

type haConfigModel struct {
	ID                    types.String          `tfsdk:"id"`
	HaEnabled             types.Bool            `tfsdk:"ha_enabled"`
	SwarmMode             types.Bool            `tfsdk:"swarm_mode"`
	HaTopology            types.String          `tfsdk:"ha_topology"`
	HaproxyHaReplicas     types.Int64           `tfsdk:"haproxy_ha_replicas"`
	ValkeyHaReplicas      types.Int64           `tfsdk:"valkey_ha_replicas"`
	CorazaHaReplicas      types.Int64           `tfsdk:"coraza_ha_replicas"`
	HaproxyInstances      []haproxyInstanceModel `tfsdk:"haproxy_instances"`
	HaproxyPeerPort       types.Int64           `tfsdk:"haproxy_peer_port"`
	Keepalived            *keepalivedModel      `tfsdk:"keepalived"`
	ValkeySentinelEnabled types.Bool            `tfsdk:"valkey_sentinel_enabled"`
	ValkeySentinelHosts   types.List            `tfsdk:"valkey_sentinel_hosts"`
	ValkeySentinelService types.String          `tfsdk:"valkey_sentinel_service"`
}

func (r *HaConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ha_config"
}

func (r *HaConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the High Availability singleton configuration in coreX Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Fixed singleton identifier.",
				Computed:    true,
			},
			"ha_enabled":          boolAttr("Master toggle for HA.", false),
			"swarm_mode":          boolAttr("Whether swarm mode is active.", false),
			"ha_topology":         stringAttr("HA topology (single, active-passive, active-active).", false),
			"haproxy_ha_replicas": intAttr("Number of HAProxy HA replicas.", false),
			"valkey_ha_replicas":  intAttr("Number of Valkey HA replicas.", false),
			"coraza_ha_replicas":  intAttr("Number of Coraza HA replicas.", false),
			"haproxy_instances": schema.ListNestedAttribute{
				Description: "HAProxy instances with Data Plane API endpoints.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":     stringAttr("Instance name.", false),
						"url":      stringAttr("Data Plane API URL.", false),
						"user":     stringAttr("Data Plane API user.", false),
						"password": stringAttrSensitive("Data Plane API password.", false),
					},
				},
			},
			"haproxy_peer_port": intAttr("HAProxy peer port.", false),
			"keepalived": schema.SingleNestedAttribute{
				Description: "Keepalived configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"vip":               stringAttr("Virtual IP.", false),
					"virtual_router_id": intAttr("Virtual router ID.", false),
					"priority":          intAttr("Priority.", false),
					"interface":         stringAttr("Network interface.", false),
					"auth_password":     stringAttrSensitive("Keepalived auth password.", false),
					"peer_addresses":    listAttr("Peer addresses.", false),
					"advert_int":        intAttr("Advertisement interval (seconds).", false),
					"preempt":           boolAttr("Whether to preempt.", false),
					"track_script":      stringAttr("Track script name.", false),
				},
			},
			"valkey_sentinel_enabled": boolAttr("Enable Valkey sentinel.", false),
			"valkey_sentinel_hosts":   listAttr("Valkey sentinel hosts.", false),
			"valkey_sentinel_service": stringAttr("Valkey sentinel service name.", false),
		},
	}
}

func (r *HaConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.cli = getClient(req, resp)
}

func (r *HaConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan haConfigModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := plan.toAPI(ctx)
	result, err := r.cli.UpdateHaConfig(ctx, cfg)
	if err != nil {
		resp.Diagnostics.AddError("Failed to set ha config", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	plan.ID = types.StringValue(haConfigID)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *HaConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state haConfigModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, err := r.cli.GetHaConfig(ctx)
	if err != nil {
		if notFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read ha config", err.Error())
		return
	}

	state.fromAPI(ctx, cfg)
	state.ID = types.StringValue(haConfigID)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *HaConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan haConfigModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := plan.toAPI(ctx)
	result, err := r.cli.UpdateHaConfig(ctx, cfg)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update ha config", err.Error())
		return
	}

	plan.fromAPI(ctx, result)
	plan.ID = types.StringValue(haConfigID)
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *HaConfigResource) Delete(ctx context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Singleton: delete is a no-op.
	_ = ctx
}

func (m *haConfigModel) toAPI(ctx context.Context) *client.HaConfig {
	cfg := &client.HaConfig{
		HaEnabled:             m.HaEnabled.ValueBool(),
		SwarmMode:             m.SwarmMode.ValueBool(),
		HaTopology:            m.HaTopology.ValueString(),
		HaproxyHaReplicas:     int(m.HaproxyHaReplicas.ValueInt64()),
		ValkeyHaReplicas:      int(m.ValkeyHaReplicas.ValueInt64()),
		CorazaHaReplicas:      int(m.CorazaHaReplicas.ValueInt64()),
		HaproxyPeerPort:       int(m.HaproxyPeerPort.ValueInt64()),
		ValkeySentinelEnabled: m.ValkeySentinelEnabled.ValueBool(),
		ValkeySentinelHosts:   stringListToSlice(ctx, m.ValkeySentinelHosts),
		ValkeySentinelService: m.ValkeySentinelService.ValueString(),
	}
	instances := make([]client.HaproxyInstance, 0, len(m.HaproxyInstances))
	for _, i := range m.HaproxyInstances {
		inst := client.HaproxyInstance{
			Name: i.Name.ValueString(),
			URL:  i.URL.ValueString(),
		}
		if !i.User.IsNull() {
			inst.User = nilIfEmpty(i.User.ValueString())
		}
		if !i.Password.IsNull() {
			inst.Password = nilIfEmpty(i.Password.ValueString())
		}
		instances = append(instances, inst)
	}
	cfg.HaproxyInstances = instances
	if m.Keepalived != nil {
		k := &client.KeepalivedConfig{
			Vip:             m.Keepalived.Vip.ValueString(),
			VirtualRouterID: int(m.Keepalived.VirtualRouterID.ValueInt64()),
			Priority:        int(m.Keepalived.Priority.ValueInt64()),
			Interface:       m.Keepalived.Interface.ValueString(),
			PeerAddresses:   stringListToSlice(ctx, m.Keepalived.PeerAddresses),
			AdvertInt:       int(m.Keepalived.AdvertInt.ValueInt64()),
			Preempt:         m.Keepalived.Preempt.ValueBool(),
		}
		if !m.Keepalived.AuthPassword.IsNull() {
			k.AuthPassword = nilIfEmpty(m.Keepalived.AuthPassword.ValueString())
		}
		if !m.Keepalived.TrackScript.IsNull() {
			k.TrackScript = nilIfEmpty(m.Keepalived.TrackScript.ValueString())
		}
		cfg.Keepalived = *k
	}
	return cfg
}

func (m *haConfigModel) fromAPI(ctx context.Context, cfg *client.HaConfig) {
	m.HaEnabled = types.BoolValue(cfg.HaEnabled)
	m.SwarmMode = types.BoolValue(cfg.SwarmMode)
	m.HaTopology = types.StringValue(cfg.HaTopology)
	m.HaproxyHaReplicas = types.Int64Value(int64(cfg.HaproxyHaReplicas))
	m.ValkeyHaReplicas = types.Int64Value(int64(cfg.ValkeyHaReplicas))
	m.CorazaHaReplicas = types.Int64Value(int64(cfg.CorazaHaReplicas))
	m.HaproxyPeerPort = types.Int64Value(int64(cfg.HaproxyPeerPort))

	instances := make([]haproxyInstanceModel, 0, len(cfg.HaproxyInstances))
	for _, i := range cfg.HaproxyInstances {
		im := haproxyInstanceModel{
			Name: types.StringValue(i.Name),
			URL:  types.StringValue(i.URL),
		}
		if i.User != nil {
			im.User = types.StringValue(*i.User)
		} else {
			im.User = types.StringNull()
		}
		if i.Password != nil {
			im.Password = types.StringValue(*i.Password)
		} else {
			im.Password = types.StringNull()
		}
		instances = append(instances, im)
	}
	m.HaproxyInstances = instances

	k := &keepalivedModel{
		Vip:             types.StringValue(cfg.Keepalived.Vip),
		VirtualRouterID: types.Int64Value(int64(cfg.Keepalived.VirtualRouterID)),
		Priority:        types.Int64Value(int64(cfg.Keepalived.Priority)),
		Interface:       types.StringValue(cfg.Keepalived.Interface),
		PeerAddresses:   sliceToStringList(ctx, cfg.Keepalived.PeerAddresses),
		AdvertInt:       types.Int64Value(int64(cfg.Keepalived.AdvertInt)),
		Preempt:         types.BoolValue(cfg.Keepalived.Preempt),
	}
	if cfg.Keepalived.AuthPassword != nil {
		k.AuthPassword = types.StringValue(*cfg.Keepalived.AuthPassword)
	} else {
		k.AuthPassword = types.StringNull()
	}
	if cfg.Keepalived.TrackScript != nil {
		k.TrackScript = types.StringValue(*cfg.Keepalived.TrackScript)
	} else {
		k.TrackScript = types.StringNull()
	}
	m.Keepalived = k

	m.ValkeySentinelEnabled = types.BoolValue(cfg.ValkeySentinelEnabled)
	m.ValkeySentinelHosts = sliceToStringList(ctx, cfg.ValkeySentinelHosts)
	m.ValkeySentinelService = types.StringValue(cfg.ValkeySentinelService)
}

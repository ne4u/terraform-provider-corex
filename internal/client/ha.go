package client

import (
	"context"
)

// HaproxyInstance represents a single HAProxy instance with its Data Plane API endpoint.
type HaproxyInstance struct {
	Name     string  `json:"name"`
	URL      string  `json:"url"`
	User     *string `json:"user"`
	Password *string `json:"password"`
}

// KeepalivedConfig represents the keepalived configuration nested in HaConfig.
type KeepalivedConfig struct {
	Vip              string   `json:"vip"`
	VirtualRouterID  int      `json:"virtual_router_id"`
	Priority         int      `json:"priority"`
	Interface        string   `json:"interface"`
	AuthPassword     *string  `json:"auth_password"`
	PeerAddresses    []string `json:"peer_addresses"`
	AdvertInt        int      `json:"advert_int"`
	Preempt          bool     `json:"preempt"`
	TrackScript      *string  `json:"track_script"`
}

// HaConfig represents the High Availability configuration (singleton).
type HaConfig struct {
	HaEnabled              bool              `json:"ha_enabled"`
	SwarmMode              bool              `json:"swarm_mode"`
	HaTopology             string            `json:"ha_topology"`
	HaproxyHaReplicas      int               `json:"haproxy_ha_replicas"`
	ValkeyHaReplicas       int               `json:"valkey_ha_replicas"`
	CorazaHaReplicas       int               `json:"coraza_ha_replicas"`
	HaproxyInstances       []HaproxyInstance `json:"haproxy_instances"`
	HaproxyPeerPort        int               `json:"haproxy_peer_port"`
	Keepalived             KeepalivedConfig  `json:"keepalived"`
	ValkeySentinelEnabled  bool              `json:"valkey_sentinel_enabled"`
	ValkeySentinelHosts    []string          `json:"valkey_sentinel_hosts"`
	ValkeySentinelService  string            `json:"valkey_sentinel_service"`
}

// --- HA Config (singleton) ---

func (c *Client) GetHaConfig(ctx context.Context) (*HaConfig, error) {
	var cfg HaConfig
	err := c.Get(ctx, "/ha/config", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Client) UpdateHaConfig(ctx context.Context, cfg *HaConfig) (*HaConfig, error) {
	var result HaConfig
	err := c.UpdateWithApply(ctx, "/ha/config", cfg, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

package client

import (
	"context"
	"fmt"
	"time"
)

// Backend represents a coreX backend pool.
type Backend struct {
	ID                    int                    `json:"id"`
	Name                  string                 `json:"name"`
	Mode                  string                 `json:"mode"`
	Protocol              string                 `json:"protocol"`
	Algorithm             string                 `json:"algorithm"`
	StickySessions        bool                   `json:"sticky_sessions"`
	CookieName            *string                `json:"cookie_name"`
	BalanceArgs           *string                `json:"balance_args"`
	HealthCheckEnabled    bool                   `json:"health_check_enabled"`
	HealthCheckInterval   int                    `json:"health_check_interval"`
	HealthCheckURI        string                 `json:"health_check_uri"`
	HealthCheckMethod     string                 `json:"health_check_method"`
	HealthCheckExpectStatus *string              `json:"health_check_expect_status"`
	HealthCheckExpectBody *string                `json:"health_check_expect_body"`
	Retries               int                    `json:"retries"`
	Redispatch            bool                   `json:"redispatch"`
	TimeoutQueue          *int                   `json:"timeout_queue"`
	TimeoutCheck          *int                   `json:"timeout_check"`
	TimeoutTunnel         *int                   `json:"timeout_tunnel"`
	HTTPReuse             *string                `json:"http_reuse"`
	Fullconn              *int                   `json:"fullconn"`
	StickTable            bool                   `json:"stick_table"`
	StickTableSize        string                 `json:"stick_table_size"`
	StickTableExpire      string                 `json:"stick_table_expire"`
	StickTableType        string                 `json:"stick_table_type"`
	Resolvers             *string                `json:"resolvers"`
	HostHeader            *string                `json:"host_header"`
	RestoreClientIP       bool                   `json:"restore_client_ip"`
	ClientIPHeader        string                 `json:"client_ip_header"`
	FcgiAppID             *int                   `json:"fcgi_app_id"`
	Options               map[string]interface{} `json:"options"`
	HaproxyOptions        []HaproxyOption        `json:"haproxy_options"`
	Servers               []Server               `json:"servers"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
}

// HaproxyOption represents a raw HAProxy directive.
type HaproxyOption struct {
	Target    string `json:"target"`
	Directive string `json:"directive"`
	Value     string `json:"value"`
	Enabled   bool   `json:"enabled"`
}

// Server represents a backend server.
type Server struct {
	ID                  int                    `json:"id"`
	BackendID           int                    `json:"backend_id"`
	Name                string                 `json:"name"`
	Address             string                 `json:"address"`
	Port                int                    `json:"port"`
	Weight              int                    `json:"weight"`
	Maxconn             int                    `json:"maxconn"`
	Check               bool                   `json:"check"`
	Backup              bool                   `json:"backup"`
	Inter               *int                   `json:"inter"`
	Rise                *int                   `json:"rise"`
	Fall                *int                   `json:"fall"`
	Slowstart           *int                   `json:"slowstart"`
	Maxqueue            *int                   `json:"maxqueue"`
	SSL                 bool                   `json:"ssl"`
	Verify              *string                `json:"verify"`
	Verifyhost          *string                `json:"verifyhost"`
	Ciphers             *string                `json:"ciphers"`
	ALPN                *string                `json:"alpn"`
	SNI                 *string                `json:"sni"`
	CheckSSL            bool                   `json:"check_ssl"`
	CheckSNI            *string                `json:"check_sni"`
	CheckPort           *int                   `json:"check_port"`
	SendProxy           bool                   `json:"send_proxy"`
	SendProxyV2         bool                   `json:"send_proxy_v2"`
	Resolve             bool                   `json:"resolve"`
	InitAddr            *string                `json:"init_addr"`
	AgentCheck          bool                   `json:"agent_check"`
	AgentPort           *int                   `json:"agent_port"`
	Track               *string                `json:"track"`
	Protocol            string                 `json:"protocol"`
	Options             map[string]interface{} `json:"options"`
	CaCertificateID     *int                   `json:"ca_certificate_id"`
	ClientCertificateID *int                   `json:"client_certificate_id"`
}

// BackendRule represents a routing rule from a listener to a backend.
type BackendRule struct {
	ID            int                  `json:"id"`
	ListenerID    int                  `json:"listener_id"`
	BackendID     int                  `json:"backend_id"`
	Name          *string              `json:"name"`
	Priority      int                  `json:"priority"`
	ConditionType string               `json:"condition_type"`
	ConditionName *string              `json:"condition_name"`
	Operator      string               `json:"operator"`
	Value         *string              `json:"value"`
	Enabled       bool                 `json:"enabled"`
	Conditions    []BackendRuleCondition `json:"conditions"`
}

// BackendRuleCondition represents a condition within a backend rule.
type BackendRuleCondition struct {
	ConditionType string  `json:"condition_type"`
	ConditionName *string `json:"condition_name"`
	Operator      string  `json:"operator"`
	Value         *string `json:"value"`
	Join          string  `json:"join"`
}

// --- Backend CRUD ---

func (c *Client) ListBackends(ctx context.Context) ([]Backend, error) {
	var backends []Backend
	err := c.Get(ctx, "/backends", &backends)
	return backends, err
}

func (c *Client) GetBackend(ctx context.Context, id int) (*Backend, error) {
	var b Backend
	err := c.Get(ctx, fmt.Sprintf("/backends/%d", id), &b)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (c *Client) CreateBackend(ctx context.Context, b *Backend) (*Backend, error) {
	var result Backend
	err := c.CreateWithApply(ctx, "/backends", b, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateBackend(ctx context.Context, id int, b *Backend) (*Backend, error) {
	var result Backend
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/backends/%d", id), b, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteBackend(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/backends/%d", id))
}

// --- Server CRUD ---

func (c *Client) GetServer(ctx context.Context, id int) (*Server, error) {
	var s Server
	err := c.Get(ctx, fmt.Sprintf("/servers/%d", id), &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) AddServer(ctx context.Context, backendID int, s *Server) (*Server, error) {
	var result Server
	err := c.CreateWithApply(ctx, fmt.Sprintf("/backends/%d/servers", backendID), s, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateServer(ctx context.Context, id int, s *Server) (*Server, error) {
	var result Server
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/servers/%d", id), s, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteServer(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/servers/%d", id))
}

// --- Backend Rule CRUD ---

func (c *Client) ListBackendRules(ctx context.Context) ([]BackendRule, error) {
	var rules []BackendRule
	err := c.Get(ctx, "/backend-rules", &rules)
	return rules, err
}

func (c *Client) CreateBackendRule(ctx context.Context, r *BackendRule) (*BackendRule, error) {
	var result BackendRule
	err := c.CreateWithApply(ctx, "/backend-rules", r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetBackendRule(ctx context.Context, id int) (*BackendRule, error) {
	var r BackendRule
	err := c.Get(ctx, fmt.Sprintf("/backend-rules/%d", id), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) UpdateBackendRule(ctx context.Context, id int, r *BackendRule) (*BackendRule, error) {
	var result BackendRule
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/backend-rules/%d", id), r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteBackendRule(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/backend-rules/%d", id))
}

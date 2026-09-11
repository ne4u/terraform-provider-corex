package client

import (
	"context"
	"fmt"
	"time"
)

// Listener represents a coreX listener (frontend bind).
type Listener struct {
	ID              int                    `json:"id"`
	Name            string                 `json:"name"`
	BindAddress     string                 `json:"bind_address"`
	BindPort        int                    `json:"bind_port"`
	Mode            string                 `json:"mode"`
	Protocol        string                 `json:"protocol"`
	Enabled         bool                   `json:"enabled"`
	SslEnabled      bool                   `json:"ssl_enabled"`
	CertificateID   *int                   `json:"certificate_id"`
	CertificateIDs  []int                  `json:"certificate_ids"`
	HTTP2           bool                   `json:"http2"`
	Quic            bool                   `json:"quic"`
	ALPN            *string                `json:"alpn"`
	ProxyProtocol   bool                   `json:"proxy_protocol"`
	ForceHTTPS      bool                   `json:"force_https"`
	DefaultBackendID *int                  `json:"default_backend_id"`
	Options         map[string]interface{} `json:"options"`
	HaproxyOptions  []HaproxyOption        `json:"haproxy_options"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// --- Listener CRUD ---

// ListListeners returns all listeners.
func (c *Client) ListListeners(ctx context.Context) ([]Listener, error) {
	var listeners []Listener
	err := c.Get(ctx, "/listeners", &listeners)
	return listeners, err
}

// GetListener returns a single listener by ID.
func (c *Client) GetListener(ctx context.Context, id int) (*Listener, error) {
	var l Listener
	err := c.Get(ctx, fmt.Sprintf("/listeners/%d", id), &l)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

// CreateListener creates a new listener and applies the HAProxy config.
func (c *Client) CreateListener(ctx context.Context, l *Listener) (*Listener, error) {
	var result Listener
	err := c.CreateWithApply(ctx, "/listeners", l, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateListener updates an existing listener and applies the HAProxy config.
func (c *Client) UpdateListener(ctx context.Context, id int, l *Listener) (*Listener, error) {
	var result Listener
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/listeners/%d", id), l, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteListener deletes a listener and applies the HAProxy config.
func (c *Client) DeleteListener(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/listeners/%d", id))
}

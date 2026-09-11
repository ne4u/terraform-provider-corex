package client

import (
	"context"
	"fmt"
	"time"
)

// FcgiApp represents a coreX FastCGI application.
type FcgiApp struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Docroot     string     `json:"docroot"`
	Index       string     `json:"index"`
	PathInfo    string     `json:"path_info"`
	KeepConn    bool       `json:"keep_conn"`
	MpxsConns   bool       `json:"mpxs_conns"`
	MaxReqs     int        `json:"max_reqs"`
	Params      []FcgiParam `json:"params"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// FcgiParam represents a FastCGI parameter.
type FcgiParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// --- FcgiApp CRUD ---

// ListFcgiApps returns all FastCGI applications.
func (c *Client) ListFcgiApps(ctx context.Context) ([]FcgiApp, error) {
	var apps []FcgiApp
	err := c.Get(ctx, "/fcgi-apps", &apps)
	return apps, err
}

// GetFcgiApp returns a single FastCGI application by ID.
func (c *Client) GetFcgiApp(ctx context.Context, id int) (*FcgiApp, error) {
	var a FcgiApp
	err := c.Get(ctx, fmt.Sprintf("/fcgi-apps/%d", id), &a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// CreateFcgiApp creates a new FastCGI application and applies the HAProxy config.
func (c *Client) CreateFcgiApp(ctx context.Context, a *FcgiApp) (*FcgiApp, error) {
	var result FcgiApp
	err := c.CreateWithApply(ctx, "/fcgi-apps", a, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateFcgiApp updates an existing FastCGI application and applies the HAProxy config.
func (c *Client) UpdateFcgiApp(ctx context.Context, id int, a *FcgiApp) (*FcgiApp, error) {
	var result FcgiApp
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/fcgi-apps/%d", id), a, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteFcgiApp deletes a FastCGI application and applies the HAProxy config.
func (c *Client) DeleteFcgiApp(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/fcgi-apps/%d", id))
}

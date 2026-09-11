package client

import "context"

// withApply runs a mutation function and then applies the HAProxy config.
// This is used for HAProxy control-plane resources (backends, listeners, WAF, etc.).
// MCP gateway resources should NOT use this — they auto-regenerate their config server-side.
func (c *Client) withApply(ctx context.Context, mutate func() error) error {
	if err := mutate(); err != nil {
		return err
	}
	return c.ApplyConfig(ctx)
}

// CreateWithApply performs a POST with auto-apply.
func (c *Client) CreateWithApply(ctx context.Context, path string, body, out interface{}) error {
	return c.withApply(ctx, func() error {
		return c.do(ctx, "POST", path, body, out)
	})
}

// UpdateWithApply performs a PUT with auto-apply.
func (c *Client) UpdateWithApply(ctx context.Context, path string, body, out interface{}) error {
	return c.withApply(ctx, func() error {
		return c.do(ctx, "PUT", path, body, out)
	})
}

// DeleteWithApply performs a DELETE with auto-apply.
func (c *Client) DeleteWithApply(ctx context.Context, path string) error {
	return c.withApply(ctx, func() error {
		return c.do(ctx, "DELETE", path, nil, nil)
	})
}

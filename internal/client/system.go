package client

import (
	"context"
)

// --- Global HAProxy Options (singleton list, PUT with apply) ---

// GetGlobalOptions retrieves the global HAProxy options list.
func (c *Client) GetGlobalOptions(ctx context.Context) ([]HaproxyOption, error) {
	var opts []HaproxyOption
	err := c.Get(ctx, "/haproxy/global-options", &opts)
	return opts, err
}

// UpdateGlobalOptions replaces the global HAProxy options list and applies.
func (c *Client) UpdateGlobalOptions(ctx context.Context, opts []HaproxyOption) ([]HaproxyOption, error) {
	var result []HaproxyOption
	err := c.UpdateWithApply(ctx, "/haproxy/global-options", opts, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

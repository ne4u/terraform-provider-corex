package client

import (
	"context"
	"fmt"
)

// Setting represents a key/value setting in coreX Manager.
type Setting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetSetting retrieves a setting by key.
func (c *Client) GetSetting(ctx context.Context, key string) (*Setting, error) {
	var s Setting
	err := c.Get(ctx, fmt.Sprintf("/settings/%s", key), &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// SetSetting sets a setting value via PUT.
func (c *Client) SetSetting(ctx context.Context, key, value string) error {
	body := map[string]string{"value": value}
	return c.Put(ctx, fmt.Sprintf("/settings/%s", key), body, nil)
}

// DeleteSetting deletes a setting by key.
func (c *Client) DeleteSetting(ctx context.Context, key string) error {
	return c.Delete(ctx, fmt.Sprintf("/settings/%s", key))
}

// --- MaxMind license key ---

// MaxMindLicenseKey represents the MaxMind license key setting.
type MaxMindLicenseKey struct {
	Value string `json:"value"`
}

// GetMaxmindLicenseKey retrieves the MaxMind license key.
func (c *Client) GetMaxmindLicenseKey(ctx context.Context) (*MaxMindLicenseKey, error) {
	var mk MaxMindLicenseKey
	err := c.Get(ctx, "/settings/maxmind/license-key", &mk)
	if err != nil {
		return nil, err
	}
	return &mk, nil
}

// SetMaxmindLicenseKey sets the MaxMind license key via PUT.
func (c *Client) SetMaxmindLicenseKey(ctx context.Context, value string) error {
	body := map[string]string{"value": value}
	return c.Put(ctx, "/settings/maxmind/license-key", body, nil)
}

// DeleteMaxmindLicenseKey deletes the MaxMind license key.
func (c *Client) DeleteMaxmindLicenseKey(ctx context.Context) error {
	return c.Delete(ctx, "/settings/maxmind/license-key")
}

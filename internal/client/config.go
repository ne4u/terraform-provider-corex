package client

import (
	"context"
	"time"
)

// ConfigStatus represents the current status of the coreX configuration.
type ConfigStatus struct {
	Status         string     `json:"status"`
	PendingChanges int        `json:"pending_changes"`
	LastAppliedAt  *time.Time `json:"last_applied_at"`
}

// ConfigSnapshot represents a saved snapshot of the coreX configuration.
type ConfigSnapshot struct {
	ID          int       `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
}

// GetConfigStatus retrieves the current configuration status.
func (c *Client) GetConfigStatus(ctx context.Context) (*ConfigStatus, error) {
	var s ConfigStatus
	err := c.Get(ctx, "/config/status", &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetConfigPreview retrieves the configuration preview as a string.
func (c *Client) GetConfigPreview(ctx context.Context) (string, error) {
	var resp struct {
		Preview string `json:"preview"`
	}
	if err := c.Get(ctx, "/config/preview", &resp); err != nil {
		return "", err
	}
	return resp.Preview, nil
}

// GetConfigDiff retrieves the configuration diff as a string.
func (c *Client) GetConfigDiff(ctx context.Context) (string, error) {
	var resp struct {
		Diff string `json:"diff"`
	}
	if err := c.Get(ctx, "/config/diff", &resp); err != nil {
		return "", err
	}
	return resp.Diff, nil
}

// ListConfigSnapshots lists all configuration snapshots.
func (c *Client) ListConfigSnapshots(ctx context.Context) ([]ConfigSnapshot, error) {
	var snaps []ConfigSnapshot
	err := c.Get(ctx, "/config/snapshots", &snaps)
	return snaps, err
}

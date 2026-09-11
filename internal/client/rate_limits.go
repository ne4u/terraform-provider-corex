package client

import (
	"context"
	"fmt"
	"time"
)

// RateLimit represents a coreX rate limit.
type RateLimit struct {
	ID             int       `json:"id"`
	ListenerID     int       `json:"listener_id"`
	Name           string    `json:"name"`
	LimitType      string    `json:"limit_type"`
	Events         int       `json:"events"`
	WindowSeconds  int       `json:"window_seconds"`
	Burst          int       `json:"burst"`
	Action         string    `json:"action"`
	DurationSeconds int      `json:"duration_seconds"`
	Expression     string    `json:"expression"`
	RateKey        string    `json:"rate_key"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// --- RateLimit CRUD ---

// ListRateLimits returns all rate limits.
func (c *Client) ListRateLimits(ctx context.Context) ([]RateLimit, error) {
	var limits []RateLimit
	err := c.Get(ctx, "/rate-limits", &limits)
	return limits, err
}

// GetRateLimit returns a single rate limit by ID.
func (c *Client) GetRateLimit(ctx context.Context, id int) (*RateLimit, error) {
	var r RateLimit
	err := c.Get(ctx, fmt.Sprintf("/rate-limits/%d", id), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// CreateRateLimit creates a new rate limit and applies the HAProxy config.
func (c *Client) CreateRateLimit(ctx context.Context, r *RateLimit) (*RateLimit, error) {
	var result RateLimit
	err := c.CreateWithApply(ctx, "/rate-limits", r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateRateLimit updates an existing rate limit and applies the HAProxy config.
func (c *Client) UpdateRateLimit(ctx context.Context, id int, r *RateLimit) (*RateLimit, error) {
	var result RateLimit
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/rate-limits/%d", id), r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteRateLimit deletes a rate limit and applies the HAProxy config.
func (c *Client) DeleteRateLimit(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/rate-limits/%d", id))
}

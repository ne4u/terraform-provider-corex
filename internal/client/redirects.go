package client

import (
	"context"
	"fmt"
	"time"
)

// Redirect represents a coreX redirect rule.
type Redirect struct {
	ID            int       `json:"id"`
	ListenerID    int       `json:"listener_id"`
	ListenerIDs   []int     `json:"listener_ids"`
	Priority      int       `json:"priority"`
	Name          string    `json:"name"`
	Source        string    `json:"source"`
	Target        string    `json:"target"`
	Type          string    `json:"type"`
	Code          int       `json:"code"`
	PreserveQuery bool      `json:"preserve_query"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Rewrite represents a coreX URL rewrite rule.
type Rewrite struct {
	ID           int       `json:"id"`
	ListenerID   int       `json:"listener_id"`
	ListenerIDs  []int     `json:"listener_ids"`
	Priority     int       `json:"priority"`
	Name         string    `json:"name"`
	HostMatch    string    `json:"host_match"`
	SourceRegex  string    `json:"source_regex"`
	Target       string    `json:"target"`
	Type         string    `json:"type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// --- Redirect CRUD ---

// ListRedirects returns all redirect rules.
func (c *Client) ListRedirects(ctx context.Context) ([]Redirect, error) {
	var redirects []Redirect
	err := c.Get(ctx, "/redirects", &redirects)
	return redirects, err
}

// GetRedirect returns a single redirect rule by ID.
func (c *Client) GetRedirect(ctx context.Context, id int) (*Redirect, error) {
	var r Redirect
	err := c.Get(ctx, fmt.Sprintf("/redirects/%d", id), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// CreateRedirect creates a new redirect rule and applies the HAProxy config.
func (c *Client) CreateRedirect(ctx context.Context, r *Redirect) (*Redirect, error) {
	var result Redirect
	err := c.CreateWithApply(ctx, "/redirects", r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateRedirect updates an existing redirect rule and applies the HAProxy config.
func (c *Client) UpdateRedirect(ctx context.Context, id int, r *Redirect) (*Redirect, error) {
	var result Redirect
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/redirects/%d", id), r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteRedirect deletes a redirect rule and applies the HAProxy config.
func (c *Client) DeleteRedirect(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/redirects/%d", id))
}

// --- Rewrite CRUD ---

// ListRewrites returns all rewrite rules.
func (c *Client) ListRewrites(ctx context.Context) ([]Rewrite, error) {
	var rewrites []Rewrite
	err := c.Get(ctx, "/rewrites", &rewrites)
	return rewrites, err
}

// GetRewrite returns a single rewrite rule by ID.
func (c *Client) GetRewrite(ctx context.Context, id int) (*Rewrite, error) {
	var r Rewrite
	err := c.Get(ctx, fmt.Sprintf("/rewrites/%d", id), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// CreateRewrite creates a new rewrite rule and applies the HAProxy config.
func (c *Client) CreateRewrite(ctx context.Context, r *Rewrite) (*Rewrite, error) {
	var result Rewrite
	err := c.CreateWithApply(ctx, "/rewrites", r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateRewrite updates an existing rewrite rule and applies the HAProxy config.
func (c *Client) UpdateRewrite(ctx context.Context, id int, r *Rewrite) (*Rewrite, error) {
	var result Rewrite
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/rewrites/%d", id), r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteRewrite deletes a rewrite rule and applies the HAProxy config.
func (c *Client) DeleteRewrite(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/rewrites/%d", id))
}

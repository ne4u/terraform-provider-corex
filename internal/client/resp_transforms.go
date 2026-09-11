package client

import (
	"context"
	"fmt"
	"time"
)

// ResponseTransform represents a coreX response transformation rule.
type ResponseTransform struct {
	ID             int       `json:"id"`
	BackendID      int       `json:"backend_id"`
	BackendIDs     []int     `json:"backend_ids"`
	Priority       int       `json:"priority"`
	Name           string    `json:"name"`
	Enabled        bool      `json:"enabled"`
	TransformType  string    `json:"transform_type"`
	FindRegex      string    `json:"find_regex"`
	ReplaceString  string    `json:"replace_string"`
	InjectHeader   string    `json:"inject_header"`
	InjectValue    string    `json:"inject_value"`
	MaskPattern    string    `json:"mask_pattern"`
	MaskReplacement string   `json:"mask_replacement"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// --- ResponseTransform CRUD ---

// ListResponseTransforms returns all response transform rules.
func (c *Client) ListResponseTransforms(ctx context.Context) ([]ResponseTransform, error) {
	var transforms []ResponseTransform
	err := c.Get(ctx, "/resp-transforms", &transforms)
	return transforms, err
}

// GetResponseTransform returns a single response transform rule by ID.
func (c *Client) GetResponseTransform(ctx context.Context, id int) (*ResponseTransform, error) {
	var t ResponseTransform
	err := c.Get(ctx, fmt.Sprintf("/resp-transforms/%d", id), &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateResponseTransform creates a new response transform rule and applies the HAProxy config.
func (c *Client) CreateResponseTransform(ctx context.Context, t *ResponseTransform) (*ResponseTransform, error) {
	var result ResponseTransform
	err := c.CreateWithApply(ctx, "/resp-transforms", t, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateResponseTransform updates an existing response transform rule and applies the HAProxy config.
func (c *Client) UpdateResponseTransform(ctx context.Context, id int, t *ResponseTransform) (*ResponseTransform, error) {
	var result ResponseTransform
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/resp-transforms/%d", id), t, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteResponseTransform deletes a response transform rule and applies the HAProxy config.
func (c *Client) DeleteResponseTransform(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/resp-transforms/%d", id))
}

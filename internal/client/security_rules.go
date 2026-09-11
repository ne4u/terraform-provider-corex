package client

import (
	"context"
	"fmt"
	"time"
)

// SecurityRule represents a coreX security rule.
type SecurityRule struct {
	ID             int            `json:"id"`
	Name           string         `json:"name"`
	Enabled        bool           `json:"enabled"`
	ListenerIDs    []int          `json:"listener_ids"`
	Expression     string         `json:"expression"`
	Action         string         `json:"action"`
	Log            bool           `json:"log"`
	NoLog          bool           `json:"no_log"`
	StatusCode     *int           `json:"status_code"`
	RedirectURL    *string        `json:"redirect_url"`
	RedirectCode   *int           `json:"redirect_code"`
	ErrorPageID    *int           `json:"error_page_id"`
	Priority       int            `json:"priority"`
	ExpressionAST  map[string]interface{} `json:"expression_ast"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// --- SecurityRule CRUD ---

// ListSecurityRules returns all security rules.
func (c *Client) ListSecurityRules(ctx context.Context) ([]SecurityRule, error) {
	var rules []SecurityRule
	err := c.Get(ctx, "/security-rules", &rules)
	return rules, err
}

// GetSecurityRule returns a single security rule by ID.
func (c *Client) GetSecurityRule(ctx context.Context, id int) (*SecurityRule, error) {
	var r SecurityRule
	err := c.Get(ctx, fmt.Sprintf("/security-rules/%d", id), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// CreateSecurityRule creates a new security rule and applies the HAProxy config.
func (c *Client) CreateSecurityRule(ctx context.Context, r *SecurityRule) (*SecurityRule, error) {
	var result SecurityRule
	err := c.CreateWithApply(ctx, "/security-rules", r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateSecurityRule updates an existing security rule and applies the HAProxy config.
func (c *Client) UpdateSecurityRule(ctx context.Context, id int, r *SecurityRule) (*SecurityRule, error) {
	var result SecurityRule
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/security-rules/%d", id), r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteSecurityRule deletes a security rule and applies the HAProxy config.
func (c *Client) DeleteSecurityRule(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/security-rules/%d", id))
}

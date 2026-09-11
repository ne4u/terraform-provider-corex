package client

import (
	"context"
	"fmt"
	"time"
)

// RiskRuleset represents a named collection of risk rules.
type RiskRuleset struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Enabled     bool      `json:"enabled"`
	Slug        string    `json:"slug"`
	Priority    int       `json:"priority"`
	RuleCount   int       `json:"rule_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RiskRule represents a single risk-scoring security rule.
type RiskRule struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Enabled     bool      `json:"enabled"`
	ListenerIDs []int     `json:"listener_ids"`
	Expression  string    `json:"expression"`
	Points      int       `json:"points"`
	Category    *string   `json:"category"`
	Log         bool      `json:"log"`
	RulesetID   int       `json:"ruleset_id"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// --- Risk Ruleset CRUD ---

func (c *Client) ListRiskRulesets(ctx context.Context) ([]RiskRuleset, error) {
	var sets []RiskRuleset
	err := c.Get(ctx, "/risk-rulesets", &sets)
	return sets, err
}

func (c *Client) GetRiskRuleset(ctx context.Context, id int) (*RiskRuleset, error) {
	var rs RiskRuleset
	err := c.Get(ctx, fmt.Sprintf("/risk-rulesets/%d", id), &rs)
	if err != nil {
		return nil, err
	}
	return &rs, nil
}

func (c *Client) CreateRiskRuleset(ctx context.Context, rs *RiskRuleset) (*RiskRuleset, error) {
	var result RiskRuleset
	err := c.CreateWithApply(ctx, "/risk-rulesets", rs, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateRiskRuleset(ctx context.Context, id int, rs *RiskRuleset) (*RiskRuleset, error) {
	var result RiskRuleset
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/risk-rulesets/%d", id), rs, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteRiskRuleset(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/risk-rulesets/%d", id))
}

// --- Risk Rule CRUD ---

func (c *Client) ListRiskRules(ctx context.Context) ([]RiskRule, error) {
	var rules []RiskRule
	err := c.Get(ctx, "/risk-rules", &rules)
	return rules, err
}

func (c *Client) GetRiskRule(ctx context.Context, id int) (*RiskRule, error) {
	var r RiskRule
	err := c.Get(ctx, fmt.Sprintf("/risk-rules/%d", id), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) CreateRiskRule(ctx context.Context, r *RiskRule) (*RiskRule, error) {
	var result RiskRule
	err := c.CreateWithApply(ctx, "/risk-rules", r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateRiskRule(ctx context.Context, id int, r *RiskRule) (*RiskRule, error) {
	var result RiskRule
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/risk-rules/%d", id), r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteRiskRule(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/risk-rules/%d", id))
}

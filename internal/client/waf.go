package client

import (
	"context"
	"fmt"
	"time"
)

// WafRule represents a coreX WAF rule.
type WafRule struct {
	ID                        int       `json:"id"`
	ListenerID                int       `json:"listener_id"`
	BackendID                 int       `json:"backend_id"`
	Name                      string    `json:"name"`
	Enabled                   bool      `json:"enabled"`
	RuleSet                   string    `json:"rule_set"`
	RuleSetVersion            string    `json:"rule_set_version"`
	RuleSetURL                string    `json:"rule_set_url"`
	RuleSetSHA256             string    `json:"rule_set_sha256"`
	RuleSetAutoUpdate         bool      `json:"rule_set_auto_update"`
	RuleSetUpdateIntervalHours int      `json:"rule_set_update_interval_hours"`
	Engine                    string    `json:"engine"`
	ParanoiaLevel             int       `json:"paranoia_level"`
	InboundAnomalyThreshold   int       `json:"inbound_anomaly_threshold"`
	OutboundAnomalyThreshold  int       `json:"outbound_anomaly_threshold"`
	Action                    string    `json:"action"`
	RedirectURL               string    `json:"redirect_url"`
	StatusCode                int       `json:"status_code"`
	PathPattern               string    `json:"path_pattern"`
	HTTPMethods               []string  `json:"http_methods"`
	FailOpen                  bool      `json:"fail_open"`
	SiemIntegrationID         *int      `json:"siem_integration_id"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

// WafException represents a coreX WAF exception.
type WafException struct {
	ID         int       `json:"id"`
	WafRuleID  int       `json:"waf_rule_id"`
	Name       string    `json:"name"`
	RuleID     *string   `json:"rule_id"`
	RuleTag    *string   `json:"rule_tag"`
	RuleMsg    *string   `json:"rule_msg"`
	Zone       string    `json:"zone"`
	Variable   *string   `json:"variable"`
	Matcher    string    `json:"matcher"`
	Value      string    `json:"value"`
	Description string   `json:"description"`
	Action     string    `json:"action"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// --- WafRule CRUD ---

// ListWafRules returns all WAF rules.
func (c *Client) ListWafRules(ctx context.Context) ([]WafRule, error) {
	var rules []WafRule
	err := c.Get(ctx, "/waf-rules", &rules)
	return rules, err
}

// GetWafRule returns a single WAF rule by ID.
func (c *Client) GetWafRule(ctx context.Context, id int) (*WafRule, error) {
	var r WafRule
	err := c.Get(ctx, fmt.Sprintf("/waf-rules/%d", id), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// CreateWafRule creates a new WAF rule and applies the HAProxy config.
func (c *Client) CreateWafRule(ctx context.Context, r *WafRule) (*WafRule, error) {
	var result WafRule
	err := c.CreateWithApply(ctx, "/waf-rules", r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateWafRule updates an existing WAF rule and applies the HAProxy config.
func (c *Client) UpdateWafRule(ctx context.Context, id int, r *WafRule) (*WafRule, error) {
	var result WafRule
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/waf-rules/%d", id), r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteWafRule deletes a WAF rule and applies the HAProxy config.
func (c *Client) DeleteWafRule(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/waf-rules/%d", id))
}

// --- WafException CRUD ---

// ListWafExceptions returns all WAF exceptions.
func (c *Client) ListWafExceptions(ctx context.Context) ([]WafException, error) {
	var exceptions []WafException
	err := c.Get(ctx, "/waf-exceptions", &exceptions)
	return exceptions, err
}

// GetWafException returns a single WAF exception by ID.
func (c *Client) GetWafException(ctx context.Context, id int) (*WafException, error) {
	var e WafException
	err := c.Get(ctx, fmt.Sprintf("/waf-exceptions/%d", id), &e)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// CreateWafException creates a new WAF exception and applies the HAProxy config.
func (c *Client) CreateWafException(ctx context.Context, e *WafException) (*WafException, error) {
	var result WafException
	err := c.CreateWithApply(ctx, "/waf-exceptions", e, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateWafException updates an existing WAF exception and applies the HAProxy config.
func (c *Client) UpdateWafException(ctx context.Context, id int, e *WafException) (*WafException, error) {
	var result WafException
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/waf-exceptions/%d", id), e, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteWafException deletes a WAF exception and applies the HAProxy config.
func (c *Client) DeleteWafException(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/waf-exceptions/%d", id))
}

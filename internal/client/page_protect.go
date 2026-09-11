package client

import (
	"context"
	"fmt"
	"time"
)

// PageProtectPolicy represents a page protection policy.
type PageProtectPolicy struct {
	ID                int                `json:"id"`
	Name              string             `json:"name"`
	Enabled           bool               `json:"enabled"`
	BackendIDs        []int              `json:"backend_ids"`
	Mode              string             `json:"mode"`
	SampleRatePercent int                `json:"sample_rate_percent"`
	ReportPath        string             `json:"report_path"`
	Directives        map[string][]string `json:"directives"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// PageProtectScript represents a page protection script resource.
type PageProtectScript struct {
	ID           int       `json:"id"`
	URL          string    `json:"url"`
	ResourceType string    `json:"resource_type"`
	Notes        string    `json:"notes"`
	FetchMethod  string    `json:"fetch_method"`
	Ignored      bool      `json:"ignored"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PageProtectSettings represents the page-protect singleton settings.
type PageProtectSettings struct {
	MonitoringEnabled          bool     `json:"monitoring_enabled"`
	ChangeDetectionEnabled     bool     `json:"change_detection_enabled"`
	ChangeDetectionIntervalHours int    `json:"change_detection_interval_hours"`
	ReportRetentionDays        int      `json:"report_retention_days"`
	ReportPath                 string   `json:"report_path"`
	BeaconInjectionEnabled     bool     `json:"beacon_injection_enabled"`
	BeaconTrustEnabled         bool     `json:"beacon_trust_enabled"`
	BeaconPaths                []string `json:"beacon_paths"`
	BeaconContentTypes         []string `json:"beacon_content_types"`
	BeaconPatterns             []string `json:"beacon_patterns"`
	BackendIDs                 []int    `json:"backend_ids"`
	AutoPruneStaleDays         int      `json:"auto_prune_stale_days"`
}

// --- Page Protect Policy CRUD ---

func (c *Client) ListPageProtectPolicies(ctx context.Context) ([]PageProtectPolicy, error) {
	var policies []PageProtectPolicy
	err := c.Get(ctx, "/page-protect/policies", &policies)
	return policies, err
}

func (c *Client) GetPageProtectPolicy(ctx context.Context, id int) (*PageProtectPolicy, error) {
	var p PageProtectPolicy
	err := c.Get(ctx, fmt.Sprintf("/page-protect/policies/%d", id), &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) CreatePageProtectPolicy(ctx context.Context, p *PageProtectPolicy) (*PageProtectPolicy, error) {
	var result PageProtectPolicy
	err := c.CreateWithApply(ctx, "/page-protect/policies", p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdatePageProtectPolicy(ctx context.Context, id int, p *PageProtectPolicy) (*PageProtectPolicy, error) {
	var result PageProtectPolicy
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/page-protect/policies/%d", id), p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeletePageProtectPolicy(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/page-protect/policies/%d", id))
}

// --- Page Protect Script CRUD ---

func (c *Client) ListPageProtectScripts(ctx context.Context) ([]PageProtectScript, error) {
	var scripts []PageProtectScript
	err := c.Get(ctx, "/page-protect/scripts", &scripts)
	return scripts, err
}

func (c *Client) GetPageProtectScript(ctx context.Context, id int) (*PageProtectScript, error) {
	var s PageProtectScript
	err := c.Get(ctx, fmt.Sprintf("/page-protect/scripts/%d", id), &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) CreatePageProtectScript(ctx context.Context, s *PageProtectScript) (*PageProtectScript, error) {
	var result PageProtectScript
	err := c.CreateWithApply(ctx, "/page-protect/scripts", s, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdatePageProtectScript(ctx context.Context, id int, s *PageProtectScript) (*PageProtectScript, error) {
	var result PageProtectScript
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/page-protect/scripts/%d", id), s, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeletePageProtectScript(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/page-protect/scripts/%d", id))
}

// --- Page Protect Settings (singleton, PUT) ---

func (c *Client) GetPageProtectSettings(ctx context.Context) (*PageProtectSettings, error) {
	var s PageProtectSettings
	err := c.Get(ctx, "/page-protect/settings", &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) UpdatePageProtectSettings(ctx context.Context, s *PageProtectSettings) (*PageProtectSettings, error) {
	var result PageProtectSettings
	err := c.Put(ctx, "/page-protect/settings", s, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

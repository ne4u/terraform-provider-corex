package client

import (
	"context"
	"fmt"
	"time"
)

// CacheConfig represents a cache configuration for a backend.
type CacheConfig struct {
	ID                  int      `json:"id"`
	BackendID           int      `json:"backend_id"`
	HaproxyEnabled      bool     `json:"haproxy_enabled"`
	HaproxyCacheSize    int      `json:"haproxy_cache_size"`
	HaproxyCacheMaxAge  int      `json:"haproxy_cache_max_age"`
	HaproxyCacheVary    []string `json:"haproxy_cache_vary"`
	DiskCacheEnabled    bool     `json:"disk_cache_enabled"`
	DiskCachePath       string   `json:"disk_cache_path"`
	DiskCacheMaxSize    int      `json:"disk_cache_max_size"`
	DiskCacheMaxAge     int      `json:"disk_cache_max_age"`
	RFC7234Compliance   bool     `json:"rfc7234_compliance"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// CacheRule represents a cache rule attached to a backend's cache config.
type CacheRule struct {
	ID            int       `json:"id"`
	CacheConfigID int       `json:"cache_config_id"`
	MatchType     string    `json:"match_type"`
	Pattern       string    `json:"pattern"`
	Action        string    `json:"action"`
	Tier          string    `json:"tier"`
	Enabled       bool      `json:"enabled"`
	Priority      int       `json:"priority"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// --- CacheConfig CRUD ---

// GetCacheConfig retrieves the cache configuration for a backend.
func (c *Client) GetCacheConfig(ctx context.Context, backendID int) (*CacheConfig, error) {
	var cc CacheConfig
	err := c.Get(ctx, fmt.Sprintf("/cache/configs/%d", backendID), &cc)
	if err != nil {
		return nil, err
	}
	return &cc, nil
}

// CreateCacheConfig creates a cache configuration for a backend.
func (c *Client) CreateCacheConfig(ctx context.Context, cc *CacheConfig) (*CacheConfig, error) {
	var result CacheConfig
	err := c.CreateWithApply(ctx, "/cache/configs", cc, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateCacheConfig updates a cache configuration for a backend.
func (c *Client) UpdateCacheConfig(ctx context.Context, backendID int, cc *CacheConfig) (*CacheConfig, error) {
	var result CacheConfig
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/cache/configs/%d", backendID), cc, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteCacheConfig deletes the cache configuration for a backend.
func (c *Client) DeleteCacheConfig(ctx context.Context, backendID int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/cache/configs/%d", backendID))
}

// --- CacheRule CRUD ---

// ListCacheRules lists all cache rules for a backend's cache config.
func (c *Client) ListCacheRules(ctx context.Context, backendID int) ([]CacheRule, error) {
	var rules []CacheRule
	err := c.Get(ctx, fmt.Sprintf("/cache/configs/%d/rules", backendID), &rules)
	return rules, err
}

// CreateCacheRule creates a cache rule for a backend's cache config.
func (c *Client) CreateCacheRule(ctx context.Context, backendID int, r *CacheRule) (*CacheRule, error) {
	var result CacheRule
	err := c.CreateWithApply(ctx, fmt.Sprintf("/cache/configs/%d/rules", backendID), r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateCacheRule updates a cache rule by ID.
func (c *Client) UpdateCacheRule(ctx context.Context, id int, r *CacheRule) (*CacheRule, error) {
	var result CacheRule
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/cache/rules/%d", id), r, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteCacheRule deletes a cache rule by ID.
func (c *Client) DeleteCacheRule(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/cache/rules/%d", id))
}

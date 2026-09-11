package client

import (
	"context"
	"fmt"
	"time"
)

// CipherSuite represents a TLS cipher suite configuration in coreX Manager.
type CipherSuite struct {
	ID                   int      `json:"id"`
	Name                 string   `json:"name"`
	Baseline             string   `json:"baseline"`
	Ciphers              string   `json:"ciphers"`
	TlsOptions           []string `json:"tls_options"`
	MinTlsVersion        string   `json:"min_tls_version"`
	QuantumSafe          bool     `json:"quantum_safe"`
	HstsEnabled          bool     `json:"hsts_enabled"`
	HstsMaxAge           int      `json:"hsts_max_age"`
	HstsIncludeSubdomains bool     `json:"hsts_include_subdomains"`
	HstsPreload          bool     `json:"hsts_preload"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// --- CipherSuite CRUD ---

func (c *Client) ListCipherSuites(ctx context.Context) ([]CipherSuite, error) {
	var suites []CipherSuite
	err := c.Get(ctx, "/cipher-suites", &suites)
	return suites, err
}

func (c *Client) GetCipherSuite(ctx context.Context, id int) (*CipherSuite, error) {
	var suite CipherSuite
	err := c.Get(ctx, fmt.Sprintf("/cipher-suites/%d", id), &suite)
	if err != nil {
		return nil, err
	}
	return &suite, nil
}

func (c *Client) CreateCipherSuite(ctx context.Context, suite *CipherSuite) (*CipherSuite, error) {
	var result CipherSuite
	err := c.CreateWithApply(ctx, "/cipher-suites", suite, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateCipherSuite(ctx context.Context, id int, suite *CipherSuite) (*CipherSuite, error) {
	var result CipherSuite
	// Per spec, UpdateCipherSuite uses CreateWithApply semantics (PUT with apply).
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/cipher-suites/%d", id), suite, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteCipherSuite(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/cipher-suites/%d", id))
}

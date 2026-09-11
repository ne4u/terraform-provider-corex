package client

import (
	"context"
	"fmt"
	"time"
)

// Certificate represents a certificate managed by coreX Manager.
type Certificate struct {
	ID             int               `json:"id"`
	Name           string            `json:"name"`
	Domain         string            `json:"domain"`
	Kind           string            `json:"kind"`
	Provider       string            `json:"provider"`
	Email          string            `json:"email"`
	IsWildcard     bool              `json:"is_wildcard"`
	AutoRenew      bool              `json:"auto_renew"`
	KeyType        string            `json:"key_type"`
	AcmeChallenge  string            `json:"acme_challenge"`
	AcmeCA         string            `json:"acme_ca"`
	DnsProvider    string            `json:"dns_provider"`
	DnsCredentials map[string]string `json:"dns_credentials"`
	Fullchain      string            `json:"fullchain"`
	Key            string            `json:"key"`
	Chain          string            `json:"chain"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// --- Certificate CRUD ---

func (c *Client) ListCertificates(ctx context.Context) ([]Certificate, error) {
	var certs []Certificate
	err := c.Get(ctx, "/certificates", &certs)
	return certs, err
}

func (c *Client) GetCertificate(ctx context.Context, id int) (*Certificate, error) {
	var cert Certificate
	err := c.Get(ctx, fmt.Sprintf("/certificates/%d", id), &cert)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func (c *Client) CreateCertificate(ctx context.Context, cert *Certificate) (*Certificate, error) {
	var result Certificate
	err := c.CreateWithApply(ctx, "/certificates", cert, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateCertificate(ctx context.Context, id int, cert *Certificate) (*Certificate, error) {
	var result Certificate
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/certificates/%d", id), cert, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteCertificate(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/certificates/%d", id))
}

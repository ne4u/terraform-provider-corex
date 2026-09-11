package client

import (
	"context"
	"fmt"
)

// SslLabsSettings represents the SSL Labs scan settings for a certificate.
type SslLabsSettings struct {
	MaxScansPerHost int `json:"max_scans_per_host"`
}

// --- SSL Labs Settings (per certificate, PUT, no apply) ---

func (c *Client) GetSslLabsSettings(ctx context.Context, certID int) (*SslLabsSettings, error) {
	var s SslLabsSettings
	err := c.Get(ctx, fmt.Sprintf("/certificates/%d/ssllabs/settings", certID), &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) UpdateSslLabsSettings(ctx context.Context, certID int, s *SslLabsSettings) (*SslLabsSettings, error) {
	var result SslLabsSettings
	err := c.Put(ctx, fmt.Sprintf("/certificates/%d/ssllabs/settings", certID), s, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// loginResponse is the response from POST /auth/token.
type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Role        string `json:"role"`
}

// newFormRequest creates an HTTP request with form-encoded body.
func newFormRequest(ctx context.Context, method, url string, formData url.Values) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

// loginLocked authenticates with username/password and stores the JWT.
// Caller must hold c.mu.
func (c *Client) loginLocked(ctx context.Context) error {
	if c.username == "" || c.password == "" {
		return fmt.Errorf("username and password are required for authentication")
	}

	formData := url.Values{
		"username": {c.username},
		"password": {c.password},
	}
	if c.totpCode != "" {
		formData.Set("totp_code", c.totpCode)
	}

	fullURL := c.BaseURL + "/api/v1/auth/token"
	req, err := newFormRequest(ctx, http.MethodPost, fullURL, formData)
	if err != nil {
		return err
	}

	httpResp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(httpResp.Body)
		var errResp struct {
			Detail string `json:"detail"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Detail != "" {
			return fmt.Errorf("login failed (HTTP %d): %s", httpResp.StatusCode, errResp.Detail)
		}
		return fmt.Errorf("login failed (HTTP %d): %s", httpResp.StatusCode, string(respBody))
	}

	var resp loginResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	c.accessToken = resp.AccessToken
	// JWTs from coreX typically have a 24h expiry. We set a conservative 1h refresh window.
	c.tokenExpiry = time.Now().Add(1 * time.Hour)
	return nil
}

// Suppress unused import warnings (bytes is used in client.go, but we keep it here for future use).
var _ = bytes.NewBuffer

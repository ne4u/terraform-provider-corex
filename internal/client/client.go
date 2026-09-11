package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Client is the HTTP client for the coreX Manager API.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client

	// Auth
	username string
	password string
	totpCode string
	token    string

	// Internal
	mu           sync.Mutex
	accessToken  string
	tokenExpiry  time.Time
	useStaticTok bool
}

// Config holds the provider configuration for the client.
type Config struct {
	Host     string
	Username string
	Password string
	TOTPCode string
	Token    string
	Insecure bool
}

// New creates a new Client from the given Config.
func New(cfg Config) (*Client, error) {
	baseURL := strings.TrimSuffix(cfg.Host, "/")
	if baseURL == "" {
		return nil, fmt.Errorf("host is required")
	}

	transport := &http.Transport{}
	if cfg.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	c := &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Transport: transport, Timeout: 60 * time.Second},
		username:   cfg.Username,
		password:   cfg.Password,
		totpCode:   cfg.TOTPCode,
		token:      cfg.Token,
	}

	if cfg.Token != "" {
		c.accessToken = cfg.Token
		c.useStaticTok = true
	}

	return c, nil
}

// EnsureToken makes sure we have a valid access token, logging in if needed.
func (c *Client) EnsureToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.useStaticTok {
		return nil
	}
	// If we have a token that's not near expiry, keep it.
	if c.accessToken != "" && time.Since(c.tokenExpiry) < -5*time.Minute {
		return nil
	}
	return c.loginLocked(ctx)
}

// do performs an HTTP request and decodes the response into out (if non-nil).
// It automatically adds the Authorization header and handles 401 re-auth.
func (c *Client) do(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	if err := c.EnsureToken(ctx); err != nil {
		return err
	}

	resp, err := c.doRaw(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// On 401, try to re-auth once (only in username/password mode).
	if resp.StatusCode == http.StatusUnauthorized && !c.useStaticTok {
		c.mu.Lock()
		if err := c.loginLocked(ctx); err != nil {
			c.mu.Unlock()
			return fmt.Errorf("re-authentication failed: %w", err)
		}
		c.mu.Unlock()

		resp2, err := c.doRaw(ctx, method, path, body)
		if err != nil {
			return err
		}
		defer resp2.Body.Close()
		return c.decodeResponse(resp2, out)
	}

	return c.decodeResponse(resp, out)
}

// doRaw performs the raw HTTP request without auth retry logic.
func (c *Client) doRaw(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	fullURL := c.BaseURL + "/api/v1" + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	return c.HTTPClient.Do(req)
}

// doForm performs an HTTP request with form-encoded body (used for login).
func (c *Client) doForm(ctx context.Context, method, path string, formData url.Values, out interface{}) error {
	fullURL := c.BaseURL + "/api/v1" + path
	req, err := http.NewRequestWithContext(ctx, method, fullURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return c.decodeResponse(resp, out)
}

// decodeResponse reads the HTTP response and decodes into out, or returns an error.
func (c *Client) decodeResponse(resp *http.Response, out interface{}) error {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Detail string `json:"detail"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Detail != "" {
			return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, errResp.Detail)
		}
		return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, out interface{}) error {
	return c.do(ctx, http.MethodGet, path, nil, out)
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body, out interface{}) error {
	return c.do(ctx, http.MethodPost, path, body, out)
}

// Put performs a PUT request with a JSON body.
func (c *Client) Put(ctx context.Context, path string, body, out interface{}) error {
	return c.do(ctx, http.MethodPut, path, body, out)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

// PostNoApply performs a POST without triggering auto-apply (for MCP resources).
func (c *Client) PostNoApply(ctx context.Context, path string, body, out interface{}) error {
	return c.do(ctx, http.MethodPost, path, body, out)
}

// PutNoApply performs a PUT without triggering auto-apply (for MCP resources).
func (c *Client) PutNoApply(ctx context.Context, path string, body, out interface{}) error {
	return c.do(ctx, http.MethodPut, path, body, out)
}

// DeleteNoApply performs a DELETE without triggering auto-apply (for MCP resources).
func (c *Client) DeleteNoApply(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

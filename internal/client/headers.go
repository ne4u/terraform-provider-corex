package client

import (
	"context"
	"fmt"
	"time"
)

// ResponseHeader represents a coreX response header manipulation rule.
type ResponseHeader struct {
	ID          int       `json:"id"`
	ListenerID  int       `json:"listener_id"`
	ListenerIDs []int     `json:"listener_ids"`
	Header      string    `json:"header"`
	Value       string    `json:"value"`
	Action      string    `json:"action"`
	Condition   string    `json:"condition"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RequestHeader represents a coreX request header manipulation rule.
type RequestHeader struct {
	ID          int       `json:"id"`
	BackendID   int       `json:"backend_id"`
	BackendIDs  []int     `json:"backend_ids"`
	Header      string    `json:"header"`
	Value       string    `json:"value"`
	Action      string    `json:"action"`
	Condition   string    `json:"condition"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// --- ResponseHeader CRUD ---

// ListResponseHeaders returns all response header rules.
func (c *Client) ListResponseHeaders(ctx context.Context) ([]ResponseHeader, error) {
	var headers []ResponseHeader
	err := c.Get(ctx, "/response-headers", &headers)
	return headers, err
}

// GetResponseHeader returns a single response header rule by ID.
func (c *Client) GetResponseHeader(ctx context.Context, id int) (*ResponseHeader, error) {
	var h ResponseHeader
	err := c.Get(ctx, fmt.Sprintf("/response-headers/%d", id), &h)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// CreateResponseHeader creates a new response header rule and applies the HAProxy config.
func (c *Client) CreateResponseHeader(ctx context.Context, h *ResponseHeader) (*ResponseHeader, error) {
	var result ResponseHeader
	err := c.CreateWithApply(ctx, "/response-headers", h, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateResponseHeader updates an existing response header rule and applies the HAProxy config.
func (c *Client) UpdateResponseHeader(ctx context.Context, id int, h *ResponseHeader) (*ResponseHeader, error) {
	var result ResponseHeader
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/response-headers/%d", id), h, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteResponseHeader deletes a response header rule and applies the HAProxy config.
func (c *Client) DeleteResponseHeader(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/response-headers/%d", id))
}

// --- RequestHeader CRUD ---

// ListRequestHeaders returns all request header rules.
func (c *Client) ListRequestHeaders(ctx context.Context) ([]RequestHeader, error) {
	var headers []RequestHeader
	err := c.Get(ctx, "/request-headers", &headers)
	return headers, err
}

// GetRequestHeader returns a single request header rule by ID.
func (c *Client) GetRequestHeader(ctx context.Context, id int) (*RequestHeader, error) {
	var h RequestHeader
	err := c.Get(ctx, fmt.Sprintf("/request-headers/%d", id), &h)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

// CreateRequestHeader creates a new request header rule and applies the HAProxy config.
func (c *Client) CreateRequestHeader(ctx context.Context, h *RequestHeader) (*RequestHeader, error) {
	var result RequestHeader
	err := c.CreateWithApply(ctx, "/request-headers", h, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateRequestHeader updates an existing request header rule and applies the HAProxy config.
func (c *Client) UpdateRequestHeader(ctx context.Context, id int, h *RequestHeader) (*RequestHeader, error) {
	var result RequestHeader
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/request-headers/%d", id), h, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteRequestHeader deletes a request header rule and applies the HAProxy config.
func (c *Client) DeleteRequestHeader(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/request-headers/%d", id))
}

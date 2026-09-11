package client

import (
	"context"
	"fmt"
	"time"
)

// ErrorPage represents a coreX custom error page.
type ErrorPage struct {
	ID          int       `json:"id"`
	ListenerID  int       `json:"listener_id"`
	ListenerIDs []int     `json:"listener_ids"`
	Code        int       `json:"code"`
	ContentType string    `json:"content_type"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// --- ErrorPage CRUD ---

// ListErrorPages returns all error pages.
func (c *Client) ListErrorPages(ctx context.Context) ([]ErrorPage, error) {
	var pages []ErrorPage
	err := c.Get(ctx, "/error-pages", &pages)
	return pages, err
}

// GetErrorPage returns a single error page by ID.
func (c *Client) GetErrorPage(ctx context.Context, id int) (*ErrorPage, error) {
	var p ErrorPage
	err := c.Get(ctx, fmt.Sprintf("/error-pages/%d", id), &p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateErrorPage creates a new error page and applies the HAProxy config.
func (c *Client) CreateErrorPage(ctx context.Context, p *ErrorPage) (*ErrorPage, error) {
	var result ErrorPage
	err := c.CreateWithApply(ctx, "/error-pages", p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateErrorPage updates an existing error page and applies the HAProxy config.
func (c *Client) UpdateErrorPage(ctx context.Context, id int, p *ErrorPage) (*ErrorPage, error) {
	var result ErrorPage
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/error-pages/%d", id), p, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteErrorPage deletes an error page and applies the HAProxy config.
func (c *Client) DeleteErrorPage(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/error-pages/%d", id))
}

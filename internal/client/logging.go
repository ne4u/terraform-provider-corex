package client

import (
	"context"
	"fmt"
	"time"
)

// LogDestination represents a log destination attached to a listener.
type LogDestination struct {
	ID         int       `json:"id"`
	ListenerID int       `json:"listener_id"`
	Name       string    `json:"name"`
	Target     string    `json:"target"`
	Facility   string    `json:"facility"`
	Level      string    `json:"level"`
	Format     string    `json:"format"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// LoggedField represents a logged field attached to a listener.
type LoggedField struct {
	ID         int       `json:"id"`
	ListenerID int       `json:"listener_id"`
	Name       string    `json:"name"`
	Field      string    `json:"field"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// --- LogDestination CRUD ---

// ListLogDestinations lists all log destinations.
func (c *Client) ListLogDestinations(ctx context.Context) ([]LogDestination, error) {
	var dests []LogDestination
	err := c.Get(ctx, "/log-destinations", &dests)
	return dests, err
}

// GetLogDestination retrieves a log destination by ID.
func (c *Client) GetLogDestination(ctx context.Context, id int) (*LogDestination, error) {
	var d LogDestination
	err := c.Get(ctx, fmt.Sprintf("/log-destinations/%d", id), &d)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// CreateLogDestination creates a log destination.
func (c *Client) CreateLogDestination(ctx context.Context, d *LogDestination) (*LogDestination, error) {
	var result LogDestination
	err := c.CreateWithApply(ctx, "/log-destinations", d, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateLogDestination updates a log destination by ID.
func (c *Client) UpdateLogDestination(ctx context.Context, id int, d *LogDestination) (*LogDestination, error) {
	var result LogDestination
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/log-destinations/%d", id), d, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteLogDestination deletes a log destination by ID.
func (c *Client) DeleteLogDestination(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/log-destinations/%d", id))
}

// --- LoggedField CRUD ---

// ListLoggedFields lists all logged fields.
func (c *Client) ListLoggedFields(ctx context.Context) ([]LoggedField, error) {
	var fields []LoggedField
	err := c.Get(ctx, "/logged-fields", &fields)
	return fields, err
}

// GetLoggedField retrieves a logged field by ID.
func (c *Client) GetLoggedField(ctx context.Context, id int) (*LoggedField, error) {
	var f LoggedField
	err := c.Get(ctx, fmt.Sprintf("/logged-fields/%d", id), &f)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// CreateLoggedField creates a logged field.
func (c *Client) CreateLoggedField(ctx context.Context, f *LoggedField) (*LoggedField, error) {
	var result LoggedField
	err := c.CreateWithApply(ctx, "/logged-fields", f, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateLoggedField updates a logged field by ID.
func (c *Client) UpdateLoggedField(ctx context.Context, id int, f *LoggedField) (*LoggedField, error) {
	var result LoggedField
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/logged-fields/%d", id), f, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteLoggedField deletes a logged field by ID.
func (c *Client) DeleteLoggedField(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/logged-fields/%d", id))
}

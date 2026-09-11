package client

import (
	"context"
	"fmt"
	"time"
)

// User represents a coreX Manager user.
type User struct {
	ID              int       `json:"id"`
	Username        string    `json:"username"`
	Role            string    `json:"role"`
	Email           string    `json:"email"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Organization    string    `json:"organization"`
	IsActive        bool      `json:"is_active"`
	PasswordExpired bool      `json:"password_expired"`
	// Password is write-only: sent on create/update, never returned by the API.
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListUsers lists all users.
func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	var users []User
	err := c.Get(ctx, "/users", &users)
	return users, err
}

// GetUser retrieves a user by ID.
func (c *Client) GetUser(ctx context.Context, id int) (*User, error) {
	var u User
	err := c.Get(ctx, fmt.Sprintf("/users/%d", id), &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// CreateUser creates a new user.
func (c *Client) CreateUser(ctx context.Context, u *User) (*User, error) {
	var result User
	err := c.CreateWithApply(ctx, "/users", u, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateUser updates a user by ID.
func (c *Client) UpdateUser(ctx context.Context, id int, u *User) (*User, error) {
	var result User
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/users/%d", id), u, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteUser deletes a user by ID.
func (c *Client) DeleteUser(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/users/%d", id))
}

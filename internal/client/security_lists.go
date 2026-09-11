package client

import (
	"context"
	"fmt"
	"time"
)

// SecurityList represents a generic security list (network, asn, geo, ja4, pattern).
type SecurityList struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// SecurityListEntry represents an entry within a security list.
type SecurityListEntry struct {
	ID        int       `json:"id"`
	ListID    int       `json:"list_id"`
	Value     string    `json:"value"`
	Note      *string   `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DynamicFeed represents a dynamic feed that populates a security list.
type DynamicFeed struct {
	ID                  int        `json:"id"`
	Name                string     `json:"name"`
	ListType            string     `json:"list_type"`
	URL                 string     `json:"url"`
	UpdateIntervalHours int        `json:"update_interval_hours"`
	TargetListID        *int       `json:"target_list_id"`
	Enabled             bool       `json:"enabled"`
	AutoApply           bool       `json:"auto_apply"`
	Description         *string    `json:"description"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// --- SecurityList CRUD ---

func (c *Client) ListSecurityLists(ctx context.Context, listType string) ([]SecurityList, error) {
	var lists []SecurityList
	err := c.Get(ctx, fmt.Sprintf("/security-lists/%s", listType), &lists)
	return lists, err
}

func (c *Client) GetSecurityList(ctx context.Context, listType string, id int) (*SecurityList, error) {
	var list SecurityList
	err := c.Get(ctx, fmt.Sprintf("/security-lists/%s/%d", listType, id), &list)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) CreateSecurityList(ctx context.Context, listType string, body *SecurityList) (*SecurityList, error) {
	var result SecurityList
	err := c.CreateWithApply(ctx, fmt.Sprintf("/security-lists/%s", listType), body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateSecurityList(ctx context.Context, listType string, id int, body *SecurityList) (*SecurityList, error) {
	var result SecurityList
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/security-lists/%s/%d", listType, id), body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteSecurityList(ctx context.Context, listType string, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/security-lists/%s/%d", listType, id))
}

// --- SecurityListEntry CRUD ---

func (c *Client) ListSecurityListEntries(ctx context.Context, listType string, listID int) ([]SecurityListEntry, error) {
	var entries []SecurityListEntry
	err := c.Get(ctx, fmt.Sprintf("/security-lists/%s/%d/entries", listType, listID), &entries)
	return entries, err
}

func (c *Client) CreateSecurityListEntry(ctx context.Context, listType string, listID int, body *SecurityListEntry) (*SecurityListEntry, error) {
	var result SecurityListEntry
	err := c.CreateWithApply(ctx, fmt.Sprintf("/security-lists/%s/%d/entries", listType, listID), body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateSecurityListEntry(ctx context.Context, listType string, id int, body *SecurityListEntry) (*SecurityListEntry, error) {
	var result SecurityListEntry
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/security-lists/%s/entries/%d", listType, id), body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteSecurityListEntry(ctx context.Context, listType string, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/security-lists/%s/entries/%d", listType, id))
}

// --- DynamicFeed CRUD ---

func (c *Client) ListDynamicFeeds(ctx context.Context) ([]DynamicFeed, error) {
	var feeds []DynamicFeed
	err := c.Get(ctx, "/security-lists/feeds", &feeds)
	return feeds, err
}

func (c *Client) GetDynamicFeed(ctx context.Context, id int) (*DynamicFeed, error) {
	var feed DynamicFeed
	err := c.Get(ctx, fmt.Sprintf("/security-lists/feeds/%d", id), &feed)
	if err != nil {
		return nil, err
	}
	return &feed, nil
}

func (c *Client) CreateDynamicFeed(ctx context.Context, body *DynamicFeed) (*DynamicFeed, error) {
	var result DynamicFeed
	err := c.CreateWithApply(ctx, "/security-lists/feeds", body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) UpdateDynamicFeed(ctx context.Context, id int, body *DynamicFeed) (*DynamicFeed, error) {
	var result DynamicFeed
	err := c.UpdateWithApply(ctx, fmt.Sprintf("/security-lists/feeds/%d", id), body, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) DeleteDynamicFeed(ctx context.Context, id int) error {
	return c.DeleteWithApply(ctx, fmt.Sprintf("/security-lists/feeds/%d", id))
}
